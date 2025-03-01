package poteto

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync"
	"testing"

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

func TestContext_JSONRPCError(t *testing.T) {
	// Arrange
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req).(*context)

	// Mock
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(
		reflect.TypeOf(ctx),
		"JSON",
		func(_ *context, code int, value any) error {
			// Assert
			assert.Equal(t, code, http.StatusOK)
			assert.Equal(t, value, map[string]any{
				"result":  nil,
				"jsonrpc": "2.0",
				"error": map[string]any{
					"code":    code,
					"message": "message",
					"data":    "data",
				},
				"id": 10,
			})
			return nil
		},
	)

	// Act
	ctx.JSONRPCError(http.StatusOK, "message", "data", 10)
}

func TestContext_GetPath(t *testing.T) {
	// Arrange
	ctx := NewContext(nil, nil).(*context)
	testPath := "/tests"
	ctx.path = testPath

	// Act
	result := ctx.GetPath()

	// Assert
	assert.Equal(t, result, testPath)
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

func TestContext_SetQueryParam(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   url.Values
		expected      map[string]string
		expectedCount int
		maxParamCount int
	}{
		{
			name: "Normal case",
			queryParams: url.Values{
				"key1": {"value1"},
				"key2": {"value2", "value3"},
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2, value3",
			},
			expectedCount: 2,
			maxParamCount: constant.MaxQueryParamCount,
		},
		{
			name:          "Empty query params",
			queryParams:   url.Values{},
			expected:      map[string]string{},
			expectedCount: 0,
			maxParamCount: constant.MaxQueryParamCount,
		},
		{
			name: "Too many query params",
			queryParams: func() url.Values {
				values := url.Values{}
				for i := 0; i < constant.MaxQueryParamCount+1; i++ {
					values[string(rune('a'+i))] = []string{"value"}
				}
				return values
			}(),
			expected:      map[string]string{},
			expectedCount: 0,
			maxParamCount: constant.MaxQueryParamCount,
		},
		{
			name: "empty value",
			queryParams: url.Values{
				"key1": {},
			},
			expected:      map[string]string{},
			expectedCount: 0,
			maxParamCount: constant.MaxQueryParamCount,
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			// Arrange
			ctx := NewContext(nil, nil).(*context)

			// Act
			ctx.SetQueryParam(it.queryParams)

			// Assert
			assert.Equal(t, len(ctx.httpParams.(*httpParam).queryParams), it.expectedCount)

			for key, value := range it.expected {
				actualValue, ok := ctx.httpParams.GetParam(constant.ParamTypeQuery, key)
				assert.True(t, ok)
				assert.Equal(t, value, actualValue)
			}
		})
	}
}

func TestContext_SetParam(t *testing.T) {
	// Arrange
	ctx := NewContext(nil, nil).(*context)
	paramType := "test"
	paramUnit := ParamUnit{
		key:   "key",
		value: "value",
	}

	// Mock
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(
		reflect.TypeOf(ctx.httpParams),
		"AddParam",
		func(_ *httpParam, paramType string, param ParamUnit) {
			// Assert
			assert.Equal(t, paramType, paramType)
			assert.Equal(t, paramUnit, param)
		},
	)

	// Act
	ctx.SetParam(paramType, paramUnit)
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

func TestContext_Bind(t *testing.T) {
	// Arrange
	ctx := NewContext(nil, nil).(*context)
	type User struct {
		Name string `json:"name"`
	}
	user := User{}

	// Mock
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(
		reflect.TypeOf(ctx.binder),
		"Bind",
		func(_ *binder, c Context, obj any) error {
			assert.Equal(t, c, ctx)
			assert.Equal(t, &user, obj)
			return nil
		},
	)

	// Act
	err := ctx.Bind(&user)

	// Assert
	assert.Nil(t, err)
}

func TestContext_BindWithValidate(t *testing.T) {
	// Arrange
	ctx := NewContext(nil, nil).(*context)
	type User struct {
		Name string `json:"name"`
	}
	user := User{}

	// Mock
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(
		reflect.TypeOf(ctx.binder),
		"BindWithValidate",
		func(_ *binder, c Context, obj any) error {
			assert.Equal(t, c, ctx)
			assert.Equal(t, &user, obj)
			return nil
		},
	)

	// Act
	err := ctx.BindWithValidate(&user)

	// Assert
	assert.Nil(t, err)
}

func TestContext_DebugParam(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(ctx Context)
		expectedResult  string
		expectedSuccess bool
		mockError       error
	}{
		{
			name: "Normal case with path and query parameters",
			setup: func(ctx Context) {
				ctx.SetParam(constant.ParamTypePath, ParamUnit{"user_id", "1"})
				ctx.SetParam(constant.ParamTypePath, ParamUnit{"player_id", "2"})
				ctx.SetParam(constant.ParamTypeQuery, ParamUnit{"user_id", "3"})
			},
			expectedResult:  `{"path":{"player_id":"2","user_id":"1"},"query":{"user_id":"3"}}`,
			expectedSuccess: true,
			mockError:       nil,
		},
		{
			name: "Empty parameters",
			setup: func(ctx Context) {
			},
			expectedResult:  `{"path":{},"query":{}}`,
			expectedSuccess: true,
			mockError:       nil,
		},
		{
			name: "only path parameter",
			setup: func(ctx Context) {
				ctx.SetParam(constant.ParamTypePath, ParamUnit{"player_id", "2"})
			},
			expectedResult:  `{"path":{"player_id":"2"},"query":{}}`,
			expectedSuccess: true,
			mockError:       nil,
		},
		{
			name: "only query parameter",
			setup: func(ctx Context) {
				ctx.SetParam(constant.ParamTypeQuery, ParamUnit{"user_id", "3"})
			},
			expectedResult:  `{"path":{},"query":{"user_id":"3"}}`,
			expectedSuccess: true,
			mockError:       nil,
		},
		{
			name: "occur serialize error",
			setup: func(ctx Context) {
				ctx.SetParam(constant.ParamTypeQuery, ParamUnit{"user_id", "3"})
			},
			expectedResult:  "",
			expectedSuccess: false,
			mockError:       errors.New("json serialize error"),
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			// Arrange
			ctx := NewContext(nil, nil).(*context)
			// call setup function
			it.setup(ctx)

			// Mock
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyMethod(
				reflect.TypeOf(ctx.httpParams),
				"JsonSerialize",
				func(_ *httpParam) ([]byte, error) {
					if it.mockError != nil {
						return []byte{}, it.mockError
					}

					return []byte(it.expectedResult), nil
				},
			)

			// Act
			result, ok := ctx.DebugParam()

			// Assert
			assert.Equal(t, it.expectedResult, result)
			assert.Equal(t, it.expectedSuccess, ok)
		})
	}
}

