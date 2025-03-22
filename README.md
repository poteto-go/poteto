# Poteto

<img src="assets/logo.svg">

## Simple Web Framework of GoLang

We have confirmed that it works with various versions: go@1.21.x, go@1.22.x, go@1.23.x

```bash
go get -u github.com/poteto-go/poteto@latest
```

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

	p.Run()
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
