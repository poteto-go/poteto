# Poteto

<img src="assets/logo.svg">

## Simple Web Framework of GoLang

![](https://img.shields.io/badge/go-1.23-lightblue)
![](https://img.shields.io/badge/go-1.24-lightblue)

```bash
go get -u github.com/poteto-go/poteto@latest
```

If you try latest experiment version

https://github.com/poteto-go/poteto/blob/main/EXPERIMENT.md

```bash
go get -u github.com/poteto-go/poteto@exp<version>
```

## Deep Wiki

https://deepwiki.com/poteto-go/poteto

## Quick Start

```go
func main() {
	p := poteto.New()
	p.Register(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	p.GET("/", func(ctx poteto.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Hello World",
		})
	})

	userApi := poteto.Api("/users", func(leaf poteto.Leaf) {
		leaf.GET("/:id", func(ctx poteto.Context) error {
			id, _ := p.PathParam("id")
			return ctx.JSON(http.StatusOK, map[string]string{
				"id": id,
			})
		})
	})

	p.AddApi(userApi)

	p.Run("3000")
}
```

### UT

> [!NOTE]
> Poteto developers can easily test without setting up a server.

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

## OIDC

poteto provides easily oidc middleware.

- verify signature
- jwt schema (if idp google).

```go
func main() {
  p := poteto.New()

  oidcConfig := middleware.OidcConfig {
	  Idp: "google",
		ContextKey: "googleToken",
    CacheMode: true,
    JwksUrl: "https://www.googleapis.com/oauth2/v3/certs",
    CachedVerifyTokenSignature: oidc.CachedVerifyTokenSignature,
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

## Example App For Poteto

TODO

## Poteto-Cli

We support cli tool. But if you doesn't like it, you can create poteto-app w/o cli of course.

You can start hot-reload poteto app.

```sh
go install github.com/poteto-go/poteto-cli/cmd/poteto-cli@latest
```

OR build from docker image

https://hub.docker.com/repository/docker/poteto17/poteto-go/general

```sh
docker pull poteto17/poteto-go
docker -it --rm poteto17/poteto-go:1.23 bash
```

detail on:

https://github.com/poteto-go/poteto-cli
