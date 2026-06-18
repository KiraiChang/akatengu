package services_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// enums.InitEnums() 由 test.NewTestDBGinkgo() 內部的 sync.Once 統一負責，
// 此處不重複呼叫，避免雙重 register panic。
func TestJournalEntrySuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Journal Entry Integration Suite")
}
