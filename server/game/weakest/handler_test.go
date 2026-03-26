package weakest

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateYaml(questions int, finalQuestions int) string {
	var builder strings.Builder
	builder.WriteString("questions:\n")
	for index := range questions {
		builder.WriteString(fmt.Sprintf("  - question: q%d\n", index))
		builder.WriteString(fmt.Sprintf("    answer: a%d\n", index))
	}
	builder.WriteString("finalQuestions:\n")
	for index := range finalQuestions {
		builder.WriteString(fmt.Sprintf("  - question: q%d\n", index))
		builder.WriteString(fmt.Sprintf("    answer: a%d\n", index))
	}
	return builder.String()
}

type answerTestcase struct {
	correct   bool
	gscore    int
	score     int
	bank      int
	answerer  int
	question  int
	fquestion int
}

func TestGame(test *testing.T) {
	test.Parallel()

	game := NewGame()
	if game.GetToken() == "" {
		test.Fatalf("game token is '%s'", game.GetToken())
	}

	err := game.Parse(strings.NewReader(generateYaml(100, 20)))
	assert.Nil(test, err)
	assert.Equal(test, 100, len(game.Questions))
	assert.Equal(test, 20, len(game.FinalQuestions))
	assert.Equal(test, "waiting_for_players", game.State.Value)

	playerNames := []string{"player0", "player1", "player2", "player3"}
	for index, playerName := range playerNames {
		err := game.RegisterPlayer(playerName)
		assert.Nil(test, err)
		assert.Equal(test, index+1, len(game.Players))
		assert.Equal(test, playerName, game.Players[index].Name)
	}

	_, err = game.ProcessCommand("next", map[string]any{"from": "waiting_for_players"})
	assert.Nil(test, err)
	_, err = game.ProcessCommand("next", map[string]any{"from": "intro"})
	assert.Nil(test, err)
	_, err = game.ProcessCommand("next", map[string]any{"from": "round"})
	assert.Nil(test, err)

	assert.Equal(test, "questions", game.State.Value)
	assert.Equal(test, 1, game.RoundState.Number)
	assert.Greater(test, game.RoundState.Time, int64(0))
	assert.Equal(test, game.Players[0], game.RoundState.Answerer)
	assert.Equal(test, game.Questions[0], game.RoundState.Question)

	assert.Equal(test, game.State.Score, 0)
	assert.Equal(test, game.RoundState.Score, 0)
	assert.Equal(test, game.RoundState.Bank, 0)

	testcases := []answerTestcase{
		{correct: true, gscore: 0, score: 1, bank: 0, answerer: 1, question: 1},
		{correct: true, gscore: 0, score: 2, bank: 0, answerer: 2, question: 2},
		{correct: false, gscore: 0, score: 0, bank: 0, answerer: 3, question: 3},
		{correct: true, gscore: 0, score: 1, bank: 0, answerer: 0, question: 4},
		{correct: true, gscore: 0, score: 2, bank: 0, answerer: 1, question: 5},
		{correct: true, gscore: 0, score: 5, bank: 0, answerer: 2, question: 6},
		{correct: true, gscore: 0, score: 10, bank: 0, answerer: 3, question: 7},
		{correct: true, gscore: 0, score: 15, bank: 0, answerer: 0, question: 8},
		{correct: true, gscore: 0, score: 20, bank: 0, answerer: 1, question: 9},
		{correct: true, gscore: 0, score: 30, bank: 0, answerer: 2, question: 10},
		{correct: true, gscore: 40, score: 0, bank: 0, answerer: 2, question: 10},
	}

	// alter bank income to test tiebreaker
	game.Players[1].BankIncome++
	runAnswerTestcases(test, game, testcases)

	assert.Equal(test, 3, game.Players[0].RightAnswers)
	assert.Equal(test, 3, game.Players[1].RightAnswers)
	assert.Equal(test, 2, game.Players[2].RightAnswers)
	assert.Equal(test, 2, game.Players[3].RightAnswers)
	assert.Equal(test, 40, game.Players[2].BankIncome)
	assert.Equal(test, 0, game.Players[3].BankIncome)

	assert.Equal(test, "weakest_choose", game.State.Value)
	assert.Equal(test, 40, game.State.Score)
	assert.Nil(test, game.RoundState.Kicked)
	assert.Equal(test, game.Players[3], game.RoundState.Weakest)
	assert.Equal(test, game.Players[1], game.RoundState.Strongest)

	_, err = game.ProcessCommand("next", map[string]any{"from": "weakest_choose"})
	assert.NotNil(test, err)

	votes := map[string]string{
		"player0": "player2",
		"player1": "player3",
		"player3": "player2",
	}
	for playerName, voteName := range votes {
		err = game.vote(playerName, voteName)
		assert.Nil(test, err)
	}

	_, err = game.ProcessCommand("next", map[string]any{"from": "weakest_choose"})
	assert.Nil(test, err)
	assert.Equal(test, game.Players[2], game.RoundState.Kicked)

	_, err = game.ProcessCommand("next", map[string]any{"from": "weakest_reveal"})
	assert.Nil(test, err)
	assert.Equal(test, "round", game.State.Value)
	assert.Equal(test, 2, game.RoundState.Number)

	_, err = game.ProcessCommand("next", map[string]any{"from": "round"})
	assert.Nil(test, err)
	assert.Equal(test, "questions", game.State.Value)
	assert.Equal(test, 2, game.RoundState.Number)
	assert.Greater(test, game.RoundState.Time, int64(0))
	assert.Equal(test, game.Players[1], game.RoundState.Answerer)
	assert.Equal(test, game.Questions[11], game.RoundState.Question)

	assert.Equal(test, game.State.Score, 40)
	assert.Equal(test, game.RoundState.Score, 0)
	assert.Equal(test, game.RoundState.Bank, 0)

	prevRoundTime := game.RoundState.Time
	prevQuestionTime := game.RoundState.QuestionTime
	_, err = game.Tick(time.Second)
	assert.Nil(test, err)
	assert.Less(test, game.RoundState.Time, prevRoundTime)
	assert.Less(test, game.RoundState.QuestionTime, prevQuestionTime)

	testcases = []answerTestcase{
		{correct: true, gscore: 40, score: 1, bank: 0, answerer: 3, question: 12},
		{correct: true, gscore: 40, score: 2, bank: 0, answerer: 0, question: 13},
		{correct: false, gscore: 40, score: 0, bank: 0, answerer: 1, question: 14},
		{correct: true, gscore: 40, score: 1, bank: 0, answerer: 3, question: 15},
		{correct: true, gscore: 40, score: 2, bank: 0, answerer: 0, question: 16},
		{correct: true, gscore: 40, score: 5, bank: 0, answerer: 1, question: 17},
	}
	runAnswerTestcases(test, game, testcases)
	_, err = game.Tick(time.Second)
	assert.Nil(test, err)
	err = game.bank(false)
	assert.Nil(test, err)

	testcases = []answerTestcase{
		{correct: true, gscore: 40, score: 1, bank: 5, answerer: 3, question: 18},
	}
	runAnswerTestcases(test, game, testcases)
	_, err = game.Tick(4 * time.Second)
	assert.Nil(test, err)
	err = game.bank(false)
	assert.NotNil(test, err)

	err = game.bank(true)
	assert.Nil(test, err)
	assert.Equal(test, 6, game.RoundState.Bank)

	_, err = game.Tick(150 * time.Second)
	assert.Nil(test, err)
	assert.Equal(test, "weakest_choose", game.State.Value)

	votes = map[string]string{
		"player0": "player3",
		"player1": "player0",
		"player3": "player3",
	}
	for playerName, voteName := range votes {
		err = game.vote(playerName, voteName)
		assert.Nil(test, err)
	}
	_, err = game.ProcessCommand("next", map[string]any{"from": "weakest_reveal"})
	assert.Nil(test, err)

	err = game.finalAnswerer("player1")
	assert.Nil(test, err)
	assert.Equal(test, game.Players[1], game.RoundState.Answerer)

	testcases = []answerTestcase{
		{correct: true, gscore: 46, score: 0, bank: 0, answerer: 0, fquestion: 1},
		{correct: false, gscore: 46, score: 0, bank: 0, answerer: 1, fquestion: 2},
		{correct: true, gscore: 46, score: 0, bank: 0, answerer: 0, fquestion: 3},
		{correct: false, gscore: 46, score: 0, bank: 0, answerer: 1, fquestion: 4},
		{correct: true, gscore: 46, score: 0, bank: 0, answerer: 0, fquestion: 5},
		{correct: false, gscore: 46, score: 0, bank: 0, answerer: -1, fquestion: 5},
	}
	runAnswerTestcases(test, game, testcases)
	assert.Equal(test, "end", game.State.Value)
	assert.Equal(test, game.Players[1], game.RoundState.Strongest)
	assert.Equal(test, []bool{true, true, true}, game.RoundState.Strongest.FinalScore)
}

func runAnswerTestcases(test *testing.T, game *Game, testcases []answerTestcase) {
	for _, testcase := range testcases {
		err := game.answer(testcase.correct)
		assert.Nil(test, err)
		assert.Equal(test, testcase.gscore, game.State.Score)
		assert.Equal(test, testcase.score, game.RoundState.Score)
		assert.Equal(test, testcase.bank, game.RoundState.Bank)
		if testcase.answerer != -1 {
			assert.Equal(test, game.Players[testcase.answerer], game.RoundState.Answerer)
		} else {
			assert.Nil(test, game.RoundState.Answerer)
		}
		if testcase.fquestion != 0 {
			assert.Equal(test, game.FinalQuestions[testcase.fquestion], game.RoundState.Question)
		} else if testcase.question != 0 {
			assert.Equal(test, game.Questions[testcase.question], game.RoundState.Question)
		}

	}
}
