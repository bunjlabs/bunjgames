package jeopardy

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateXML(rounds int, themesPerRound int, questionsPerTheme int) string {
	var b strings.Builder
	b.WriteString(`<package><rounds>`)
	for r := range rounds {
		b.WriteString(`<round>`)
		b.WriteString(`<themes>`)
		for t := range themesPerRound {
			name := fmt.Sprintf("theme_%d_%d", r+1, t+1)
			b.WriteString(fmt.Sprintf(`<theme name="%s"><questions>`, name))
			for q := range questionsPerTheme {
				price := (q + 1) * 100 * (r + 1)
				b.WriteString(fmt.Sprintf(`<question price="%d">`, price))
				b.WriteString(fmt.Sprintf(`<scenario><atom>Question %d-%d-%d</atom><atom type="marker"></atom><atom>Post text</atom></scenario>`, r+1, t+1, q+1))
				b.WriteString(fmt.Sprintf(`<right><answer>Answer %d-%d-%d</answer></right>`, r+1, t+1, q+1))
				b.WriteString(`</question>`)
			}
			b.WriteString(`</questions></theme>`)
		}
		b.WriteString(`</themes>`)
		b.WriteString(`</round>`)
	}
	b.WriteString(`</rounds></package>`)
	return b.String()
}

func generateXMLWithFinal(rounds int, themesPerRound int, questionsPerTheme int, finalThemes int) string {
	var b strings.Builder
	b.WriteString(`<package><rounds>`)
	for r := range rounds {
		b.WriteString(`<round>`)
		b.WriteString(`<themes>`)
		for t := range themesPerRound {
			name := fmt.Sprintf("theme_%d_%d", r+1, t+1)
			b.WriteString(fmt.Sprintf(`<theme name="%s"><questions>`, name))
			for q := range questionsPerTheme {
				price := (q + 1) * 100 * (r + 1)
				b.WriteString(fmt.Sprintf(`<question price="%d">`, price))
				b.WriteString(fmt.Sprintf(`<scenario><atom>Question %d-%d-%d</atom></scenario>`, r+1, t+1, q+1))
				b.WriteString(fmt.Sprintf(`<right><answer>Answer %d-%d-%d</answer></right>`, r+1, t+1, q+1))
				b.WriteString(`</question>`)
			}
			b.WriteString(`</questions></theme>`)
		}
		b.WriteString(`</themes>`)
		b.WriteString(`</round>`)
	}
	b.WriteString(`<round><themes>`)
	for t := range finalThemes {
		name := fmt.Sprintf("final_%d", t+1)
		b.WriteString(fmt.Sprintf(`<theme name="%s"><questions>`, name))
		b.WriteString(fmt.Sprintf(`<question price="0"><scenario><atom>Final Q %d</atom></scenario><right><answer>Final A %d</answer></right></question>`, t+1, t+1))
		b.WriteString(`</questions></theme>`)
	}
	b.WriteString(`</themes></round>`)
	b.WriteString(`</rounds></package>`)
	return b.String()
}

func questionKey(themeName string, value int) string {
	return fmt.Sprintf("%s:%d", themeName, value)
}

