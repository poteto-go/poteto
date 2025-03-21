package poteto

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkServeHTTP(b *testing.B) {
	p := New()
	p.GET("/users/:id", getAllUserForTestById)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users/1", strings.NewReader(userJSON))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTPWORequestId(b *testing.B) {
	option := PotetoOption{
		WithRequestId: false,
	}
	p := NewWithOption(option)
	p.GET("/users/:id", getAllUserForTestById)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users/1", strings.NewReader(userJSON))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p.ServeHTTP(w, req)
	}
}

func TestServeHTTPWithMiddleware(t *testing.T) {
	p := New()
	p.Register(sampleMiddleware)
	p.GET("/users", getAllUserForTest)

	group := p.Combine("/tests")
	group.Register(sampleMiddleware2)
	p.GET("/tests", getAllUserForTest)
	p.GET("/tests/:id", getAllUserForTestById)

	tests := []struct {
		name          string
		reqMethod     string
		reqUrl        string
		worked        bool
		expectedKey   string
		expectedValue string
	}{
		{
			"Test Middleware",
			"GET",
			"/users",
			true,
			"Hello",
			"world",
		},
		{
			"Test MiddlewareTree",
			"GET",
			"/tests",
			true,
			"Hello2",
			"world2",
		},
		{
			"Test if MiddlewareTree includes all middleware",
			"GET",
			"/tests",
			true,
			"Hello",
			"world",
		},
		{
			"Test parent Middleware",
			"GET",
			"/tests/1",
			true,
			"Hello2",
			"world2",
		},
		{
			"Test not apply middleware without group",
			"GET",
			"/users",
			false,
			"Hello2",
			"world2",
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(it.reqMethod, it.reqUrl, nil)

			p.ServeHTTP(w, req)
			header := w.Header()
			if it.worked {
				if header[it.expectedKey][0] != it.expectedValue {
					t.Errorf("Unmatched")
				}
			} else {
				if header[it.expectedKey] != nil {
					t.Errorf("Unmatched")
				}
			}
		})
	}
}

func TestLeafHandler(t *testing.T) {
	p := New()

	p.Leaf("/users", func(leaf Leaf) {
		leaf.Register(sampleMiddleware)
		leaf.GET("/", getAllUserForTest)
		leaf.POST("/create", getAllUserForTest)
		leaf.PUT("/change", getAllUserForTest)
		leaf.DELETE("/delete", getAllUserForTest)
	})

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
			"GET",
			"/users",
			"Hello",
			"world",
			`{"name":"user"}`,
		},
		{
			"Test Post",
			"POST",
			"/users/create",
			"Hello",
			"world",
			`{"name":"user"}`,
		},
		{
			"Test Put",
			"PUT",
			"/users/change",
			"Hello",
			"world",
			`{"name":"user"}`,
		},
		{
			"Test Delete",
			"DELETE",
			"/users/delete",
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

func TestParamRouter(t *testing.T) {
	p := New().(*poteto)
	p.GET("/users/:id", getAllUserForTest)
	p.GET("/users/:id/name", getAllUserForTest)
	p.GET("/users/:id/name/:greet", getAllUserForTest)

	tests := []struct {
		name         string
		reqUrl       string
		expectedKeys []string
		expectedVals []string
	}{
		{
			"Test last param case",
			"/users/1",
			[]string{"id"},
			[]string{"1"},
		},
		{
			"Test mid param case",
			"/users/1/name",
			[]string{"id"},
			[]string{"1"},
		},
		{
			"Test two param case",
			"/users/1/name/hello",
			[]string{"id", "greet"},
			[]string{"1", "hello"},
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, it.reqUrl, nil)

			p.ServeHTTP(w, req)

			if w.Body.String()[0:10] != `{"name":"user"}`[0:10] {
				t.Errorf(
					"Unmatched actual(%s) -> expected(%s)",
					w.Body.String()[0:10],
					`{"name":"user"}`[0:10],
				)
			}

			if ctx, ok := p.cache.Get().(*context); ok {
				for i, key := range it.expectedKeys {
					val, ok := ctx.PathParam(key)
					if !ok {
						t.Errorf("Not has key %s", key)
						return
					}
					if val != it.expectedVals[i] {
						t.Errorf(
							"Unmatched actual(%s:%s) -> expected(%s:%s)",
							key,
							val,
							key,
							it.expectedVals[i],
						)
					}
				}
			}
		})
	}
}
