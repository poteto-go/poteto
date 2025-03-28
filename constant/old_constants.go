package constant

const VERSION = "v1.8.1"

const (
	MAX_DOMAIN_LENGTH      int    = 255
	PARAM_PREFIX           string = ":"
	PARAM_TYPE_PATH        string = "path"
	PARAM_TYPE_QUERY       string = "query"
	MAX_QUERY_PARAM_LENGTH int    = 32
	ALGORITHM_HS256               = "HS256"
	AUTH_SCHEME                   = "Bearer"
)

const (
	STORE_REQUEST_ID string = "requestId"
)

// Header
const (
	HEADER_ACCESS_CONTROL_ORIGIN string = "Access-Control-Allow-Origin"
	HEADER_ORIGIN                string = "Origin"
	HEADER_VARY                  string = "vary"
	HEADER_CONTENT_TYPE          string = "Content-Type"
	APPLICATION_JSON             string = "application/json"
	CONTENT_SECURITY_POLICY      string = "Content-Security-Policy"
	X_FRAME_OPTION               string = "X-Frame-Options"
	STRICT_TRANSPORT_SECURITY    string = "Strict-Transport-Security"
	X_DOWNLOAD_OPTION            string = "X-Download-Options"
	X_CONTENT_TYPE_OPTION        string = "X-Content-Type-Options"
	REFERRER_POLICY              string = "Referrer-Policy"
	HEADER_AUTHORIZATION         string = "Authorization"
	HEADER_CONTENT_LENGTH        string = "Content-Length"
	HEADER_X_REQUEST_ID          string = "X-Request-Id"
	HEADER_X_FORWARDED_FOR       string = "X-Forwarded-For"
	HEADER_X_REAL_IP             string = "X-Real-Ip"
)

// Workflow
const (
	START_UP_WORKFLOW string = "startUp"
)
