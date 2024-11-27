# Poteto

![](https://github.com/user-attachments/assets/7e503083-0af0-4b95-8277-46dfb8166cb9)

## Simple Web Framework of GoLang

```sh
go get github.com/poteto0/poteto@v0.23.1
go mod tidy
```

## Example App For Poteto

https://github.com/poteto0/poteto-sample-api/tree/main

## Feature

### Leaf router & middlewareTree

```go
func main() {
	p := poteto.New()

	// Leaf >= 0.21.0
	p.Leaf("/users", func(userApi poteto.Leaf) {
		userApi.Register(middleware.CamaraWithConfig(middleware.DefaultCamaraConfig))
		userApi.GET("/", controller.UserHandler)
		userApi.GET("/:name", controller.UserIdHandler)
	})

	p.Run(":8000")
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

	"github.com/poteto0/poteto"
	"github.com/poteto0/poteto/middleware"
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
		userApi.GET("/:name", controller.UserIdHandler)
	})

	p.Run(":8000")
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

func UserIdHandler(ctx poteto.Context) error {
	name, _ := ctx.PathParam("name")
	user := User{
		Name: name,
	}
	return ctx.JSON(http.StatusOK, user)
}

```
