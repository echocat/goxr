package usagescanner

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestScanForUsages(t *testing.T) {
	g := NewGomegaWithT(t)

	usages, err := ScanForUsages(".")
	g.Expect(err).To(BeNil())
	g.Expect(usages).ToNot(BeNil())
	g.Expect(usages.Resolve()).To(Equal([]string{}))
}
