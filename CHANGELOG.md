# 1.x.x

## 1.9.X

@2025/03/30 ~

### 1.9.1

@2025/03/30

- PATCH: fix jwks url v1 -> v3 by @poteto0 in #290

### 1.9.0

@2025/03/30

- FEAT: oidc middleware verify signature with jwks url by @poteto0 in #285

```go
func main() {
  p := poteto.New()
  oidcConfig := middleware.OidcConfig {
    Idp: "google",
    ContextKey: "googleToken",
    JwksUrl: "https://www.googleapis.com/oauth2/v3/certs",
    CustomVerifyTokenSignature: oidc.DefaultVerifyTokenSignature,
  }
  p.Register(
    middleware.OidcWithConfig(
      oidcConfig,
    )
  )
  p.POST("/login", func(ctx poteto.Context) error {
      var claims oidc.GoogleOidcClaims
      token, _ := ctx.Get("googleToken")
      json.Unmarshal(token.([]byte), &claims)
      ...
      return ctx.JSON(200, map[string]string{"message": "success"})
  })
}
```

- ENH: support custom verify function by @poteto0 in #285
- DOC: readme update by @poteto0 in #288
- CHORE: update security policy in #289

## 1.8.X

@2025/03/27 ~

### 1.8.1

@2025/03/28 ~

- CHORE: add experiment version by @poteto0 in #286
- CHORE: fix quickstart by @poteto0 in #286

### 1.8.0

@2025/03/27 ~

- FEAT: client oidc token & parse it (support google format)

```go
func main() {
  p := poteto.New()
  p.Register(
    middleware.OidcWithConfig(
      middleware.DefaultOidcConfig,
    )
  )

  p.POST("/login", func(ctx poteto.Context) error {
    var claims oidc.GoogleOidcClaims
    token, _ := ctx.Get("googleToken")
    json.Unmarshal(token.([]byte), &claims)
  })
}
```

- BUG: make jwsConfig public

## 1.7.X

@2025/03/23 ~

### 1.7.0

@2025/03/23

- FEAT: `Poteto.AddApi(Poteto)` add router & middlewareTree from `Poteto` by @poteto0 in #274

```go
p := poteto.New()

userApi := poteto.Api("/users", func(leaf Leaf) {
  p.GET("/", <handler>)
})

p.AddApi(userApi)
```

- FEAT: `Api(string, LeafHandler) *poteto` with basePath Poteto by @poteto0 in #274
- CHORE: split benchmark test by @poteto0 in #274
- DOC: readme rework by @poteto0 in #277
- POL: add security policy by @poteto0 in #277
- CHORE: issue template rework by @poteto0 in #277
- CHORE: create pr template by @poteto0 in #277

## 1.6.X

### 1.6.3

@2025/03/22

- VULNS: Bump github.com/golang-jwt/jwt/v5 from 5.2.1 to 5.2.2 in #271

### 1.6.2

@2025/03/21

- REF: ut rework & check bug is not occur by @poteto0 in #268
- Bump github.com/goccy/go-yaml from 1.15.23 to 1.16.0 in #267
- Bump golang.org/x/net from 0.34.0 to 0.36.0 in #264
- Bump github.com/stretchr/testify from 1.8.4 to 1.10.0 in #263

### 1.6.1

- GEMINI: add `rules/.geminirules` by @poteto0 in #261
- REF: refactor ut on context by @poteto0 in #261
- UT: better coverage on context by @poteto0 in #261
- UT: add benchmark for `RequestId` by @poteto0 in #261
- REF: refactor ut on binder by @poteto0 in #260
- CHANGE: now `ctx.Bind` throw `perror.ErrZeroLengthContent` | `perror.ErrNotApplicationJson` by @poteto0 in #260

### 1.6.0

