package poteto

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/poteto-go/poteto/constant"
)

func BenchmarkContext_JSON(b *testing.B) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/example.com", strings.NewReader(userJSON))
	ctx := NewContext(w, req).(*context)

	testUser := user{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx.JSON(http.StatusOK, testUser)
	}
}

func BenchmarkContext_RequestId(b *testing.B) {
	tests := []struct {
		name  string
		setup func(ctx Context)
	}{
		{
			name: "Get from store",
			setup: func(ctx Context) {
				ctx.Set(constant.StoredRequestId, "uuid")
			},
		},
		{
			name: "Get from header",
			setup: func(ctx Context) {
				req := ctx.GetRequest()
				req.Header.Set(constant.HeaderRequestId, "uuid")
			},
		},
		{
			name: "Generate new uuid",
			setup: func(ctx Context) {
			},
		},
	}

	for _, it := range tests {
		b.Run(it.name, func(b *testing.B) {
			// Arrange
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			ctx := NewContext(w, req).(*context)
			// call setup function
			it.setup(ctx)

			b.ReportAllocs()
			b.ResetTimer()

			// Act
			for i := 0; i < b.N; i++ {
				ctx.RequestId()
			}
		})
	}
}
