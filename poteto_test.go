package poteto

import (
	stdContext "context"
	"errors"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/goccy/go-json"
	"github.com/poteto-go/poteto/constant"
)

func TestAddRouteToPoteto(t *testing.T) {
	poteto := New()

	tests := []struct {
		name   string
		method string
		path   string
		want   bool
	}{
		{"success add new route", http.MethodGet, "/users/find", false},
		{"fail add already existed route", http.MethodGet, "/users/find", true},
		{"success add new method already existed route", http.MethodPost, "/users/find", false},
		{"success add new method already existed route", http.MethodPut, "/users/find", false},
		{"success add new method already existed route", http.MethodPatch, "/users/find", false},
		{"success add new method already existed route", http.MethodDelete, "/users/find", false},
		{"success add new method already existed route", http.MethodHead, "/users/find", false},
		{"success add new method already existed route", http.MethodOptions, "/users/find", false},
		{"success add new method already existed route", http.MethodTrace, "/users/find", false},
		{"success add new method already existed route", http.MethodConnect, "/users/find", false},
	}

	for _, it := range tests {
		t.Run(it.name, func(tt *testing.T) {
			err := func() error {
				switch it.method {
				case http.MethodGet:
					return poteto.GET(it.path, nil)
				case http.MethodPost:
					return poteto.POST(it.path, nil)
				case http.MethodPut:
					return poteto.PUT(it.path, nil)
				case http.MethodPatch:
					return poteto.PATCH(it.path, nil)
				case http.MethodDelete:
					return poteto.DELETE(it.path, nil)
				case http.MethodHead:
					return poteto.HEAD(it.path, nil)
				case http.MethodOptions:
					return poteto.OPTIONS(it.path, nil)
				case http.MethodTrace:
					return poteto.TRACE(it.path, nil)
				case http.MethodConnect:
					return poteto.CONNECT(it.path, nil)
				default:
					return nil
				}
			}()

			if it.want {
				if err == nil {
					t.Errorf("FATAL: success already existed route")
				}
			} else {
				if err != nil {
					t.Errorf("FATAL: fail new route")
				}
			}
		})
	}
}

func TestRunAndStop(t *testing.T) {
	p := New()

	tests := []struct {
		name  string
		port1 string
		port2 string
	}{
		{"Test 127.0.0.1:8080", "127.0.0.1:8080", ""},
		{"Test 8080", "8080", ""},
		{"Test collision panic", "127.0.0.1:8080", "127.0.0.1:8080"},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			errChan := make(chan error)
			go func() {
				errChan <- p.Run(it.port1)
			}()

			errChan2 := make(chan error)
			if it.port2 != "" {
				go func() {
					errChan2 <- p.Run(it.port2)
				}()
			}

			select {
			case <-time.After(500 * time.Millisecond):
				if err := p.Stop(stdContext.Background()); err != nil {
					t.Errorf("Unmatched")
				}
			case <-errChan:
				return
			case <-errChan2:
				return
			}
		})
	}
}

func TestRunTLS(t *testing.T) {
	cert, _ := os.ReadFile("./_fixture/certs/cert.pem")
	key, _ := os.ReadFile("./_fixture/certs/key.pem")

	p := New()

	errChan := make(chan error)
	go func() {
		errChan <- p.RunTLS("8080", cert, key)
	}()

	select {
	case <-time.After(500 * time.Millisecond):
		if err := p.Stop(stdContext.Background()); err != nil {
			t.Errorf("Unmatched")
		}
	case <-errChan:
		return
	}
}

func TestRunTLSStartUpWorkflows(t *testing.T) {
	cert, _ := os.ReadFile("./_fixture/certs/cert.pem")
	key, _ := os.ReadFile("./_fixture/certs/key.pem")

	p := New()

	isCalled := false
	calledFunc := func() error {
		isCalled = true
		return nil
	}

	p.RegisterWorkflow(constant.StartUpWorkflow, 1, calledFunc)

	errChan := make(chan error)
	go func() {
		errChan <- p.RunTLS("8080", cert, key)
	}()

	select {
	case <-time.After(500 * time.Millisecond):
		if !isCalled {
			t.Errorf("Unmatched")
		}
		if err := p.Stop(stdContext.Background()); err != nil {
			t.Errorf("Unmatched")
		}
	case <-errChan:
		return
	}
}

