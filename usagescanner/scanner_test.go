package usagescanner

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScanForUsages(t *testing.T) {
	usages, err := ScanForUsages(".")
	assert.NoError(t, err)
	assert.Equal(t, []string{}, usages.Resolve())
}
