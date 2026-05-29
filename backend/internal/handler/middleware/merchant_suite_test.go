package middleware

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMerchantSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Merchant Middleware Suite")
}
