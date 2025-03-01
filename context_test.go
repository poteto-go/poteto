package poteto

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/agiledragon/gomonkey"
	"github.com/google/uuid"
	"github.com/poteto-go/poteto/constant"
	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	// Act
	ctx := NewContext(nil, nil)

	// Assert
	assert.NotNil(t, ctx)
}

func TestContext_JSON(t *testing.T) {
	// Arrange
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	context := NewContext(w, req).(*context)

	tests := []struct {
		name     string
		code     int
		val      testVal
		expected string
	}{
		{
			"status ok & can serialize",
			http.StatusOK,
			testVal{Name: "test", Val: "val"},
			`{"name":"test","val":"val"}`,
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			// Act
			context.JSON(it.code, it.val)
			resBody := w.Body.String()
			header := w.Header()
			status := w.Code

			// Assert
			assert.Equal(t, it.expected[:len(resBody)-1], resBody[:len(resBody)-1])
			assert.Equal(t, constant.ApplicationJson, header.Get(constant.HeaderContentType))
			assert.Equal(t, it.code, status)
		})
	}
}

func TestContext_QueryParam(t *testing.T) {
	// Arrange
	ctx := NewContext(nil, nil).(*context)

	// Mock
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(
		reflect.TypeOf(ctx.httpParams),
		"GetParam",
		func(_ *httpParam, paramType, key string) (string, bool) {
			assert.Equal(t, constant.ParamTypeQuery, paramType)
			assert.Equal(t, "test", key)
			return "test", true
		},
	)

	// Act
	result, ok := ctx.QueryParam("test")

	// Assert
	assert.Equal(t, "test", result)
	assert.Equal(t, true, ok)
}

func TestContext_PathParam(t *testing.T) {
	// Arrange
	ctx := NewContext(nil, nil).(*context)

	// Mock
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(
		reflect.TypeOf(ctx.httpParams),
		"GetParam",
		func(_ *httpParam, paramType, key string) (string, bool) {
			assert.Equal(t, constant.ParamTypePath, paramType)
			assert.Equal(t, ":id", key)
			return "mocked_path_value", true
		},
	)

	// Act
	result, ok := ctx.PathParam("id")

	// Assert
	assert.Equal(t, "mocked_path_value", result)
	assert.True(t, ok)
}

func TestContext_SetPath(t *testing.T) {
	// Arrange
	ctx := NewContext(nil, nil).(*context)
	testPath := "/tests"

	// Act
	ctx.SetPath(testPath)

	// Assert
	assert.Equal(t, testPath, ctx.path)
}

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

func TestContext_RemoteIP(t *testing.T) {
	// Arrange
	ctx := NewContext(nil, nil).(*context)

	// Mock
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(
		reflect.TypeOf(ctx.ipHandler),
		"GetRemoteIP",
		func(_ *ipHandler, c Context) (string, error) {
			assert.Equal(t, c, ctx)
			return "127.0.0.1", nil
		},
	)

	// Act
	result, err := ctx.GetRemoteIP()

	// Assert
	assert.Equal(t, "127.0.0.1", result)
	assert.Nil(t, err)
}

func TestContext_GetIPFromXFFHeader(t *testing.T) {
	// Arrange
	ctx := NewContext(nil, nil).(*context)

	// Mock
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(
		reflect.TypeOf(ctx.ipHandler),
		"GetIPFromXFFHeader",
		func(_ *ipHandler, c Context) (string, error) {
			assert.Equal(t, c, ctx)
			return "12.0.0.1", nil
		},
	)

	// Act
	result, err := ctx.GetIPFromXFFHeader()

	// Assert
	assert.Equal(t, "12.0.0.1", result)
	assert.Nil(t, err)
}

func TestContext_RealIP(t *testing.T) {
	// Arrange
	ctx := NewContext(nil, nil).(*context)

	// Mock
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(
		reflect.TypeOf(ctx.ipHandler),
		"RealIP",
		func(_ *ipHandler, c Context) (string, error) {
			assert.Equal(t, c, ctx)
			return "12.0.0.1", nil
		},
	)

	// Act
	result, err := ctx.RealIP()

	// Assert
	assert.Equal(t, "12.0.0.1", result)
	assert.Nil(t, err)
}

func TestGetLogger(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/example.com", strings.NewReader(userJSON))
	ctx := NewContext(w, req).(*context)

	logger := func(msg string) {
		return
	}
	ctx.SetLogger(logger)

	if ctx.Logger() == nil {
		t.Errorf("Unmatched")
	}
}

