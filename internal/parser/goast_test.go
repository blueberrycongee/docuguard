package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/blueberrycongee/docuguard/pkg/types"
)

func TestGoParser_ExtractSymbol(t *testing.T) {
	p := NewGoParser()

	code, line, err := p.ExtractSymbol(
		"../../testdata/code/payment.go",
		"CalculateShipping",
		types.BindingFunc,
	)

	require.NoError(t, err)
	assert.NotEmpty(t, code)
	assert.Greater(t, line, 0)
	assert.Contains(t, code, "func CalculateShipping")
	assert.Contains(t, code, "1000") // 代码中的实际值
}

func TestGoParser_ExtractSymbol_NotFound(t *testing.T) {
	p := NewGoParser()

	code, line, err := p.ExtractSymbol(
		"../../testdata/code/payment.go",
		"NonExistentFunction",
		types.BindingFunc,
	)

	require.NoError(t, err)
	assert.Empty(t, code)
	assert.Equal(t, 0, line)
}