func TestContext_WriteHeader(t *testing.T) {
	// Arrange
	w := httptest.NewRecorder()
	ctx := NewContext(w, nil).(*context)

	// Mock
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(
		reflect.TypeOf(ctx.response),
		"WriteHeader",
		func(_ *response, code int) {
			// Assert
			assert.Equal(t, code, http.StatusOK)
		},
	)

	// Act
	ctx.WriteHeader(http.StatusOK)
}

func TestContext_GetResponse(t *testing.T) {
	// Arrange
	ctx := NewContext(httptest.NewRecorder(), nil).(*context)

	// Act
	result := ctx.GetResponse()

	// Assert
	assert.Equal(t, result, ctx.response)
}

func TestContext_SetResponseHeader(t *testing.T) {
	// Arrange
	ctx := NewContext(httptest.NewRecorder(), nil).(*context)

	// Mock
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(
		reflect.TypeOf(ctx.response),
		"SetHeader",
		func(_ *response, key, value string) {
			// Assert
			assert.Equal(t, key, constant.HeaderContentType)
			assert.Equal(t, value, constant.ApplicationJson)
		},
	)

	// Act
	ctx.SetResponseHeader(constant.HeaderContentType, constant.ApplicationJson)
}

func TestContext_GetRequest(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(nil, req).(*context)

	// Act
	result := ctx.GetRequest()

	// Assert
	assert.Equal(t, req, result)
}

func TestContext_GetRequestHeaderParam(t *testing.T) {
	tests := []struct {
		name        string
		headerKey   string
		headerValue string
		expected    string
	}{
		{
			"Normal case",
			"X-Request-Id",
			"test-id",
			"test-id",
		},
		{
			"Empty value",
			"X-Empty",
			"",
			"",
		},
		{
			"Not existing value",
			"X-NotExist",
			"",
			"",
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			// Arrange
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set(it.headerKey, it.headerValue)
			ctx := NewContext(nil, req).(*context)

			// Act
			result := ctx.GetRequestHeaderParam(it.headerKey)

			// Assert
			assert.Equal(t, it.expected, result)
		})
	}
}

func TestContext_ExtractRequestHeaderParam(t *testing.T) {
	tests := []struct {
		name        string
		headerKey   string
		headerValue []string
		expected    []string
	}{
		{
			"Normal case",
			"X-Test",
			[]string{"value1", "value2"},
			[]string{"value1", "value2"},
		},
		{
			"Empty value",
			"X-Empty",
			[]string{},
			[]string{},
		},
		{
			"Not existing value",
			"X-NotExist",
			nil,
			nil,
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			// Arrange
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header[it.headerKey] = it.headerValue
			ctx := NewContext(nil, req).(*context)

			// Act
			result := ctx.ExtractRequestHeaderParam(it.headerKey)

			// Assert
			assert.Equal(t, it.expected, result)
		})
	}
}

