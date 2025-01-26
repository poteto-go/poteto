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
)

var debugOption = PotetoOption{
	WithRequestId:   false,
	DebugMode:       true,
	ListenerNetwork: "tcp",
}

func TestAddRouteToPotetoDebug(t *testing.T) {
	poteto := NewWithOption(debugOption)

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

func TestRunAndStopDebug(t *testing.T) {
	p := NewWithOption(debugOption)

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

func TestRunTLSDebug(t *testing.T) {
	cert, _ := os.ReadFile("./_fixture/certs/cert.pem")
	key, _ := os.ReadFile("./_fixture/certs/key.pem")

	p := NewWithOption(debugOption)

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

func TestSetupServerHandleListenErrorDebug(t *testing.T) {
	defer monkey.UnpatchAll()

	p := NewWithOption(debugOption)
	monkey.Patch(net.Listen, func(pro, add string) (net.Listener, error) {
		return nil, errors.New("error")
	})

	if err := p.setupServer(); err == nil {
		t.Errorf("Unmatched")
	}
}
