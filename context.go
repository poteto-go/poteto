package poteto

import (
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/harakeishi/gats"
	"github.com/poteto-go/poteto/constant"
	"github.com/poteto-go/poteto/utils"
	"github.com/poteto-go/tslice"
)

type Context interface {
	JSON(code int, value any) error
	JSONRPCError(code int, message string, data string, id int) error
	Bind(object any) error
	WriteHeader(code int)
	JsonSerialize(value any) error
	JsonDeserialize(object any) error

	SetQueryParam(queryParams url.Values)
	SetParam(paramType string, paramUnit ParamUnit)
	PathParam(key string) (string, bool)
	QueryParam(key string) (string, bool)
	DebugParam() (string, bool)
	SetPath(path string)
	GetPath() string
	Set(key string, val any)
	Get(key string) (any, bool)

	GetResponse() *response
	SetResponseHeader(key, value string)
	GetRequest() *http.Request
	GetRequestHeaderParam(key string) string
	ExtractRequestHeaderParam(key string) []string

	NoContent() error

	// set request id to store
	// and return value
	RequestId() string

	GetRemoteIP() (string, error)
	RegisterTrustIPRange(ranges *net.IPNet)
	GetIPFromXFFHeader() (string, error)
	RealIP() (string, error)
	Reset(w http.ResponseWriter, r *http.Request)
	SetLogger(logger any)
	Logger() any
}

type context struct {
	response   Response
	request    *http.Request
	ipHandler  IPHandler
	path       string
	httpParams HttpParam
	store      map[string]any
	logger     any
	lock       sync.RWMutex

	// Method
	binder Binder
}

func NewContext(w http.ResponseWriter, r *http.Request) Context {
	return &context{
		response:   NewResponse(w),
		request:    r,
		ipHandler:  &ipHandler{isTrustPrivateIp: true},
		path:       "",
		httpParams: NewHttpParam(),
		binder:     NewBinder(),
	}
}

func (ctx *context) JSON(code int, value any) error {
	ctx.SetResponseHeader(constant.HEADER_CONTENT_TYPE, constant.APPLICATION_JSON)
	ctx.response.SetStatus(code)
	return ctx.JsonSerialize(value)
}

func (ctx *context) JSONRPCError(code int, message string, data string, id int) error {
	return ctx.JSON(http.StatusOK, map[string]any{
		"result":  nil,
		"jsonrpc": "2.0",
		"error": map[string]any{
			"code":    code,
			"message": message,
			"data":    data,
		},
		"id": id,
	})
}

func (ctx *context) GetPath() string {
	return ctx.path
}

func (ctx *context) SetPath(path string) {
	ctx.path = path
}

func (ctx *context) SetQueryParam(queryParams url.Values) {
	if len(queryParams) > constant.MAX_QUERY_PARAM_LENGTH {
		utils.PotetoPrint("too many query params should be < 32\n")
		return
	}

	for key, value := range queryParams {
		if len(value) == 0 {
			continue
		}

		paramUnit := ParamUnit{key, tslice.ToString(value)}

		ctx.SetParam(constant.PARAM_TYPE_QUERY, paramUnit)
	}
}

func (ctx *context) SetParam(paramType string, paramUnit ParamUnit) {
	ctx.httpParams.AddParam(paramType, paramUnit)
}

func (ctx *context) PathParam(key string) (string, bool) {
	key = constant.PARAM_PREFIX + key
	return ctx.httpParams.GetParam(constant.PARAM_TYPE_PATH, key)
}

func (ctx *context) QueryParam(key string) (string, bool) {
	return ctx.httpParams.GetParam(constant.PARAM_TYPE_QUERY, key)
}

func (ctx *context) Bind(object any) error {
	return ctx.binder.Bind(ctx, object)
}

// DebugParam return all http parameters
// use for debug or log
func (ctx *context) DebugParam() (string, bool) {
	val, err := ctx.httpParams.JsonSerialize()
	if err != nil {
		return "", false
	}

	return string(val), true
}

func (ctx *context) WriteHeader(code int) {
	ctx.response.WriteHeader(code)
}

func (ctx *context) GetResponse() *response {
	return ctx.response.(*response)
}

func (ctx *context) SetResponseHeader(key, value string) {
	ctx.response.SetHeader(key, value)
}

func (ctx *context) GetRequest() *http.Request {
	return ctx.request
}

func (ctx *context) GetRequestHeaderParam(key string) string {
	return ctx.request.Header.Get(key)
}

// extract from header directly
func (ctx *context) ExtractRequestHeaderParam(key string) []string {
	return ctx.request.Header[key]
}

func (ctx *context) JsonSerialize(value any) error {
	encoder := json.NewEncoder(ctx.GetResponse())
	return encoder.Encode(value)
}

func (ctx *context) JsonDeserialize(object any) error {
	decoder := json.NewDecoder(ctx.GetRequest().Body)
	return decoder.Decode(object)
}

func (c *context) NoContent() error {
	c.response.WriteHeader(http.StatusNoContent)
	// to provide the same interface as ctx.JSON()
	return nil
}

func (ctx *context) Set(key string, val any) {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	if ctx.store == nil {
		ctx.store = make(map[string]any)
	}
	ctx.store[key] = val
}

func (ctx *context) Get(key string) (any, bool) {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	val, ok := ctx.store[key]
	return val, ok
}

func (ctx *context) RequestId() string {
	// get from store
	val, ok := ctx.Get(constant.STORE_REQUEST_ID)
	if id, err := gats.ToString(val); ok && err == nil {
		return id
	}

	// get from header
	if id := ctx.GetRequestHeaderParam(constant.HEADER_X_REQUEST_ID); id != "" {
		return id
	}

	// generate uuid@V4
	uuid, err := uuid.NewRandom()
	if err != nil {
		return ""
	}
	return uuid.String()
}

func (ctx *context) GetRemoteIP() (string, error) {
	return ctx.ipHandler.GetRemoteIP(ctx)
}

func (ctx *context) RegisterTrustIPRange(ranges *net.IPNet) {
	ctx.ipHandler.RegisterTrustIPRange(ranges)
}

func (ctx *context) GetIPFromXFFHeader() (string, error) {
	return ctx.ipHandler.GetIPFromXFFHeader(ctx)
}

func (ctx *context) RealIP() (string, error) {
	return ctx.ipHandler.RealIP(ctx)
}

// using same binder
func (ctx *context) Reset(w http.ResponseWriter, r *http.Request) {
	ctx.request = r
	ctx.response = NewResponse(w)
	ctx.httpParams = NewHttpParam()
	ctx.store = make(map[string]any)
}

func (ctx *context) SetLogger(logger any) {
	ctx.logger = logger
}

func (ctx *context) Logger() any {
	return ctx.logger
}
