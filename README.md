# Poteto

<img src="assets/logo.svg">

## Simple Web Framework of GoLang

We have confirmed that it works with various versions: go@1.21.x, go@1.22.x, go@1.23.x

```bash
go get -u github.com/poteto-go/poteto@latest
# or
go mod tidy
```

## Example App For Poteto

https://github.com/poteto-go/poteto-examples

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

## Poteto Option
```env
WITH_REQUEST_ID=true
DEBUG_MODE=false
LISTENER_NETWORK=tcp
```

## Feature

### JSONRPCAdapter (`>=0.26.0`)

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
  p := New()

  rpc := TestCalculator{}
  // you can access "/add/Calculator.Add"
  p.POST("/add", func(ctx Context) error {
    return PotetoJsonRPCAdapter[Calculator, AdditionArgs](ctx, &rpc)
  })

  p.Run("8080")
}
```

### Leaf router & middlewareTree (`>=0.21.0`)

```go
func main() {
	p := poteto.New()

	// Leaf >= 0.21.0
	p.Leaf("/users", func(userApi poteto.Leaf) {
		userApi.Register(middleware.CamaraWithConfig(middleware.DefaultCamaraConfig))
		userApi.GET("/", controller.UserHandler)
		userApi.GET("/:name", controller.UserIdHandler)
	})

	p.Run("127.0.0.1:8000")
}
```

### Get RequestId Easily

```go
func handler(ctx poteto.Context) error {
	requestId := ctx.RequestId()
}
```

## How to use

```go:main.go
package main

import (
	"net/http"

	"github.com/poteto-go/poteto"
	"github.com/poteto-go/poteto/middleware"
)

func main() {
	p := poteto.New()

	// CORS
	p.Register(middleware.CORSWithConfig(
		middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		},
	))

	// Leaf >= 0.21.0
	p.Leaf("/users", func(userApi poteto.Leaf) {
		userApi.Register(middleware.CamaraWithConfig(middleware.DefaultCamaraConfig))
		userApi.GET("/", controller.UserHandler)
		userApi.GET("/:name", controller.UserNameHandler)
	})

	p.Run("127.0.0.1:8000")
}

type User struct {
	Name any `json:"name"`
}

func UserHandler(ctx poteto.Context) error {
	user := User{
		Name: "user",
	}
	return ctx.JSON(http.StatusOK, user)
}

func UserNameHandler(ctx poteto.Context) error {
	name, _ := ctx.PathParam("name")
	user := User{
		Name: name,
	}
	return ctx.JSON(http.StatusOK, user)
}

```
