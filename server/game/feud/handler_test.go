package feud

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateYaml(roundQuestions int, finalQuestions int, answersPerQuestion int) string {
	var b strings.Builder
	b.WriteString("questions:\n")
	for i := range roundQuestions {
		b.WriteString(fmt.Sprintf("  - text: Round Q%d\n", i+1))
		b.WriteString("    answers:\n")
		for j := range answersPerQuestion {
			b.WriteString(fmt.Sprintf("      - text: A%d_%d\n", i+1, j+1))
			b.WriteString(fmt.Sprintf("        value: %d\n", (j+1)*10))
		}
	}
	b.WriteString("finalQuestions:\n")
	for i := range finalQuestions {
		b.WriteString(fmt.Sprintf("  - text: Final Q%d\n", i+1))
		b.WriteString("    answers:\n")
		b.WriteString(fmt.Sprintf("      - text: FA%d\n", i+1))
		b.WriteString(fmt.Sprintf("        value: %d\n", (i+1)*20))
	}
	return b.String()
}

func TestGame(test *testing.T) {
	test.Parallel()

	game := NewGame()
	if game.GetToken() == "" {
		test.Fatalf("game token is '%s'", game.GetToken())
	}

	err := game.Parse(strings.NewReader(generateYaml(3, 5, 4)))
	assert.Nil(test, err)
	assert.Equal(test, 3, len(game.Questions))
	assert.Equal(test, 5, len(game.FinalQuestions))
	assert.Equal(test, "waiting_for_players", game.State)

	err = game.RegisterPlayer("TEAM1")
	assert.Nil(test, err)
	err = game.RegisterPlayer("TEAM2")
	assert.Nil(test, err)
	assert.Equal(test, 2, len(game.Players))

	err = game.RegisterPlayer("TEAM1")
	assert.Nil(test, err)
	assert.Equal(test, 2, len(game.Players))

	_, err = game.ProcessCommand("next", map[string]any{"from": "waiting_for_players"})
	assert.Nil(test, err)
	assert.Equal(test, "intro", game.State)

	err = game.RegisterPlayer("LATE")
	assert.NotNil(test, err)

	_, err = game.ProcessCommand("next", map[string]any{"from": "intro"})
	assert.Nil(test, err)
	assert.Equal(test, "round", game.State)

	_, err = game.ProcessCommand("next", map[string]any{"from": "round"})
	assert.Nil(test, err)
	assert.Equal(test, "button", game.State)
	assert.NotNil(test, game.Question)
	assert.Equal(test, "Round Q1", game.Question.Text)

	_, err = game.ProcessCommand("buttonClick", map[string]any{"player": "TEAM1"})
	assert.Nil(test, err)
	assert.Equal(test, game.Players[0], game.Answerer)

	_, err = game.ProcessCommand("answer", map[string]any{"correct": true, "answerIndex": 0.0})
	assert.Nil(test, err)
	assert.True(test, game.Question.Answers[0].IsOpened)
	assert.Equal(test, game.Players[1], game.Answerer)

	_, err = game.ProcessCommand("setAnswerer", map[string]any{"player": "TEAM1"})
	assert.Nil(test, err)
	assert.Equal(test, "answers", game.State)
	assert.Equal(test, game.Players[0], game.Answerer)

	_, err = game.ProcessCommand("answer", map[string]any{"correct": true, "answerIndex": 1.0})
	assert.Nil(test, err)
	assert.True(test, game.Question.Answers[1].IsOpened)

	_, err = game.ProcessCommand("answer", map[string]any{"correct": false})
	assert.Nil(test, err)
	assert.Equal(test, 1, game.Answerer.Strikes)

	_, err = game.ProcessCommand("answer", map[string]any{"correct": false})
	assert.Nil(test, err)
	assert.Equal(test, 2, game.Answerer.Strikes)

	_, err = game.ProcessCommand("answer", map[string]any{"correct": false})
	assert.Nil(test, err)
	assert.Equal(test, game.Players[1], game.Answerer)

	_, err = game.ProcessCommand("answer", map[string]any{"correct": true, "answerIndex": 2.0})
	assert.Nil(test, err)
	assert.Equal(test, "answers_reveal", game.State)
	expectedScore := 10 + 20 + 30
	assert.Equal(test, expectedScore, game.Players[1].Score)

	for game.State == "answers_reveal" {
		_, err = game.ProcessCommand("next", map[string]any{"from": "answers_reveal"})
		assert.Nil(test, err)
	}

	assert.Equal(test, "round", game.State)
	assert.Equal(test, 2, game.Round)

	_, err = game.ProcessCommand("next", map[string]any{"from": "round"})
	assert.Nil(test, err)
	assert.Equal(test, "button", game.State)

	_, err = game.ProcessCommand("next", map[string]any{"from": "button"})
	assert.EqualError(test, err, "nothing to do")

	_, err = game.ProcessCommand("setAnswerer", map[string]any{"player": "TEAM2"})
	assert.Nil(test, err)
	assert.Equal(test, "answers", game.State)
	assert.Equal(test, game.Players[1], game.Answerer)

	for i := range 4 {
		_, err = game.ProcessCommand("answer", map[string]any{"correct": true, "answerIndex": i})
		assert.Nil(test, err)
	}
	assert.Equal(test, "answers_reveal", game.State)

	for game.State == "answers_reveal" {
		_, err = game.ProcessCommand("next", map[string]any{"from": "answers_reveal"})
		assert.Nil(test, err)
	}

	assert.Equal(test, "round", game.State)
	assert.Equal(test, 3, game.Round)

	_, err = game.ProcessCommand("next", map[string]any{"from": "round"})
	assert.Nil(test, err)

	_, err = game.ProcessCommand("setAnswerer", map[string]any{"player": "TEAM1"})
	assert.Nil(test, err)

	for i := range 4 {
		_, err = game.ProcessCommand("answer", map[string]any{"correct": true, "answerIndex": i})
		assert.Nil(test, err)
	}

	for game.State == "answers_reveal" {
		_, err = game.ProcessCommand("next", map[string]any{"from": "answers_reveal"})
		assert.Nil(test, err)
	}

	assert.Equal(test, "final", game.State)
	assert.NotNil(test, game.Answerer)

	_, err = game.ProcessCommand("next", map[string]any{"from": "final"})
	assert.Nil(test, err)
	assert.Equal(test, "final_questions", game.State)
	assert.NotNil(test, game.Question)

	for game.State == "final_questions" {
		_, err = game.ProcessCommand("answer", map[string]any{"correct": true, "answerIndex": 0.0})
		assert.Nil(test, err)
	}

	assert.Equal(test, "final_questions_reveal", game.State)

	for game.State == "final_questions_reveal" {
		_, err = game.ProcessCommand("next", map[string]any{"from": "final_questions_reveal"})
		assert.Nil(test, err)
	}

	assert.Equal(test, "final", game.State)
	assert.Equal(test, 2, game.Round)

	_, err = game.ProcessCommand("next", map[string]any{"from": "final"})
	assert.Nil(test, err)

	for game.State == "final_questions" {
		_, err = game.ProcessCommand("answer", map[string]any{"correct": false})
		assert.Nil(test, err)
	}

	for game.State == "final_questions_reveal" {
		_, err = game.ProcessCommand("next", map[string]any{"from": "final_questions_reveal"})
		assert.Nil(test, err)
	}

	assert.Equal(test, "end", game.State)

	_, err = game.ProcessCommand("next", map[string]any{"from": "end"})
	assert.EqualError(test, err, "nothing to do")

	_, err = game.ProcessCommand("unknown_method", map[string]any{})
	assert.EqualError(test, err, "unknown method")
}

