package poteto

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/poteto0/poteto/constant"
)

func TestBind(t *testing.T) {
	binder := NewBinder()

	type User struct {
		Name string `json:"name"`
		Mail string `json:"mail"`
	}

	tests := []struct {
		name     string
		body     []byte
		worked   bool
		expected User
	}{
		{
			"Test Normal Case",
			[]byte(`{"name":"test", "mail":"example"}`),
			true,
			User{Name: "test", Mail: "example"},
		},
		{
			"Test Error Case",
			[]byte(`{"name":"test",, "mail":"example"}`),
			false,
			User{Name: "test", Mail: "example"},
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			user := User{}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "https://example.com", bytes.NewBufferString(string(it.body)))
			req.Header.Set(constant.HEADER_CONTENT_TYPE, constant.APPLICATION_JSON)
			ctx := NewContext(w, req).(*context)

			err := binder.Bind(ctx, &user)
			if err != nil {
				if it.worked {
					t.Errorf("unexpected error")
				}
				return
			}

			if !it.worked {
				t.Errorf("unexpected not error")
				return
			}

			if it.expected.Name != user.Name {
				t.Errorf("Unmatched")
			}

			if it.expected.Mail != user.Mail {
				t.Errorf("Unmatched")
			}
		})
	}
}

func BenchmarkBind(b *testing.B) {
	binder := NewBinder()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "https://example.com", bytes.NewBufferString(string(userJSON)))
	req.Header.Set(constant.HEADER_CONTENT_TYPE, constant.APPLICATION_JSON)
	ctx := NewContext(w, req).(*context)

	testUser := user{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		binder.Bind(ctx, &testUser)
	}
}