- REF: refactor bash script by @poteto0 in #258
- TYPO: fix typo in workflow by @poteto0 in #258
- OP: add googlecodeassist by @poteto0 in #258
- FEAT: check handler is defined `p.Check(method, path) bool` by @poteto0 in #257
- FEAT: `p.Chain(middlewares)(handler)` make your handler great by @poteto0 in #256

## 1.5.X

### 1.5.1

- OPT: Memory optimization of httpParams by @poteto0 in #250
- REF: refactor with new linter by @poteto0 in #249
- OP: using `golangci-lint` by @poteto0 in #247

### 1.5.0

- FEAT: ctx.BindWithValidate is validate body by github.com/go-playground/validator/v10 by @poteto0 in #240
- REF: Make constant chamelCase (this change is not breaking) by @poteto0 in #237
- OPT: delete no-needed key for route by @poteto0 in #234

## 1.4.X

### 1.4.1

- DOC: godoc on context by @poteto0 in #231
- DOC: godoc on leaf by @poteto0 in #231
- DOC: godoc on router by @poteto0 in #230
- DOC: godoc on response by @poteto0 in #230

### 1.4.0

- FEAT: `response` fullfill interface for `ReponseController` by @poteto0 in #226
- FEAT: `poteto.Play` for ut w/o server by @poteto0 in #226

KEYNOTE:

```go
func main() {
	p := poteto.New()

	p.GET("/users", func(ctx poteto.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"id":   "1",
			"name": "tester",
		})
	})

	res := p.Play(http.MethodGet, "/users")
	resBodyStr := res.Body.String
	// => {"id":"1","name":"tester"}
}
```

## 1.3.X

### 1.3.6

- DEP: uodate tslice `v0.3.0` -> `v0.4.0` by @poteto0 in #223
- REF: rewrote weak functions in tslice for the application by @poteto0 in #223
- REF: rewrote sort that had become confusing by @poteto0 in #223

### 1.3.5

- DEP: `tslice` 0.2.1 -> 0.3.0 by @poteto0 in #222
- REF: using tslice on sort by @poteto0 in #222

### 1.3.4

- DEP: add tslice by @poteto0 in #221
- REF: fix typo & delete no-needed comment by @poteto0 in #221
- Bump github.com/ybbus/jsonrpc/v3 from 3.1.5 to 3.1.6 in #220
- Bump github.com/goccy/go-yaml from 1.15.17 to 1.15.19 in #219

### 1.3.3

- Bump github.com/goccy/go-yaml from 1.15.15 to 1.15.17 in #216
- Bump github.com/goccy/go-json from 0.10.4 to 0.10.5 in #217

### 1.3.2

- BUG: fix not load env by @poteto0 in #215

### 1.3.1

- DEBUG: add some debug mode log by @poteto0 in #213
- FEAT: env for setting poteto option by @poteto0 in #213
- Bump github.com/goccy/go-yaml from 1.15.13 to 1.15.15 in #211

### 1.3.0

- DEP: Poteto's go version update -> 1.23 by @poteto0 in #207
- FEAT: startUpWorkflows run function just before server start in #204
  NOTE:

```go
func main() {
  p := New()

  // run function just before server#Serve
  p.RegisterWorkflow(constant.StartUpWorkflow, 1, func() error)
}
```

## 1.2.X

### 1.2.1

- [poteto-cli] SAME by @poteto0 in https://github.com/poteto-go/poteto-cli/pull/15
- LINT: fix with linter by @poteto0 in #200
- BUG: fix not work linter by @poteto0 in #199

### 1.2.0

- [poteto-cli] DEP: make poteto latest by @poteto0 in https://github.com/poteto-go/poteto-cli/pull/13
- BUG: fix version by @poteto0 in #196
- CI/CD: potetobot comment on PR by @poteto0 in #194
- OP: add linter by @poteto0 in #194

## 1.1.X

### 1.1.1

