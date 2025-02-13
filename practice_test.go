package poteto_test

import (
	"net/http"
	"testing"

	"github.com/poteto-go/poteto"
)

func TestPotetoPlay(t *testing.T) {
	p := poteto.New()

	p.GET("/users", func(ctx poteto.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"id":   "1",
			"name": "tester",
		})
	})

	p.POST("/users", func(ctx poteto.Context) error {
		var user map[string]string
		ctx.Bind(user)
		return ctx.JSON(http.StatusOK, map[string]string{
			"id":   "1",
			"name": "tester",
		})
	})

	res := p.Play(http.MethodGet, "/users")
	respBodyStr := res.Body.String()[0 : len(res.Body.String())-1]
	expected := `{"id":"1","name":"tester"}`
	if respBodyStr != expected {
		t.Errorf("unmatched: actual(%s) - expected(%s)", respBodyStr, expected)
	}

	res2 := p.Play(http.MethodPost, "/users", `{"id":"1","name":"tester"}`)
	respBodyStr2 := res2.Body.String()[0 : len(res2.Body.String())-1]
	if respBodyStr2 != expected {
		t.Errorf("unmatched: actual(%s) - expected(%s)", respBodyStr2, expected)
	}
}