func TestRegisterPlayerReplacement(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(1, 5, 2)))
	assert.Nil(test, err)

	err = game.RegisterPlayer("TEAM1")
	assert.Nil(test, err)
	err = game.RegisterPlayer("TEAM2")
	assert.Nil(test, err)
	assert.Equal(test, 2, len(game.Players))
	assert.Equal(test, "TEAM1", game.Players[0].Name)
	assert.Equal(test, "TEAM2", game.Players[1].Name)

	err = game.RegisterPlayer("TEAM3")
	assert.Nil(test, err)
	assert.Equal(test, 2, len(game.Players))
	assert.Equal(test, "TEAM3", game.Players[0].Name)
	assert.Equal(test, "TEAM2", game.Players[1].Name)

	err = game.RegisterPlayer("TEAM2")
	assert.Nil(test, err)
	assert.Equal(test, 2, len(game.Players))

	err = game.RegisterPlayer("TEAM4")
	assert.Nil(test, err)
	assert.Equal(test, 2, len(game.Players))
	assert.Equal(test, "TEAM4", game.Players[0].Name)
	assert.Equal(test, "TEAM2", game.Players[1].Name)
}

func TestStrikesPassToOpponent(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(1, 5, 4)))
	assert.Nil(test, err)

	err = game.RegisterPlayer("T1")
	assert.Nil(test, err)
	err = game.RegisterPlayer("T2")
	assert.Nil(test, err)

	_, _ = game.ProcessCommand("next", map[string]any{"from": "waiting_for_players"})
	_, _ = game.ProcessCommand("next", map[string]any{"from": "intro"})
	_, _ = game.ProcessCommand("next", map[string]any{"from": "round"})

	_, _ = game.ProcessCommand("setAnswerer", map[string]any{"player": "T1"})
	assert.Equal(test, "answers", game.State)

	_, _ = game.ProcessCommand("answer", map[string]any{"correct": false})
	_, _ = game.ProcessCommand("answer", map[string]any{"correct": false})
	_, _ = game.ProcessCommand("answer", map[string]any{"correct": false})
	assert.Equal(test, game.Players[1], game.Answerer)

	_, _ = game.ProcessCommand("answer", map[string]any{"correct": false})
	_, _ = game.ProcessCommand("answer", map[string]any{"correct": false})
	_, _ = game.ProcessCommand("answer", map[string]any{"correct": false})
	assert.Equal(test, "answers_reveal", game.State)
	assert.Equal(test, 0, game.Players[0].Score)
	assert.Equal(test, 0, game.Players[1].Score)
}

