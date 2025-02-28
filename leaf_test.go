package poteto

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLeaf(t *testing.T) {
	p := New()

	leaf := NewLeaf(p, "/users")
	leaf.Register(sampleMiddleware)
	leaf.GET("/", getAllUserForTest)
	leaf.POST("/create", getAllUserForTest)
	leaf.PUT("/change", getAllUserForTest)
	leaf.PATCH("/patch", getAllUserForTest)
	leaf.DELETE("/delete", getAllUserForTest)
	leaf.HEAD("/head", getAllUserForTest)
	leaf.OPTIONS("/options", getAllUserForTest)
	leaf.TRACE("/trace", getAllUserForTest)
	leaf.CONNECT("/connect", getAllUserForTest)

	tests := []struct {
		name          string
		reqMethod     string
		reqUrl        string
		expectedKey   string
		expectedValue string
		expectedRes   string
	}{
		{
			"Test Get",
			http.MethodGet,
			"/users",
			"Hello",
			"world",
			`{"name":"user"}`,
		},
		{
			"Test Post",
			http.MethodPost,
			"/users/create",
			"Hello",
			"world",
			`{"name":"user"}`,
		},
		{
			"Test Put",
			http.MethodPut,
			"/users/change",
			"Hello",
			"world",
			`{"name":"user"}`,
		},
		{
			"Test Patch",
			http.MethodPatch,
			"/users/patch",
			"Hello",
			"world",
			`{"name":"user"}`,
		},
		{
			"Test Delete",
			http.MethodDelete,
			"/users/delete",
			"Hello",
			"world",
			`{"name":"user"}`,
		},
		{
			"Test Head",
			http.MethodHead,
			"/users/head",
			"Hello",
			"world",
			`{"name":"user"}`,
		},
		{
			"Test Options",
			http.MethodOptions,
			"/users/options",
			"Hello",
			"world",
			`{"name":"user"}`,
		},
		{
			"Test Trace",
			http.MethodTrace,
			"/users/trace",
			"Hello",
			"world",
			`{"name":"user"}`,
		},
		{
			"Test Connect",
			http.MethodConnect,
			"/users/connect",
			"Hello",
			"world",
			`{"name":"user"}`,
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(it.reqMethod, it.reqUrl, nil)

			p.ServeHTTP(w, req)
			if w.Header()[it.expectedKey][0] != it.expectedValue {
				t.Errorf("Unmatched")
			}
			if w.Body.String()[0:10] != it.expectedRes[0:10] {
				t.Errorf("Unmatched")
			}
		})
	}
}