func TestGame(test *testing.T) {
	test.Parallel()

	game := NewGame()
	if game.GetToken() == "" {
		test.Fatalf("game token is '%s'", game.GetToken())
	}

	err := game.Parse(strings.NewReader(generateXML(2, 3, 4)))
	assert.Nil(test, err)
	assert.Equal(test, 2, len(game.Rounds))
	assert.False(test, game.Rounds[0].IsFinal)
	assert.False(test, game.Rounds[1].IsFinal)
	assert.Equal(test, "waiting_for_players", game.State.Value)

	err = game.RegisterPlayer("PLAYER1")
	assert.Nil(test, err)
	err = game.RegisterPlayer("PLAYER2")
	assert.Nil(test, err)
	err = game.RegisterPlayer("PLAYER3")
	assert.Nil(test, err)
	assert.Equal(test, 3, len(game.Players))

	err = game.RegisterPlayer("PLAYER1")
	assert.Nil(test, err)
	assert.Equal(test, 3, len(game.Players))

	_, err = game.ProcessCommand("next", map[string]any{"from": "waiting_for_players"})
	assert.Nil(test, err)
	assert.Equal(test, "intro", game.State.Value)

	_, err = game.ProcessCommand("next", map[string]any{"from": "intro"})
	assert.Nil(test, err)
	assert.Equal(test, "themes_all", game.State.Value)

	_, err = game.ProcessCommand("next", map[string]any{"from": "themes_all"})
	assert.Nil(test, err)
	assert.Equal(test, "round", game.State.Value)

	_, err = game.ProcessCommand("next", map[string]any{"from": "round"})
	assert.Nil(test, err)
	assert.Equal(test, "round_themes", game.State.Value)

	_, err = game.ProcessCommand("next", map[string]any{"from": "round_themes"})
	assert.Nil(test, err)
	assert.Equal(test, "questions", game.State.Value)

	_, err = game.ProcessCommand("next", map[string]any{"from": "questions"})
	assert.EqualError(test, err, "nothing to do")

	_, err = game.ProcessCommand("chooseQuestion", map[string]any{"question": questionKey("theme_1_1", 100)})
	assert.Nil(test, err)
	assert.Equal(test, "question", game.State.Value)
	assert.NotNil(test, game.State.Question)

	_, err = game.ProcessCommand("next", map[string]any{"from": "question"})
	assert.Nil(test, err)
	assert.Equal(test, "answer", game.State.Value)

	_, err = game.ProcessCommand("buttonClick", map[string]any{"player": "PLAYER1"})
	assert.Nil(test, err)
	assert.Equal(test, game.Players[0], game.State.Answerer)

	_, err = game.ProcessCommand("answer", map[string]any{"correct": true})
	assert.Nil(test, err)
	assert.Equal(test, 100, game.Players[0].Balance)
	assert.Nil(test, game.State.Answerer)
	assert.Equal(test, "question_end", game.State.Value)

	_, err = game.ProcessCommand("next", map[string]any{"from": "question_end"})
	assert.Nil(test, err)
	assert.Equal(test, "questions", game.State.Value)
	assert.Nil(test, game.State.Question)

	_, err = game.ProcessCommand("chooseQuestion", map[string]any{"question": questionKey("theme_1_1", 200)})
	assert.Nil(test, err)
	assert.Equal(test, "question", game.State.Value)

	_, err = game.ProcessCommand("next", map[string]any{"from": "question"})
	assert.Nil(test, err)

	_, err = game.ProcessCommand("buttonClick", map[string]any{"player": "PLAYER2"})
	assert.Nil(test, err)

	_, err = game.ProcessCommand("answer", map[string]any{"correct": false})
	assert.Nil(test, err)
	assert.Equal(test, -200, game.Players[1].Balance)
	assert.Nil(test, game.State.Answerer)

	_, err = game.ProcessCommand("buttonClick", map[string]any{"player": "PLAYER3"})
	assert.Nil(test, err)

	_, err = game.ProcessCommand("answer", map[string]any{"correct": true})
	assert.Nil(test, err)
	assert.Equal(test, 200, game.Players[2].Balance)

	_, err = game.ProcessCommand("skipQuestion", map[string]any{})
	assert.EqualError(test, err, "nothing to do")

	for _, theme := range game.Round.Themes {
		for _, q := range theme.Questions {
			if q.IsProcessed {
				continue
			}
			key := questionKey(theme.Name, q.Value)
			_, err = game.ProcessCommand("chooseQuestion", map[string]any{"question": key})
			assert.Nil(test, err)

			_, err = game.ProcessCommand("skipQuestion", map[string]any{})
			assert.Nil(test, err)

			if game.State.Value == "question_end" {
				_, err = game.ProcessCommand("next", map[string]any{"from": "question_end"})
				assert.Nil(test, err)
			}
		}
	}

	assert.Equal(test, "round", game.State.Value)
	assert.Equal(test, 2, game.Round.Number)

	_, err = game.ProcessCommand("setBalance", map[string]any{"balanceList": []any{100.0, 200.0, 300.0}})
	assert.Nil(test, err)
	assert.Equal(test, 100, game.Players[0].Balance)
	assert.Equal(test, 200, game.Players[1].Balance)
	assert.Equal(test, 300, game.Players[2].Balance)

	_, err = game.ProcessCommand("setRound", map[string]any{"round": 1.0})
	assert.Nil(test, err)
	assert.Equal(test, "round", game.State.Value)
	assert.Equal(test, 1, game.Round.Number)

	_, err = game.ProcessCommand("unknown_method", map[string]any{})
	assert.EqualError(test, err, "unknown method")
}

