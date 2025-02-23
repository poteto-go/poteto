package poteto

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/google/uuid"
	"github.com/poteto-go/poteto/constant"
)

func TestJSON(t *testing.T) {
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
			context.JSON(it.code, it.val)
			if body := w.Body.String(); body[0:27] != it.expected[0:27] {
				t.Errorf("FATAL: context json")
			}

			if header := w.Header(); header[constant.HeaderContentType][0] != constant.ApplicationJson {
				t.Errorf("FATAL: wrong content-type")
			}
		})
	}
}

func TestQueryParam(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		key1      string
		expected1 string
		key2      string
		expected2 string
	}{
		{
			"Test valid case",
			"/example.com?hello=world",
			"hello",
			"world",
			"unknown",
			"",
		},
		{
			"Test nothing param case",
			"/example.com?hello",
			"hello",
			"",
			"unknown",
			"",
		},
		{
			"too many param case",
			"/example.com?a=a&b=b&c=c&d=d&e=e&f=f#g=g&h=h&i=i&j=j&k=k&l=l&m=m&n=n&o=o&p=p&q=q&r=r&s=s&t=t&u=u&v=v&w=w&x=x&y=y&z=z&a1=a1&b1=b1&c1=c1&d1=d1&e1=e1&f1=f1&g1=g1&h1=h1",
			"a",
			"",
			"unknown",
			"",
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", it.url, nil)
			ctx := NewContext(w, req).(*context)

			queryParams := req.URL.Query()
			ctx.SetQueryParam(queryParams)

			queryParam1, _ := ctx.QueryParam(it.key1)
			if queryParam1 != it.expected1 {
				t.Errorf("Cannot Get Query Param")
			}

			queryParam2, _ := ctx.QueryParam(it.key2)
			if queryParam2 != it.expected2 {
				t.Errorf("Cannot Get nil If Unknown key")
			}
		})
	}
}

func TestPathParam(t *testing.T) {
	url := "/example.com/users/1"

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", url, nil)
	ctx := NewContext(w, req).(*context)

	ctx.SetParam(constant.ParamTypePath, ParamUnit{key: ":id", value: "1"})

	tests := []struct {
		name        string
		key         string
		expected    string
		expected_ok bool
	}{
		{"Can get PathParam", "id", "1", true},
		{"If unexpected key", "unexpected", "", false},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			param, ok := ctx.PathParam(it.key)

			if param != it.expected {
				t.Errorf("Unmatched")
			}

			if ok != it.expected_ok {
				t.Errorf("unmatched")
			}
		})

	}
}

func TestSetAndPath(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	ctx := NewContext(w, req).(*context)

	expected := "http://expected.com"
	ctx.SetPath(expected)

	path := ctx.GetPath()
	if path != expected {
		t.Errorf("Not Matched")
	}
}

func BenchmarkJSON(b *testing.B) {
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

func TestRemoteIP(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/example.com", strings.NewReader(userJSON))
	ctx := NewContext(w, req).(*context)

	if _, err := ctx.GetRemoteIP(); err != nil {
		t.Errorf("Error occurred")
	}
}

func TestGetIPFromXFFHeaderByContext(t *testing.T) {
	iph := &ipHandler{}
	_, ipnet, _ := net.ParseCIDR("10.0.0.0/24")
	iph.RegisterTrustIPRange(ipnet)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(constant.HeaderXForwardedFor, "11.0.0.1, 12.0.0.1, 10.0.0.2, 10.0.0.1")
	ctx := NewContext(w, req).(*context)

	ipString, _ := ctx.GetIPFromXFFHeader()
	if ipString != "12.0.0.1" {
		t.Errorf("Not matched")
	}
}

func TestRealIP(t *testing.T) {
	tests := []struct {
		name        string
		headerKey   string
		headerValue string
		expected    string
	}{
		{
			"Get from Real Ip",
			constant.HeaderXRealIp,
			"11.0.0.1",
			"11.0.0.1",
		},
		{
			"Get from XFF",
			constant.HeaderXForwardedFor,
			"11.0.0.1",
			"11.0.0.1",
		},
		{
			"Get from RemoteAddr",
			"",
			"",
			"192.0.2.1",
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			if it.headerKey != "" {
				req.Header.Set(it.headerKey, it.headerValue)
			}
			ctx := NewContext(w, req).(*context)

			ipString, _ := ctx.RealIP()
			if ipString[0:8] != it.expected[0:8] {
				t.Errorf(ipString)
				t.Errorf("Not matched")
			}
		})
	}
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
			req.Header.Set(constant.HeaderContentType, constant.ApplicationJson)
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