func TestCertParseError(t *testing.T) {
	cert := []byte("hello")
	key, _ := os.ReadFile("./_fixture/certs/key.pem")

	p := New()

	errChan := make(chan error)
	go func() {
		errChan <- p.RunTLS("8080", cert, key)
	}()

	select {
	case <-time.After(500 * time.Millisecond):
		t.Errorf("not occur error")
		if err := p.Stop(stdContext.Background()); err != nil {
			t.Errorf("Unmatched")
		}
	case <-errChan:
		// Pass case
		return
	}
}

func TestSetLogger(t *testing.T) {
	p := New().(*poteto)
	logger := func(msg string) {
		return
	}

	if p.logger != nil {
		t.Error("Unmatched")
	}

	p.SetLogger(logger)
	if p.logger == nil {
		t.Errorf("Unmatched")
	}
}

func TestRunHandlerErrorInSetupServer(t *testing.T) {
	defer monkey.UnpatchAll()

	p := New()
	monkey.Patch((*poteto).setupServer, func(p *poteto) error {
		return errors.New("error")
	})

	if err := p.Run("90"); err == nil {
		t.Errorf("Unmatched")
	}
}

func TestRunStartUpWorkflowsError(t *testing.T) {
	p := New()
	calledFunc := func() error {
		return errors.New("error")
	}

	p.RegisterWorkflow(constant.StartUpWorkflow, 1, calledFunc)

	errChan := make(chan error)
	go func() {
		errChan <- p.Run("3032")
	}()

	select {
	case <-time.After(500 * time.Millisecond):
		t.Errorf("Unmatched")
		p.Stop(stdContext.Background())
	case <-errChan:
		// Pass case
		p.Stop(stdContext.Background())
	}
}

func TestSetupServerHandleListenError(t *testing.T) {
	defer monkey.UnpatchAll()

	p := New()
	monkey.Patch(net.Listen, func(pro, add string) (net.Listener, error) {
		return nil, errors.New("error")
	})

	if err := p.setupServer(); err == nil {
		t.Errorf("Unmatched")
	}
}

func TestStopHandleError(t *testing.T) {
	defer monkey.UnpatchAll()

	p := New()
	monkey.Patch((*http.Server).Shutdown, func(srv *http.Server, ctx stdContext.Context) error {
		return errors.New("error")
	})

	if err := p.Stop(stdContext.Background()); err == nil {
		t.Errorf("Unmatched")
	}
}

func TestCheck(t *testing.T) {
	p := New()

	p.GET("/users", getAllUserForTest)

	tests := []struct {
		name     string
		method   string
		path     string
		expected bool
	}{
		{"hit handler", http.MethodGet, "/users", true},
		{"different path", http.MethodGet, "/unexpected", false},
		{"different method", http.MethodPost, "/users", false},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			result := p.Check(it.method, it.path)
			if result != it.expected {
				t.Errorf("unmatched: actual(%v) - expected(%v)", result, it.expected)
			}
		})
	}
}

func TestChain(t *testing.T) {
	p := New()

	p.GET(
		"/users",
		p.Chain(
			sampleMiddleware,
			sampleMiddleware2,
		)(getAllUserForTest),
	)

	res := p.Play(http.MethodGet, "/users")
	hv1 := res.Header().Get("Hello")
	hv2 := res.Header().Get("Hello2")
	resBodyStr := res.Body.String()

	if resBodyStr[:len(resBodyStr)-1] != `{"name":"user"}` {
		t.Error("unmatched")
	}

	if hv1 != "world" {
		t.Error("unmatched")
	}

	if hv2 != "world2" {
		t.Error("unmatched")
	}
}