func TestGameWithFinal(test *testing.T) {
	test.Parallel()

	game := NewGame()

	err := game.Parse(strings.NewReader(generateXMLWithFinal(1, 2, 1, 3)))
	assert.Nil(test, err)
	assert.Equal(test, 2, len(game.Rounds))
	assert.True(test, game.Rounds[1].IsFinal)

	err = game.RegisterPlayer("PLAYER1")
	assert.Nil(test, err)
	err = game.RegisterPlayer("PLAYER2")
	assert.Nil(test, err)

	_, err = game.ProcessCommand("next", map[string]any{"from": "waiting_for_players"})
	assert.Nil(test, err)
	_, err = game.ProcessCommand("next", map[string]any{"from": "intro"})
	assert.Nil(test, err)
	_, err = game.ProcessCommand("next", map[string]any{"from": "themes_all"})
	assert.Nil(test, err)
	_, err = game.ProcessCommand("next", map[string]any{"from": "round"})
	assert.Nil(test, err)
	_, err = game.ProcessCommand("next", map[string]any{"from": "round_themes"})
	assert.Nil(test, err)

	for _, theme := range game.Round.Themes {
		for _, q := range theme.Questions {
			if q.IsProcessed {
				continue
			}
			key := questionKey(theme.Name, q.Value)
			_, err = game.ProcessCommand("chooseQuestion", map[string]any{"question": key})
			assert.Nil(test, err)

			_, err = game.ProcessCommand("next", map[string]any{"from": "question"})
			assert.Nil(test, err)

			_, err = game.ProcessCommand("buttonClick", map[string]any{"player": "PLAYER1"})
			assert.Nil(test, err)

			_, err = game.ProcessCommand("answer", map[string]any{"correct": true})
			assert.Nil(test, err)

			if game.State.Value == "question_end" {
				_, err = game.ProcessCommand("next", map[string]any{"from": "question_end"})
				assert.Nil(test, err)
			}
		}
	}

	assert.Equal(test, "round", game.State.Value)
	assert.Equal(test, 2, game.Round.Number)
	assert.True(test, game.Round.IsFinal)

	_, err = game.ProcessCommand("next", map[string]any{"from": "round"})
	assert.Nil(test, err)
	assert.Equal(test, "final_themes", game.State.Value)

	themes := game.Round.Themes
	_, err = game.ProcessCommand("removeFinalTheme", map[string]any{"theme": themes[0].Name})
	assert.Nil(test, err)
	assert.Equal(test, "final_themes", game.State.Value)

	_, err = game.ProcessCommand("removeFinalTheme", map[string]any{"theme": themes[1].Name})
	assert.Nil(test, err)
	assert.Equal(test, "final_bets", game.State.Value)
	assert.NotNil(test, game.State.Question)

	game.Players[0].Balance = 500
	game.Players[1].Balance = 300

	_, err = game.ProcessCommand("finalBet", map[string]any{"player": "PLAYER1", "bet": 200.0})
	assert.Nil(test, err)
	assert.Equal(test, 200, game.Players[0].FinalBet)

	_, err = game.ProcessCommand("finalBet", map[string]any{"player": "PLAYER2", "bet": 150.0})
	assert.Nil(test, err)
	assert.Equal(test, 150, game.Players[1].FinalBet)

	_, err = game.ProcessCommand("next", map[string]any{"from": "final_bets"})
	assert.Nil(test, err)
	assert.Equal(test, "final_question", game.State.Value)

	_, err = game.ProcessCommand("next", map[string]any{"from": "final_question"})
	assert.Nil(test, err)
	assert.Equal(test, "final_answer", game.State.Value)

	_, err = game.ProcessCommand("finalAnswer", map[string]any{"player": "PLAYER1", "answer": "my answer"})
	assert.Nil(test, err)
	assert.Equal(test, "my answer", game.Players[0].FinalAnswer)

	_, err = game.ProcessCommand("finalAnswer", map[string]any{"player": "PLAYER2", "answer": "other answer"})
	assert.Nil(test, err)

	_, err = game.ProcessCommand("next", map[string]any{"from": "final_answer"})
	assert.Nil(test, err)
	assert.Equal(test, "final_player_answer", game.State.Value)
	assert.NotNil(test, game.State.Answerer)

	_, err = game.ProcessCommand("finalPlayerAnswer", map[string]any{"correct": true})
	assert.Nil(test, err)
	assert.Equal(test, "final_player_bet", game.State.Value)

	_, err = game.ProcessCommand("next", map[string]any{"from": "final_player_bet"})
	assert.Nil(test, err)

	if game.State.Value == "final_player_answer" {
		_, err = game.ProcessCommand("finalPlayerAnswer", map[string]any{"correct": false})
		assert.Nil(test, err)
		_, err = game.ProcessCommand("next", map[string]any{"from": "final_player_bet"})
		assert.Nil(test, err)
	}

	assert.Equal(test, "game_end", game.State.Value)
}