func TestContext_JsonSerialize(t *testing.T) {
	tests := []struct {
		name          string
		input         any
		expected      string
		expectedError bool
		mockError     error
	}{
		{
			name:          "Normal case",
			input:         map[string]any{"key": "value"},
			expected:      `{"key":"value"}`,
			expectedError: false,
			mockError:     nil,
		},
		{
			name:          "Empty map",
			input:         map[string]any{},
			expected:      `{}`,
			expectedError: false,
			mockError:     nil,
		},
		{
			name:          "error on json encode",
			input:         make(chan int), // channel is not json encodable
			expected:      "",
			expectedError: true,
			mockError:     nil,
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			// Arrange
			w := httptest.NewRecorder()
			ctx := NewContext(w, nil).(*context)

			// Mock
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			// Act
			err := ctx.JsonSerialize(it.input)
			resBody := w.Body.String()

			// Assert
			if it.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// check
			if it.expected != "" {
				assert.Equal(t, resBody[:len(resBody)-1], it.expected[:len(resBody)-1])
			}
		})
	}
}

func TestContext_JsonDeserialize(t *testing.T) {
	// Define a struct for test
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name         string
		body         []byte
		contentType  string
		expectedUser User
	}{
		{
			name:         "Valid JSON",
			body:         []byte(`{"name": "John Doe", "age": 30}`),
			contentType:  constant.ApplicationJson,
			expectedUser: User{Name: "John Doe", Age: 30},
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			// Arrange
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(it.body))
			req.Header.Set(constant.HeaderContentType, it.contentType)
			ctx := NewContext(w, req).(*context)

			// Mock
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			// Mock io.ReadAll
			patches.ApplyFunc(
				json.NewDecoder,
				func(_ io.Reader) *json.Decoder {
					return &json.Decoder{}
				},
			)

			// Mock (*json.Decoder).Decode
			patches.ApplyMethod(
				reflect.TypeOf(&json.Decoder{}),
				"Decode",
				func(_ *json.Decoder, v any) error {
					assert.Equal(t, &it.expectedUser, v)
					return nil
				},
			)

			// Act
			user := &User{}
			err := ctx.JsonDeserialize(user)

			// Assert
			assert.Nil(t, err)
		})
	}
}

func TestContext_NoContent(t *testing.T) {
	// Arrange
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req).(*context)

	// Act
	ctx.NoContent()

	// Assert
	assert.Equal(t, w.Code, http.StatusNoContent)
}

func TestContext_SetAndGet(t *testing.T) {
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
			assert.Equal(t, ok, true)
			assert.Equal(t, test.value, val)
		}()
	}

	wg.Wait()
}

func TestContext_RequestId(t *testing.T) {
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
			if it.header != "" || it.stored != "" {
				assert.Equal(t, requestId, it.expected)
			}

			// random case
			if it.header == "" && it.stored == "" {
				assert.NotEqual(t, requestId, it.expected)
			}
		})
	}

	t.Run("Error on uuid()", func(t *testing.T) {
		// Arrange
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := NewContext(w, req).(*context)

		// Mock
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyFunc(
			uuid.NewRandom,
			func() (uuid.UUID, error) {
				return uuid.UUID{}, errors.New("error")
			},
		)

		// Act
		requestId := ctx.RequestId()

		// Assert
		assert.Equal(t, "", requestId)
	})
}
func TestContext_GetRemoteIP(t *testing.T) {
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

func TestContest_RegisterTrustIPRange(t *testing.T) {
	// Arrange
	ctx := NewContext(nil, nil).(*context)
	_, ipnet, _ := net.ParseCIDR("10.0.0.0/24")

	// Mock
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(
		reflect.TypeOf(ctx.ipHandler),
		"RegisterTrustIPRange",
		func(_ *ipHandler, ranges *net.IPNet) {
			// Assert
			assert.Equal(t, ipnet, ranges)
		},
	)

	// Act
	ctx.RegisterTrustIPRange(ipnet)
}

func TestContext_Reset(t *testing.T) {
	// Arrange
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req).(*context)

	// Set values to ensure they are reset
	ctx.Set("test", "value")
	ctx.SetPath("/old")
	ctx.SetQueryParam(url.Values{"old": {"value"}})

	newW := httptest.NewRecorder()
	newReq := httptest.NewRequest("POST", "/new", nil)

	// Act
	ctx.Reset(newW, newReq)

	// Assert
	// check request
	assert.Equal(t, newReq, ctx.request)
	// check response
	assert.NotEqual(t, w, ctx.response)
	// check path
	assert.Equal(t, "", ctx.GetPath())
	// check query params
	_, ok := ctx.QueryParam("old")
	assert.False(t, ok)
	// check store
	_, ok = ctx.Get("test")
	assert.False(t, ok)
}

func TestContext_Logger(t *testing.T) {
	// Arrange
	ctx := NewContext(nil, nil).(*context)
	logger := func(msg string) {}

	// Act
	ctx.SetLogger(logger)
	result := ctx.Logger()

	// Assert
	assert.NotNil(t, result)
}