func TestServeHTTP(t *testing.T) {
	p := New()

	p.GET("/users", func(ctx Context) error {
		if qp, ok := ctx.QueryParam("query"); ok {
			return ctx.JSON(http.StatusOK, map[string]string{
				"result": "query",
				"value":  qp,
			})
		}

		return ctx.JSON(http.StatusOK, map[string]string{
			"result": "get-handle",
			"value":  "none",
		})
	})

	p.POST("/users", func(ctx Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"result": "post-handle",
			"value":  "none",
		})
	})

	p.PUT("/users", func(ctx Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"result": "put-handle",
			"value":  "none",
		})
	})

	p.PATCH("/users", func(ctx Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"result": "patch-handle",
			"value":  "none",
		})
	})

	p.CONNECT("/users", func(ctx Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"result": "connect-handle",
			"value":  "none",
		})
	})

	p.DELETE("/users", func(ctx Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"result": "delete-handle",
			"value":  "none",
		})
	})

	p.HEAD("/users", func(ctx Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"result": "head-handle",
			"value":  "none",
		})
	})

	p.OPTIONS("/users", func(ctx Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"result": "options-handle",
			"value":  "none",
		})
	})

	p.TRACE("/users", func(ctx Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"result": "trace-handle",
			"value":  "none",
		})
	})

	p.GET("/users/:id", func(ctx Context) error {
		if pp, ok := ctx.PathParam("id"); ok {
			return ctx.JSON(http.StatusOK, map[string]string{
				"result": "path",
				"value":  pp,
			})
		}

		if qp, ok := ctx.QueryParam("query"); ok {
			return ctx.JSON(http.StatusOK, map[string]string{
				"result": "query",
				"value":  qp,
			})
		}

		return ctx.JSON(http.StatusOK, map[string]string{
			"result": "handle",
			"value":  "none",
		})
	})

	tests := []struct {
		name   string
		method string
		url    string
		code   uint
		result string
		value  string
	}{
		{
			"can handle get method",
			http.MethodGet,
			"/users",
			http.StatusOK,
			"get-handle",
			"none",
		},
		{
			"can handle post method",
			http.MethodPost,
			"/users",
			http.StatusOK,
			"post-handle",
			"none",
		},
		{
			"can handle put method",
			http.MethodPut,
			"/users",
			http.StatusOK,
			"put-handle",
			"none",
		},
		{
			"can handle patch method",
			http.MethodPatch,
			"/users",
			http.StatusOK,
			"patch-handle",
			"none",
		},
		{
			"can handle delete method",
			http.MethodDelete,
			"/users",
			http.StatusOK,
			"delete-handle",
			"none",
		},
		{
			"can handle head method",
			http.MethodHead,
			"/users",
			http.StatusOK,
			"head-handle",
			"none",
		},
		{
			"can handle options method",
			http.MethodOptions,
			"/users",
			http.StatusOK,
			"options-handle",
			"none",
		},
		{
			"can handle trace method",
			http.MethodTrace,
			"/users",
			http.StatusOK,
			"trace-handle",
			"none",
		},
		{
			"can get path param",
			http.MethodGet,
			"/users/1",
			http.StatusOK,
			"path",
			"1",
		},
		{
			"can get query param",
			http.MethodGet,
			"/users?query=1",
			http.StatusOK,
			"query",
			"1",
		},
		{
			"can get path param with query",
			http.MethodGet,
			"/users/1?query=1",
			http.StatusOK,
			"path",
			"1",
		},
		{
			"can handle not found",
			http.MethodGet,
			"/unexpected",
			http.StatusNotFound,
			"query",
			"1",
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			result := map[string]string{}
			response := p.Play(it.method, it.url)

			if response.Code != int(it.code) {
				t.Errorf("unmatched code %d != %d", response.Code, it.code)
			}

			if response.Code == http.StatusNotFound {
				return
			}

			json.Unmarshal(response.Body.Bytes(), &result)

			if result["result"] != it.result {
				t.Errorf("unmatched result %s != %s", result["result"], it.result)
			}

			if result["value"] != it.value {
				t.Errorf("unmatched value %s != %s", result["value"], it.value)
			}
		})
	}
}
