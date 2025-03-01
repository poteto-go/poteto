package poteto

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poteto-go/poteto/constant"
	"github.com/poteto-go/poteto/perror"
	"github.com/stretchr/testify/assert"
)

func TestNewBind(t *testing.T) {
	binder := NewBinder()
	assert.NotNil(t, binder)
}

func TestBinder_Bind(t *testing.T) {
	binder := NewBinder()

	type User struct {
		Name string `json:"name"`
		Mail string `json:"mail"`
	}

	t.Run("Success", func(t *testing.T) {
		// Arrange
		expected := User{Name: "test", Mail: "example"}
		req := httptest.NewRequest(
			http.MethodGet,
			"https://example.com",
			bytes.NewBufferString(`{"name":"test", "mail":"example"}`),
		)
		req.Header.Set(constant.HeaderContentType, constant.ApplicationJson)
		ctx := NewContext(httptest.NewRecorder(), req).(*context)

		// Act
		user := User{}
		err := binder.Bind(ctx, &user)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, user)
	})

	t.Run("MarshallError", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(
			http.MethodGet,
			"https://example.com",
			bytes.NewBufferString(`{"name":"test",, "mail":"example"}`),
		)
		req.Header.Set(constant.HeaderContentType, constant.ApplicationJson)
		ctx := NewContext(httptest.NewRecorder(), req).(*context)

		// Act
		user := User{}
		err := binder.Bind(ctx, &user)

		// Assert
		assert.Error(t, err)
	})

	t.Run("ZeroLengthError", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(
			http.MethodGet,
			"https://example.com",
			bytes.NewBufferString(``),
		)
		req.Header.Set(constant.HeaderContentType, constant.ApplicationJson)
		ctx := NewContext(httptest.NewRecorder(), req).(*context)

		// Act
		user := User{}
		err := binder.Bind(ctx, &user)

		// Assert
		assert.ErrorIs(t, err, perror.ErrZeroLengthContent)
	})

	t.Run("NotApplicationJsonError", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(
			http.MethodGet,
			"https://example.com",
			bytes.NewBufferString(`{"name":"test", "mail":"example"}`),
		)
		ctx := NewContext(httptest.NewRecorder(), req).(*context)

		// Act
		user := User{}
		err := binder.Bind(ctx, &user)

		// Assert
		assert.ErrorIs(t, err, perror.ErrNotApplicationJson)
	})
}

func TestBinder_BindWithValidate(t *testing.T) {
	binder := NewBinder()

	type User struct {
		Name string `json:"name"`
		Mail string `json:"mail" validate:"required,email"`
	}

	tests := []struct {
		name        string
		body        []byte
		expectError bool
		expected    User
	}{
		{
			"test ok validate",
			[]byte(`{"name":"test", "mail":"test@example.com"}`),
			false, User{Name: "test", Mail: "test@example.com"},
		},
		{
			"test fatal validate",
			[]byte(`{"name":"test", "mail":"example"}`),
			true, User{},
		},
		{
			"test fatal bind",
			[]byte(`{"name":"test",, "mail":"example"}`),
			true, User{},
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "https://example.com", bytes.NewBufferString(string(it.body)))
			req.Header.Set(constant.HEADER_CONTENT_TYPE, constant.APPLICATION_JSON)
			ctx := NewContext(w, req).(*context)

			user := User{}
			err := binder.BindWithValidate(ctx, &user)
			if it.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, it.expected, user)
			}
		})
	}
}

func BenchmarkBind_Bind(b *testing.B) {
	type User struct {
		Name string `json:"name"`
		Mail string `json:"mail"`
	}

	binder := NewBinder()

	// Arrange
	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodGet,
		"https://example.com",
		bytes.NewBufferString(`{"name":"test", "mail":"example"}`),
	)
	req.Header.Set(constant.HeaderContentType, constant.ApplicationJson)
	ctx := NewContext(w, req).(*context)

	testUser := User{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		binder.Bind(ctx, &testUser)
	}
}