func TestBindOnContext(t *testing.T) {
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
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			user := User{}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/example.com", bytes.NewBufferString(string(it.body)))
			req.Header.Set(constant.HeaderContentType, constant.ApplicationJson)
			ctx := NewContext(w, req).(*context)

			err := ctx.Bind(&user)
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

func TestBindWithValidateOnContext(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Mail string `json:"mail" validate:"required,email"`
	}

	tests := []struct {
		name     string
		body     []byte
		worked   bool
		expected User
	}{
		{
			"test ok validate",
			[]byte(`{"name":"test", "mail":"test@example.com"}`),
			true, User{Name: "test", Mail: "test@example.com"},
		},
		{
			"test fatal validate",
			[]byte(`{"name":"test", "mail":"example"}`),
			false, User{},
		},
		{
			"test fatal bind",
			[]byte(`{"name":"test",, "mail":"example"}`),
			false, User{},
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			user := User{}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/example.com", bytes.NewBufferString(string(it.body)))
			req.Header.Set(constant.HEADER_CONTENT_TYPE, constant.APPLICATION_JSON)
			ctx := NewContext(w, req).(*context)

			err := ctx.BindWithValidate(&user)
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

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req).(*context)

	ctx.NoContent()

	if w.Result().Status != "204 No Content" {
		t.Errorf("Unmatched")
	}
}

func TestSetAndGet(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req).(*context)

	tests := []struct {
		key   string
		value string
	}{
		{"key", "value"},
		{"key", "value"},
		{"key", "value"},
		{"key", "value"},
		{"key", "value"},
		{"key", "value"},
		{"key", "value"},
		{"key", "value"},
		{"key", "value"},
		{"key", "value"},
		{"key", "value"},
		{"key", "value"},
	}

	var wg sync.WaitGroup
	for _, test := range tests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx.Set(test.key, test.value)

			val, ok := ctx.Get(test.key)
			if !ok || val != test.value {
				t.Errorf("Unmatched")
			}
		}()
	}

	wg.Wait()
}

func TestRequestId(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		stored   string
		expected string
	}{
		{
			"Test from ReqHeader",
			"uuid",
			"",
			"uuid",
		},
		{
			"Test from stored",
			"",
			"uuid",
			"uuid",
		},
		{
			"Test random case",
			"",
			"",
			"uuid",
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)

			if it.header != "" {
				req.Header.Set(constant.HeaderRequestId, it.header)
			}

			ctx := NewContext(w, req).(*context)

			if it.stored != "" {
				ctx.Set(constant.StoredRequestId, it.stored)
			}

			requestId := ctx.RequestId()
			if requestId != it.expected {
				if it.header != "" || it.stored != "" {
					t.Errorf("Unmatched")
				}
			}

			// random case
			if it.header == "" && it.stored == "" {
				if requestId == it.expected {
					t.Errorf("Unmatched")
				}
			}
		})
	}
}

func TestRequestIdErrorGenInUuid(t *testing.T) {
	defer monkey.UnpatchAll()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req).(*context)

	monkey.Patch(uuid.NewRandom, func() (uuid.UUID, error) {
		return uuid.UUID{}, errors.New("error")
	})

	val := ctx.RequestId()
	if val != "" {
		t.Errorf("Unmatched")
	}
}

func TestDebugParam(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test/", nil)
	ctx := NewContext(w, req)

	ctx.SetParam(constant.ParamTypePath, ParamUnit{"user_id", "1"})
	ctx.SetParam(constant.ParamTypePath, ParamUnit{"player_id", "2"})
	ctx.SetParam(constant.ParamTypeQuery, ParamUnit{"user_id", "1"})

	expected := `{"path":{"player_id":"2","user_id":"1"},"query":{"user_id":"1"}}`

	debugParam, _ := ctx.DebugParam()
	if debugParam != expected {
		t.Errorf(
			"Unmatched actual(%s) -> expected(%s)",
			debugParam,
			expected,
		)
	}
}

func TestJsonDeserialize(t *testing.T) {
	defer monkey.UnpatchAll()
	tests := []struct {
		name string
		err  any
	}{
		{"UnmarshalTypeError", &json.UnmarshalTypeError{}},
		{"SyntaxError", &json.SyntaxError{}},
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test/", nil)
	ctx := NewContext(w, req)

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			monkey.Patch((*json.Decoder).Decode, func(d *json.Decoder, v any) error {
				return it.err.(error)
			})

			if err := ctx.JsonDeserialize(&user{}); err == nil {
				t.Errorf("Not occur error: %v", err)
			}
		})
	}
}

func TestRegisterTrustIPRangeInContext(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test/", nil)
	ctx := NewContext(w, req)
	_, ipnet, _ := net.ParseCIDR("10.0.0.0/24")
	ctx.RegisterTrustIPRange(ipnet)
}

func TestJSONRPCError(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req).(*context)

	ctx.JSONRPCError(200, "message", "data", 10)
}
