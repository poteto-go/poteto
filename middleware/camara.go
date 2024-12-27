package middleware

import (
	"github.com/poteto0/poteto"
	"github.com/poteto0/poteto/constant"
)

type CamaraConfig struct {
	ContentSecurityPolicy   string `yaml:"content_security_policy"`
	XFrameOption            string `yaml:"x_frame_option"`
	StrictTransportSecurity string `yaml:"strict_transport_security"`
	XDownloadOption         string `yaml:"x_download_option"`
	XContentTypeOption      string `yaml:"x_content_type_option"`
	ReferrerPolicy          string `yaml:"referrer_policy"`
}

var DefaultCamaraConfig = CamaraConfig{
	ContentSecurityPolicy:   "default-src 'self';base-uri 'self';block-all-mixed-content;font-src 'self' https: data:;frame-ancestors 'self';img-src 'self' data:;object-src 'none';script-src 'self';script-src-attr 'none';style-src 'self' https: 'unsafe-inline';upgrade-insecure-requests",
	XFrameOption:            "SAMEORIGIN",
	StrictTransportSecurity: "max-age=15552000; includeSubDomains",
	XDownloadOption:         "noopen",
	XContentTypeOption:      "nosniff",
	ReferrerPolicy:          "no-referrer",
}

// Provide Some Security Header
func CamaraWithConfig(config CamaraConfig) poteto.MiddlewareFunc {
	if config.ContentSecurityPolicy == "" {
		config.ContentSecurityPolicy = DefaultCamaraConfig.ContentSecurityPolicy
	}

	if config.XDownloadOption == "" {
		config.XDownloadOption = DefaultCamaraConfig.XDownloadOption
	}

	if config.XFrameOption == "" {
		config.XFrameOption = DefaultCamaraConfig.XFrameOption
	}

	if config.StrictTransportSecurity == "" {
		config.StrictTransportSecurity = DefaultCamaraConfig.StrictTransportSecurity
	}

	if config.XContentTypeOption == "" {
		config.XContentTypeOption = DefaultCamaraConfig.XContentTypeOption
	}

	if config.ReferrerPolicy == "" {
		config.ReferrerPolicy = DefaultCamaraConfig.ReferrerPolicy
	}

	return func(next poteto.HandlerFunc) poteto.HandlerFunc {
		return func(ctx poteto.Context) error {
			// * XXS
			// CSP Header
			ctx.SetResponseHeader(
				constant.CONTENT_SECURITY_POLICY,
				config.ContentSecurityPolicy,
			)

			// * Fishing
			// Cannot open in Server
			ctx.SetResponseHeader(
				constant.X_DOWNLOAD_OPTION,
				config.XDownloadOption,
			)

			// * Click Jacking
			// X-Frame-Option: Cannot use in iframe except Same Origin
			ctx.SetResponseHeader(
				constant.X_FRAME_OPTION,
				config.XFrameOption,
			)

			// * Sec Transport
			// Strict-Transport-Security: Required https?
			ctx.SetResponseHeader(
				constant.STRICT_TRANSPORT_SECURITY,
				config.StrictTransportSecurity,
			)

			// * MIME Sniffing
			ctx.SetResponseHeader(
				constant.X_CONTENT_TYPE_OPTION,
				config.XContentTypeOption,
			)

			// * Session HighJack
			ctx.SetResponseHeader(
				constant.REFERRER_POLICY,
				config.ReferrerPolicy,
			)

			return next(ctx)
		}
	}
}