func TestAuctionAndBagcat(test *testing.T) {
	test.Parallel()

	xml := `<package><rounds><round><themes>
		<theme name="T1"><questions>
			<question price="100"><type name="auction"></type>
				<scenario><atom>Auction Q</atom></scenario>
				<right><answer>Auction A</answer></right>
			</question>
			<question price="200"><type name="bagcat"><param name="theme">Secret Theme</param></type>
				<scenario><atom>Bagcat Q</atom></scenario>
				<right><answer>Bagcat A</answer></right>
			</question>
			<question price="300">
				<scenario><atom>Standard Q</atom></scenario>
				<right><answer>Standard A</answer></right>
			</question>
		</questions></theme>
	</themes></round></rounds></package>`

	game := NewGame()
	err := game.Parse(strings.NewReader(xml))
	assert.Nil(test, err)

	err = game.RegisterPlayer("P1")
	assert.Nil(test, err)
	err = game.RegisterPlayer("P2")
	assert.Nil(test, err)

	_, err = game.ProcessCommand("next", map[string]any{"from": "waiting_for_players"})
	assert.Nil(test, err)
	_, err = game.ProcessCommand("next", map[string]any{"from": "intro"})
	assert.Nil(test, err)
	_, err = game.ProcessCommand("next", map[string]any{"from": "themes_all"})
	assert.Nil(test, err)
	_, err = game.ProcessCommand("next", map[string]any{"from": "round"})
	assert.Nil(test, err)
	_, err = game.ProcessCommand("next", map[string]any{"from": "round_themes"})
	assert.Nil(test, err)

	_, err = game.ProcessCommand("chooseQuestion", map[string]any{"question": questionKey("T1", 100)})
	assert.Nil(test, err)
	assert.Equal(test, "question_event", game.State.Value)
	assert.Equal(test, "auction", game.State.Question.Type)

	_, err = game.ProcessCommand("setAnswererAndBet", map[string]any{"player": "P1", "bet": 500.0})
	assert.Nil(test, err)
	assert.Equal(test, "question", game.State.Value)
	assert.Equal(test, game.Players[0], game.State.Answerer)

	_, err = game.ProcessCommand("next", map[string]any{"from": "question"})
	assert.Nil(test, err)
	assert.Equal(test, "answer", game.State.Value)

	_, err = game.ProcessCommand("answer", map[string]any{"correct": true})
	assert.Nil(test, err)
	assert.Equal(test, 500, game.Players[0].Balance)

	if game.State.Value == "question_end" {
		_, err = game.ProcessCommand("next", map[string]any{"from": "question_end"})
		assert.Nil(test, err)
	}

	_, err = game.ProcessCommand("chooseQuestion", map[string]any{"question": questionKey("T1", 200)})
	assert.Nil(test, err)
	assert.Equal(test, "question_event", game.State.Value)
	assert.Equal(test, "bagcat", game.State.Question.Type)

	_, err = game.ProcessCommand("setAnswererAndBet", map[string]any{"player": "P2", "bet": 300.0})
	assert.Nil(test, err)

	_, err = game.ProcessCommand("next", map[string]any{"from": "question"})
	assert.Nil(test, err)

	_, err = game.ProcessCommand("answer", map[string]any{"correct": false})
	assert.Nil(test, err)
	assert.Equal(test, -300, game.Players[1].Balance)
}
