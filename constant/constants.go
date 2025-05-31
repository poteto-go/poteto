package constant

const Version = "v1.10.4"

const (
	// max length of domain /
	MaxDomainLength int = 255

	// path parameter prefix
	ParamPrefix    string = ":"
	ParamTypePath  string = "path"
	ParamTypeQuery string = "query"

	// max count of query param
	MaxQueryParamCount int = 32
)

const (
	AlgorithmHS256 = "HS256"
	AuthScheme     = "Bearer"
)

const (
	StoredRequestId string = "requestId"
)

// Header
const (
	HeaderAccessControlOrigin string = "Access-Control-Allow-Origin"
	HeaderOrigin              string = "Origin"
	HeaderVary                string = "vary"
	HeaderContentType         string = "Content-Type"
	ApplicationJson           string = "application/json"
	ContentSecurityPolicy     string = "Content-Security-Policy"
	XFrameOption              string = "X-Frame-Options"
	StrictTransportSecurity   string = "Strict-Transport-Security"
	XDownloadOption           string = "X-Download-Options"
	XContentTypeOption        string = "X-Content-Type-Options"
	ReferrerPolicy            string = "Referrer-Policy"
	HeaderAuthorization       string = "Authorization"
	HeaderContentLength       string = "Content-Length"
	HeaderRequestId           string = "X-Request-Id"
	HeaderXForwardedFor       string = "X-Forwarded-For"
	HeaderXRealIp             string = "X-Real-Ip"
)

// Workflow
const (
	StartUpWorkflow string = "startUp"
)
