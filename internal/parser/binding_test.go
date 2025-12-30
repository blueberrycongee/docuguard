package parser

import (
	"testing"

	"github.com/blueberrycongee/docuguard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractBindings(t *testing.T) {
	bindings, err := ExtractBindings("../../testdata/docs/sample.md")
	require.NoError(t, err)

	assert.Len(t, bindings, 2)

	// 验证第一个绑定
	assert.Equal(t, "testdata/code/payment.go", bindings[0].CodeFile)
	assert.Equal(t, "CalculateShipping", bindings[0].CodeSymbol)
	assert.Equal(t, types.BindingFunc, bindings[0].CodeType)
	assert.Contains(t, bindings[0].DocContent, "500 元")

	// 验证第二个绑定
	assert.Equal(t, "CalculateDiscount", bindings[1].CodeSymbol)
	assert.Contains(t, bindings[1].DocContent, "VIP")
}
