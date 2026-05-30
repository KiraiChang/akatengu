package factory

import (
	"akatengu/internal/enums"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAssetPipelineSuite(t *testing.T) {
	enums.InitEnums()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Asset Pipeline Factory Suite")
}