- CHANGE: devcontainer.json to show git branch by @eaggle23 in #191
- [potet-cli] FEAT: `poteto-cli -v | --version` by @eaggle23 in https://github.com/poteto-go/poteto-cli/pull/9
- [potet-cli] CHANGE: devcontainer.json to show git branch name by @eaggle in https://github.com/poteto-go/poteto-cli/pull/10

### 1.1.0

- DOC: on poteto-cli docker by @poteto0 in #185
- [potet-cli] REF: prepare docker image by @poteto0 in https://github.com/poteto-go/poteto-cli/pull/6
- [potet-cli] BUG: fix wrong dependency by @poteto0 in https://github.com/poteto-go/poteto-cli/pull/6
- [potet-cli] BUG: fix version by @poteto0 in https://github.com/poteto-go/poteto-cli/pull/6
- [potet-cli] CHANGE: template by @poteto0 in https://github.com/poteto-go/poteto-cli/pull/6

## 1.0.x

### 1.0.1

- BUG: fix lf & docker-compose by @poteto0 in https://github.com/poteto-go/poteto-cli/pull/1
- CHANGE: using PotetoPrint by @poteto0 in https://github.com/poteto-go/poteto-cli/pull/1
- DOC: on move by @poteto0 in #182
- OP: poteto-cli & example repo moved by @poteto0 in #182

### 1.0.0

- BREAKING: new mod name by @poteto0
- DOC: fix typo examples by @poteto0 in #177
- FEAT: `poteto-cli new` generate `poteto.yaml`, too. by @poteto0 in #177
- FEAT: `poteto-cli run` start app with hot-reload. by @poteto0 in #177

# 0.x.x

## 0.26.x

### 0.26.5

- BUG: no-inline test; issue below by @poteto0 in #173
  https://github.com/poteto-go/poteto/issues/169
- TEST: ut upgrated by @poteto0 in #173
- REF: split cmd & engine by @poteto0 in #173

### 0.26.4

- REFACT: poteto & middleware by @poteto0 in #171
- FEAT: `ctx.SetResponseHeader(key, value string)` internal call `res.Header().Set(key, value string)` by @poteto0 in #171
- FEAT: `Response.SetHeader(key, value string)` internal call `writer.Header().Set(key, value string)` by @poteto0 in #171
- FEAT: `ctx.GetRequestHeaderParam(key string) string` internal call `req.Header().Get(key string) string` by @poteto0 in #171
- FEAT: `ctx.ExtractRequestHeaderParam(key string) []string` internal call `return req.Header[key]` by @poteto0 in #171
- FEAT: `AddHeader(key, value string)` internal call `writer.Add(key, value string)` by @poteto0 in #171

### 0.26.3

- TEST: fix not ut in poteto-cli by @poteto0 in #168
- FEAT: FEAT: `poteto-cli new -d | --docker` gen with Dockerfile & docker-compose.yaml by @poteto0 in #168
- FEAT: `poteto-cli new -j | --jsonrpc` gen jsonrpc template by @poteto0 in #166

### 0.26.2

- TEST: add benchmark by @poteto0 in #164
- Build(deps): bump github.com/goccy/go-yaml from 1.15.10 to 1.15.13 in #163

### 0.26.1

- EXAMPLE: add example on jsonrpc by @poteto0 in #160
- EXAMPLE: add example on fast-api by @poteto0 in #160
- EXAMPLE: add example on api by @poteto0 in #160
- BUG: fix `PotetoJSONRPCAdapter` dosen't check class by @poteto0 in #158

## 0.26.0

- FEATURE: `PotetoJSONAdapter` provides json rpc by @poteto0 in #154

KeyNote: You can serve JSONRPC server easily.

```go
type (
  Calculator struct{}
  AdditionArgs   struct {
    Add, Added int
  }
)

func (tc *TestCalculator) Add(r *http.Request, args *AdditionArgs) int {
 return args.Add + args.Added
}

func main() {
  p := poteto.New()

  rpc := TestCalculator{}
  // you can access "/add/Calculator.Add"
  p.POST("/add", func(ctx poteto.Context) error {
    return poteto.PotetoJsonRPCAdapter[Calculator, AdditionArgs](ctx, &rpc)
  })

  p.Run("8080")
}
```

