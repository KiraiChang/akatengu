package projection

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"encoding/json"
	"time"
)

// EventExportRecord 是匯出檔案中單一事件的格式。
// 刻意排除 event_id（AUTOINCREMENT，重入後會重新分配）與 merchant_id（支援跨商戶遷移）。
// 使用強型別 enum 欄位，JSON marshal/unmarshal 可自動完成字串轉換。
type EventExportRecord struct {
	EventUuid        string                `json:"event_uuid"`
	OccurredAt       time.Time             `json:"occurred_at"`
	AggregateType    enums.AggregateType   `json:"aggregate_type"`
	AggregateId      string                `json:"aggregate_id"`
	AggregateVersion int64                 `json:"aggregate_version"`
	EventType        event_types.EventType `json:"event_type"`
	Payload          json.RawMessage       `json:"payload"`
	Metadata         json.RawMessage       `json:"metadata,omitempty"`
	UpdatedBy        *string               `json:"updated_by,omitempty"`
}

// EventExportPayload 是加密時被加密的內層資料。
type EventExportPayload struct {
	ExportedAt time.Time           `json:"exported_at"`
	Events     []EventExportRecord `json:"events"`
}

// EventExportFile 是匯出檔案的最外層格式，含版本號與加密狀態。
// version 欄位控管格式版本；目前唯一支援的版本為 "1"。
type EventExportFile struct {
	Version   string `json:"version"`
	Encrypted bool   `json:"encrypted"`

	// 明碼時填入：
	ExportedAt *time.Time          `json:"exported_at,omitempty"`
	Events     []EventExportRecord `json:"events,omitempty"`

	// 加密時填入（AES-256-GCM + PBKDF2-SHA256）：
	KdfSalt     string `json:"kdf_salt,omitempty"`
	CipherNonce string `json:"cipher_nonce,omitempty"`
	Data        string `json:"data,omitempty"`
}
