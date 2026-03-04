package whirligig

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string { return &s }

func generateYaml(n int) string {
	var b strings.Builder
	for range n {
		b.WriteString(fmt.Sprintf(`- name: item
  type: standard
  questions:
    - description: q%d
      text: t
      answer:
        description: a%d
        text: x
`, n, n))
	}
	return b.String()
}

type Command struct {
	method  string
	payload map[string]any
}

func TestGame(test *testing.T) {
	test.Parallel()

	game := NewGame()
	if game.GetToken() == "" {
		test.Fatalf("game token is '%s'", game.GetToken())
	}

	err := game.Parse(strings.NewReader(generateYaml(13)))
	assert.Nil(test, err)
	assert.Equal(test, len(game.Items), 13)
	assert.Equal(test, game.State.Value, "start")

	commands := []Command{
		{"next", map[string]any{"from": strPtr("start")}},
		{"next", map[string]any{"from": strPtr("intro")}},
		{"next", map[string]any{"from": strPtr("questions")}},
		{"next", map[string]any{"from": strPtr("question_whirligig")}},
		{"next", map[string]any{"from": strPtr("question_start")}},
	}

	for _, command := range commands {
		assert.Equal(test, game.State.Value, *command.payload["from"].(*string))
		_, err = game.ProcessCommand(command.method, command.payload)
		assert.Nil(test, err)
	}

	assert.Equal(test, game.Timer.Paused, false)
	assert.Equal(test, game.Timer.PausedTime, int64(0))
	assert.Greater(test, game.Timer.Time, int64(0))

	_, err = game.ProcessCommand("timer", map[string]any{"paused": true})
	assert.Nil(test, err)
	assert.Equal(test, game.Timer.Paused, true)
	assert.Greater(test, game.Timer.PausedTime, int64(0))

	_, err = game.ProcessCommand("timer", map[string]any{"paused": false})
	assert.Nil(test, err)
	assert.Equal(test, game.Timer.Paused, false)
	assert.Equal(test, game.Timer.PausedTime, int64(0))

	_, err = game.ProcessCommand("next", map[string]any{"from": strPtr("question_discussion")})
	assert.Nil(test, err)

	_, err = game.ProcessCommand("extraTime", nil)
	assert.Nil(test, err)
	assert.Equal(test, game.State.Value, "question_discussion")

	_, err = game.ProcessCommand("next", map[string]any{"from": strPtr("question_discussion")})
	_, err = game.ProcessCommand("next", map[string]any{"from": strPtr("answer")})
	assert.Nil(test, err)

	_, err = game.ProcessCommand("timer", map[string]any{"paused": false})
	assert.EqualError(test, err, "nothing to do")

	assert.Equal(test, game.State.Value, "right_answer")
	assert.NotNil(test, game.State.Item)
	assert.Equal(test, game.State.Item.IsProcessed, false)
	assert.Equal(test, game.State.Item, &game.Items[*game.State.WhirligigPosition])
	assert.NotNil(test, game.State.Question)
	assert.Equal(test, game.State.Question, &game.State.Item.Questions[0])

	_, err = game.ProcessCommand("next", map[string]any{"from": strPtr("right_answer")})
	assert.EqualError(test, err, "nothing to do")

	assert.Equal(test, game.Score, Score{Connoisseurs: 0, Viewers: 0})

	_, err = game.ProcessCommand("score", map[string]any{"connoisseurs": 2, "viewers": 3})
	assert.Nil(test, err)
	assert.Equal(test, game.Score, Score{Connoisseurs: 2, Viewers: 3})

	_, err = game.ProcessCommand("answer", map[string]any{"correct": false})
	assert.Nil(test, err)
	assert.Equal(test, game.State.Value, "question_end")
	assert.Equal(test, game.Score, Score{Connoisseurs: 2, Viewers: 4})

	_, err = game.ProcessCommand("extraTime", nil)
	assert.EqualError(test, err, "nothing to do")

	_, err = game.ProcessCommand("next", map[string]any{"from": strPtr("question_end")})
	assert.Nil(test, err)
	assert.Equal(test, game.State.Value, "question_whirligig")

	commands = []Command{
		{"next", map[string]any{"from": strPtr("question_whirligig")}},
		{"next", map[string]any{"from": strPtr("question_start")}},
		{"next", map[string]any{"from": strPtr("question_discussion")}},
		{"next", map[string]any{"from": strPtr("answer")}},
		{"answer", map[string]any{"correct": false}},
		{"next", map[string]any{"from": strPtr("question_end")}},

		{"next", map[string]any{"from": strPtr("question_whirligig")}},
		{"next", map[string]any{"from": strPtr("question_start")}},
		{"next", map[string]any{"from": strPtr("question_discussion")}},
		{"next", map[string]any{"from": strPtr("answer")}},
		{"answer", map[string]any{"correct": false}},
	}

	for _, command := range commands {
		if command.method == "next" {
			assert.Equal(test, game.State.Value, *command.payload["from"].(*string))
		}
		_, err = game.ProcessCommand(command.method, command.payload)
		assert.Nil(test, err)
	}

	assert.Equal(test, game.State.Value, "end")
	_, err = game.ProcessCommand("next", map[string]any{"from": strPtr("end")})
	assert.EqualError(test, err, "nothing to do")
}
