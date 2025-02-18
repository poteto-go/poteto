package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poteto-go/poteto"
	"github.com/poteto-go/poteto/constant"
)

func TestCamaraWithConfigByDefault(t *testing.T) {
	tests := []struct {
		name   string
		config CamaraConfig
	}{
		{
			"Test default config",
			DefaultCamaraConfig,
		},
		{
			"If not provide config run with default config",
			CamaraConfig{
				ContentSecurityPolicy:   "",
				XDownloadOption:         "",
				XFrameOption:            "",
				StrictTransportSecurity: "",
				XContentTypeOption:      "",
				ReferrerPolicy:          "",
			},
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			camara := CamaraWithConfig(it.config)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "https://example.com/test", nil)
			context := poteto.NewContext(w, req)

			handler := func(ctx poteto.Context) error {
				return ctx.JSON(http.StatusOK, TestVal{Name: "test", Val: "val"})
			}

			camara_handler := camara(handler)
			camara_handler(context)
			header := w.Result().Header

			if header[constant.ContentSecurityPolicy][0] != DefaultCamaraConfig.ContentSecurityPolicy {
				t.Errorf("Cannot set CSP")
			}

			if header[constant.XFrameOption][0] != DefaultCamaraConfig.XFrameOption {
				t.Errorf("Cannot set XFO")
			}

			if header[constant.StrictTransportSecurity][0] != DefaultCamaraConfig.StrictTransportSecurity {
				t.Errorf("Cannot set STS")
			}

			if header[constant.XDownloadOption][0] != DefaultCamaraConfig.XDownloadOption {
				t.Errorf("Cannot set XDO")
			}

			if header[constant.XContentTypeOption][0] != DefaultCamaraConfig.XContentTypeOption {
				t.Errorf("Cannot set XCT")
			}

			if header[constant.ReferrerPolicy][0] != DefaultCamaraConfig.ReferrerPolicy {
				t.Errorf("Cannot set RP")
			}
		})
	}
}