func TestTick(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(1, 5, 2)))
	assert.Nil(test, err)

	cmd, err := game.Tick(time.Second)
	assert.Nil(test, err)
	assert.Nil(test, cmd)

	err = game.RegisterPlayer("T1")
	assert.Nil(test, err)
	err = game.RegisterPlayer("T2")
	assert.Nil(test, err)

	_, _ = game.ProcessCommand("next", map[string]any{"from": "waiting_for_players"})
	_, _ = game.ProcessCommand("next", map[string]any{"from": "intro"})
	_, _ = game.ProcessCommand("next", map[string]any{"from": "round"})

	_, _ = game.ProcessCommand("setAnswerer", map[string]any{"player": "T1"})

	for i := range 2 {
		_, _ = game.ProcessCommand("answer", map[string]any{"correct": true, "answerIndex": i})
	}

	for game.State == "answers_reveal" {
		_, _ = game.ProcessCommand("next", map[string]any{"from": "answers_reveal"})
	}
	_, _ = game.ProcessCommand("next", map[string]any{"from": "final"})

	assert.Equal(test, "final_questions", game.State)
	game.Timer = 5000

	cmd, err = game.Tick(3 * time.Second)
	assert.Nil(test, err)
	assert.Nil(test, cmd)
	assert.Equal(test, int64(2000), game.Timer)

	cmd, err = game.Tick(3 * time.Second)
	assert.Nil(test, err)
	assert.NotNil(test, cmd)
	assert.Equal(test, int64(0), game.Timer)
}
