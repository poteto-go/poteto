package poteto

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
