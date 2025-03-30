package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJwks_find(t *testing.T) {
	keys := jwks{
		Keys: []jwk{
			{
				Kid: "1",
			},
			{
				Kid: "2",
			},
		},
	}

	t.Run("find case", func(t *testing.T) {
		key, err := keys.find("1")
		assert.Nil(t, err)
		assert.Equal(t, "1", key.Kid)
	})

	t.Run("not found case", func(t *testing.T) {
		_, err := keys.find("3")
		assert.NotNil(t, err)
	})
}
