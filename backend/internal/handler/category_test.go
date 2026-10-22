package handler

import (
	"akatengu/internal/bootstrap"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

// ─────────────────────────────────────────
// Spec runner — NewMux category write routes
// ─────────────────────────────────────────

var _ = Describe("NewMux category write routes", func() {
	for _, s := range categoryAPIScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			db := newSettingTestDB(0)
			insertCFAuthFixture(db)
			mux := NewMux(db, bootstrap.Config{JWT: cfTestJWTSecret}, zap.NewNop())
			serve := func(method, path, body string) *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				mux.ServeHTTP(rec, settingMuxRequest(settingRequest{method, path, body}))
				Expect(rec.Code).To(Equal(http.StatusOK), "%s %s: %s", method, path, rec.Body.String())
				return rec
			}

			// The response is the creation event's event_store record.
			var created struct {
				EventUuid string
			}
			Expect(json.Unmarshal(serve(http.MethodPost, s.path, s.createBody).Body.Bytes(), &created)).To(Succeed())
			categoryUUID := created.EventUuid
			Expect(categoryUUID).NotTo(BeEmpty())

			serve(http.MethodPut, s.path+"/"+categoryUUID, s.updateBody)
			serve(http.MethodDelete, s.path+"/"+categoryUUID, categoryDeleteBody)

			var events []struct {
				AggregateType string `db:"aggregate_type"`
				AggregateID   string `db:"aggregate_id"`
				Version       int64  `db:"aggregate_version"`
				EventType     string `db:"event_type"`
				Source        string `db:"source"`
			}
			Expect(db.Select(&events, `SELECT aggregate_type, aggregate_id, aggregate_version, event_type,
				COALESCE(json_extract(metadata, '$.execution.source'), '') AS source FROM event_store ORDER BY event_id`)).To(Succeed())
			Expect(events).To(HaveLen(3))
			for i, suffix := range []string{"created", "updated", "deleted"} {
				Expect(events[i].AggregateType).To(Equal(s.aggregateType))
				Expect(events[i].AggregateID).To(Equal(categoryUUID))
				Expect(events[i].Version).To(Equal(int64(i + 1)))
				Expect(events[i].EventType).To(Equal(s.eventPrefix + "." + suffix))
				Expect(events[i].Source).To(Equal("api"), "written through BFS")
			}

			var row struct {
				Name     string `db:"name"`
				IsActive bool   `db:"is_active"`
				Version  int64  `db:"version"`
			}
			Expect(db.Get(&row, `SELECT name, is_active, version FROM `+s.table+` WHERE category_uuid = ? AND merchant_id = ?`,
				categoryUUID, cfTestMerchantID)).To(Succeed())
			Expect(row.Name).To(Equal(s.updatedName))
			Expect(row.IsActive).To(BeFalse())
			Expect(row.Version).To(Equal(int64(3)))
		})
	}
})
