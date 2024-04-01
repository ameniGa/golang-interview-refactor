package helpers_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "interview/pkg/helpers"
	"os"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	f, err := os.CreateTemp(".", "form*.html")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	f.WriteString(`
<div>
	{{ if .CartItems }}
	{{range .CartItems}}
	<div>Product: {{.Product}}</div>
	<div>Quantity: {{.Quantity}}</div>
	{{end}}
	{{end }}
</div>
`)

	t.Run("should render template", func(t *testing.T) {
		str, err := RenderTemplate(map[string]interface{}{"CartItems": []map[string]interface{}{
			{
				"Quantity": "2",
				"Product":  "bag",
			},
			{
				"Quantity": "3",
				"Product":  "shoes",
			}}}, f.Name())
		assert.NoError(t, err)
		assert.Contains(t, str, "<div>Product: bag</div>")
		assert.Contains(t, str, "<div>Quantity: 2</div>")
		assert.Contains(t, str, "<div>Product: shoes</div>")
		assert.Contains(t, str, "<div>Quantity: 3</div>")
	})

	t.Run("invalid path", func(t *testing.T) {
		str, err := RenderTemplate(map[string]interface{}{"CartItems": []map[string]interface{}{
			{
				"Quantity": "2",
				"Product":  "bag",
			},
		}}, "some/path")
		assert.Error(t, err)
		assert.Equal(t, "", str)
	})

	t.Run("invalid data", func(t *testing.T) {
		_, err := RenderTemplate("hello world", f.Name())
		assert.Error(t, err)
	})
}
