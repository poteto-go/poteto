package poteto

import (
	stdContext "context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/poteto-go/poteto/constant"
	"github.com/ybbus/jsonrpc/v3"
)

type (
	TestCalculator struct{}
	AdditionArgs   struct {
		Add, Added int
	}
)

func (tc *TestCalculator) Add(r *http.Request, args *AdditionArgs) int {
	return args.Add + args.Added
}

func (tc *TestCalculator) AddVoid(r *http.Request, args *AdditionArgs) {}

func TestPotetoJSONRPCAdapterCall(t *testing.T) {
	p := New()

	rpc := TestCalculator{}
	p.POST("/add", func(ctx Context) error {
		return PotetoJsonRPCAdapter[TestCalculator, AdditionArgs](ctx, &rpc)
	})

	errChan := make(chan error)
	go func() {
		errChan <- p.Run(":6000")
	}()

	// client
	added := 10
	add := 10
	rpcClient := jsonrpc.NewClient("http://localhost:6000/add")
	result, err := rpcClient.Call(stdContext.Background(), "TestCalculator.Add", &AdditionArgs{Added: added, Add: add})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	num, _ := result.Result.(json.Number).Int64()
	if int(num) != 20 {
		t.Errorf("Unmatched actual(%v) -> expected(%v)", result.Result, 20)
	}

	select {
	case <-time.After(500 * time.Millisecond):
		if err := p.Stop(stdContext.Background()); err != nil {
			t.Errorf("Unmatched")
		}
	case <-errChan:
		t.Errorf("Unexpected error occur")
	}
}

func TestPotetoJSONRPCAdapterCallReturnVoid(t *testing.T) {
	p := New()

	rpc := TestCalculator{}
	p.POST("/add", func(ctx Context) error {
		return PotetoJsonRPCAdapter[TestCalculator, AdditionArgs](ctx, &rpc)
	})

	errChan := make(chan error)
	go func() {
		errChan <- p.Run(":6001")
	}()

	// client
	added := 10
	add := 10
	rpcClient := jsonrpc.NewClient("http://localhost:6001/add")
	result, err := rpcClient.Call(stdContext.Background(), "TestCalculator.AddVoid", &AdditionArgs{Added: added, Add: add})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if result.Result != nil {
		t.Errorf("Unmatched expected nil")
	}

	select {
	case <-time.After(500 * time.Millisecond):
		if err := p.Stop(stdContext.Background()); err != nil {
			t.Errorf("Unmatched")
		}
	case <-errChan:
		t.Errorf("Unexpected error occur")
	}
}

func TestPotetoJSONRPCAdapter(t *testing.T) {
	tests := []struct {
		name     string
		req      *http.Request
		expected float64
	}{
		{
			"Test GET req",
			httptest.NewRequest("GET", "/test", nil),
			-32700,
		},
		{
			"Test POST body nil",
			httptest.NewRequest("POST", "/test", nil),
			-32600,
		},
		{
			"Test POST body not version right",
			httptest.NewRequest(
				http.MethodPost,
				"/test",
				strings.NewReader(rpcJSONId),
			),
			-32600,
		},
		{
			"Test not POST body not method right",
			httptest.NewRequest(
				http.MethodPost,
				"/test",
				strings.NewReader(rpcJSONVersion),
			),
			-32600,
		},
		{
			"Test POST body not found method",
			httptest.NewRequest(
				http.MethodPost,
				"/test",
				strings.NewReader(rpcJSONMethod),
			),
			-32601,
		},
		{
			"Test POST body not found method but class is matched",
			httptest.NewRequest(
				http.MethodPost,
				"/test",
				strings.NewReader(rpcJSONMethodWrong),
			),
			-32601,
		},
		{
			"Test POST body not found method cause class is wrong",
			httptest.NewRequest(
				http.MethodPost,
				"/test",
				strings.NewReader(rpcJSONMethodClass),
			),
			-32601,
		},
		{
			"Test params not equal length",
			httptest.NewRequest(
				http.MethodPost,
				"/test",
				strings.NewReader(rpcJSONParams),
			),
			-32600,
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			it.req.Header.Set(constant.HEADER_CONTENT_TYPE, constant.APPLICATION_JSON)
			ctx := NewContext(w, it.req).(*context)

			rpc := TestCalculator{}
			PotetoJsonRPCAdapter[TestCalculator, int](ctx, &rpc)

			var data map[string]any
			json.Unmarshal(w.Body.Bytes(), &data)
			err := data["error"].(map[string]any)
			if err["code"] != it.expected {
				t.Errorf(
					"Unmatched actual(%v) -> expected(%v)",
					err["code"],
					it.expected,
				)
			}
		})
	}
}