- FEATURE: `Poteto.RunTLS` serve https server provided cert & key file by @poteto0 in #144

## 0.25.x

### 0.25.3

- Bump github.com/goccy/go-json from 0.10.3 to 0.10.4 in #152
- Bump github.com/goccy/go-yaml from 1.15.7 to 1.15.10 in #151

### 0.25.2

- BUG: fix ut ignore path by @poteto0 in #147
- OP: test on go@1.21.x, go@1.22.x, go@1.23.x by @poteto0 in #147
- OP: only go@1.23.x upload to codecov by @poteto0 in #148

### 0.25.1

- TEST: ut progress by @poteto0 in #141
- CHANGE: appropriate error messages by @poteto0 in #140
- TEST: ut progress by @poteto0 in #136

### 0.25.0

- FEATURE: poteto-cli released by @poteto0 in #133
- DEPENDENCY: Bump github.com/goccy/go-yaml from 1.15.5 to 1.15.7 in #134

## 0.24.x

### 0.24.0

- FEATURE: mid param router ex /users/:id/name by @poteto0 in #122
- REFACTOR: some switch case by @poteto0 in #122
- FEATURE: ctx.DebugParam by @poteto0 in #125
- OPTIMIZE: middlewareTree by @eaggle23 in #131

## 0.23.x

### 0.23.4

- OPTIMIZE: performance tuning by @poteto0 in #116
- OPTIMIZE: performance tuning by @poteto0 in #117

### 0.23.3

- BUG: fix "/" routes nothing by @poteto0 in #112

### 0.23.2

- OPTIMIZE: optimize router's structure & faster by @poteto0 in #109
- FEATURE: Now, poteto follows patch, head, options, trace, and connect by @poteto0 in #109
- DOCUMENT: update some document by @poteto0 in #109

### 0.23.1

- DOCUMENT: add example app by @poteto0 #104

### 0.23.0

- BUG: fix not allocated Server by @poteto0 #101

## 0.22.x: has critical bug

### 0.22.0

- FEATURE: `Context.RealIP()` return realIp
- CHANGE: `Context.GetIPFromXFFHeader()` return just X-Forwarded-For
- DOCUMENT: update some document

## 0.21.x: has critical bug

### 0.21.0

- FEATURE: `Poteto.Leaf(path, handler)` make router great
- DOCUMENT: Update some document

## 0.20.x: has critical bug

### 0.20.0

- CHANGE: `Poteto.Run()` internal call http.Server#Serve instead of http.ListenAndServe
  You can use your protocol such as udp
- CHANGE: `Poteto.Stop(stdContext)` stop server

## 0.19.x

### 0.19.1

- `PotetoOption`: you can make WithRequestId false
  Because it is slowly With RequestId. If you don't need this, you can make app faster
- fix bug
- refactor something of private func

### 0.19.0

- `Context.Get(key)` get value by key from store.
- `Context.RequestId()` get requestId from Header or store or generate uuid@v4
- `Poteto.ServeHTTP(r, w)` call requestId and set to Header.
  - It may be to become middle ware

## 0.18.x

### 0.18.1

- Fix bug of first msg
- optimize bit

### 0.18.0

- `Poteto` has SetLogger
- You can call ctx.Logger().XXX() from context

## 0.17.x

## 0.17.2

- `Poteto.Run()` will now also accept mere numbers. For example, `8080` is converted to `127.0.0.1:8080` and processed.
- Poteto logged "http://localhost:<port>"

### 0.17.1

- warning handler collision detect

### 0.17.0

- timeout middleware
- poteto.Response.writer became public member

## 0.16.x

### 0.16.1

- fix bug
  - become: `Context.QueryParam()` & `Context.PathParam()` only return string
