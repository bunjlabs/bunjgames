package abstract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGame(test *testing.T) {
	test.Parallel()

	game := BaseGame{}
	game.Initialise()

	assert.NotNil(test, game.Token)
	assert.Greater(test, len(game.Token), 0)
}
