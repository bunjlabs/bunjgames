package feud

import (
	"bunjgames/game/abstract"
	"errors"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func NewGame() *Game {
	game := &Game{
		State:   "waiting_for_players",
		Round:   1,
		Players: []*Player{},
	}
	game.Initialise()
	game.Type = "feud"
	if os.Getenv("DEBUG") == "true" {
		game.Players = []*Player{{Name: "PLAYER1"}, {Name: "PLAYER2"}}
	}
	return game
}

func (game *Game) Parse(fileStream io.Reader) error {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	data, err := io.ReadAll(fileStream)
	if err != nil {
		return abstract.InvalidInputs
	}

	var parsed ParsedGame
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return errors.New("failed to parse YAML: " + err.Error())
	}

	if len(parsed.Questions) == 0 {
		return errors.New("at least one question is required")
	}
	if len(parsed.FinalQuestions) != 5 {
		return errors.New("exactly five final questions are required")
	}

	game.Questions = parsed.Questions
	game.FinalQuestions = parsed.FinalQuestions
	return nil
}

func (game *Game) Tick(delta time.Duration) (*abstract.Command, error) {
	game.Mutex.RLock()
	defer game.Mutex.RUnlock()

	if game.State == "final_questions" && game.Timer > 0 {
		game.Timer -= delta.Milliseconds()
		if game.Timer <= 0 {
			game.Timer = 0
			return &abstract.Command{Type: "game", Message: game}, nil
		}
	}
	return nil, nil
}

func (game *Game) RegisterPlayer(name string) error {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()
	for _, player := range game.Players {
		if player.Name == name {
			return nil
		}
	}
	if game.State != "waiting_for_players" {
		return errors.New("game is already in progress")
	}
	if len(game.Players) >= 2 {
		game.Players[0] = &Player{Name: name}
	} else {
		game.Players = append(game.Players, &Player{Name: name})
	}
	return nil
}

func (game *Game) ProcessCommand(method string, params map[string]any) (*abstract.Command, error) {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	gameCommand := &abstract.Command{
		Type:    "game",
		Message: game,
	}

	switch method {
	case "next":
		from, _ := params["from"].(string)
		return gameCommand, game.nextState(from)
	case "buttonClick":
		playerName, ok := params["player"].(string)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.buttonClick(playerName)
	case "setAnswerer":
		playerName, ok := params["player"].(string)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.setAnswerer(playerName)
	case "answer":
		isCorrect, ok1 := params["correct"].(bool)
		answerIdx, ok2 := params["answerIndex"].(float64)
		if !ok1 || !ok2 {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.answerQuestion(isCorrect, int(answerIdx))
	default:
		return nil, abstract.UnknownMethod
	}
}

func (game *Game) nextState(fromState string) error {
	if fromState != "" && game.State != fromState {
		return abstract.NothingToDo
	}
	switch game.State {
	case "waiting_for_players":
		if len(game.Players) < 2 {
			return abstract.NotEnoughPlayers
		}
		game.State = "intro"
	case "intro":
		game.State = "round"
	case "round":
		game.State = "button"
		qs := game.getRoundQuestions()
		if len(qs) > 0 {
			game.Question = qs[0]
		}
	case "button":
		return abstract.NothingToDo
	case "answers":
		return abstract.NothingToDo
	case "answers_reveal":
		answer := game.lastUnopenedAnswer()
		if answer == nil {
			game.nextRound()
		} else {
			answer.IsOpened = true
			game.Intercom("right")
		}
	case "final":
		game.State = "final_questions"
		qs := game.getFinalQuestions()
		if len(qs) > 0 {
			game.Question = qs[0]
		}
	case "final_questions":
		game.State = "final_questions_reveal"
	case "final_questions_reveal":
		var processedQ *Question
		for _, q := range game.FinalQuestions {
			if q.IsProcessed {
				processedQ = q
				break
			}
		}
		if processedQ == nil {
			if game.Round == 1 {
				game.State = "final"
				game.Round++
			} else {
				game.State = "end"
			}
			game.Answerer.FinalScore = game.sumFinalAnsweredValues()
			for _, q := range game.FinalQuestions {
				if q.IsProcessed {
					q.IsProcessed = false
				}
			}
			for _, q := range game.FinalQuestions {
				for _, a := range q.Answers {
					if a.IsOpened {
						a.IsOpened = false
					}
				}
			}
		} else {
			processedQ.IsProcessed = false
			game.Intercom("right")
		}
	case "end":
		return abstract.NothingToDo
	default:
		return abstract.InvalidInputs
	}
	return nil
}

func (game *Game) getRoundQuestions() []*Question {
	var result []*Question
	for _, q := range game.Questions {
		if !q.IsProcessed {
			result = append(result, q)
		}
	}
	return result
}

func (game *Game) getFinalQuestions() []*Question {
	var result []*Question
	for _, q := range game.FinalQuestions {
		if !q.IsProcessed {
			result = append(result, q)
		}
	}
	return result
}

func (game *Game) isFinalState() bool {
	return game.State == "final" || game.State == "final_questions" || game.State == "final_questions_reveal"
}

func (game *Game) getQuestions() []*Question {
	if game.isFinalState() {
		return game.getFinalQuestions()
	}
	return game.getRoundQuestions()
}

func (game *Game) findPlayer(name string) *Player {
	for _, p := range game.Players {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (game *Game) getOpponent(player *Player) *Player {
	for _, p := range game.Players {
		if p.Name != player.Name {
			return p
		}
	}
	return nil
}

func (game *Game) findAnswer(index int) *Answer {
	if game.Question == nil || index < 0 || index >= len(game.Question.Answers) {
		return nil
	}
	return game.Question.Answers[index]
}

func (game *Game) sumOpenedAnswers() int {
	if game.Question == nil {
		return 0
	}
	sum := 0
	for _, a := range game.Question.Answers {
		if a.IsOpened {
			sum += a.Value
		}
	}
	return sum
}

func (game *Game) countUnopenedAnswers() int {
	if game.Question == nil {
		return 0
	}
	count := 0
	for _, a := range game.Question.Answers {
		if !a.IsOpened {
			count++
		}
	}
	return count
}

func (game *Game) lastUnopenedAnswer() *Answer {
	if game.Question == nil {
		return nil
	}
	var last *Answer
	for _, a := range game.Question.Answers {
		if !a.IsOpened {
			last = a
		}
	}
	return last
}

func (game *Game) nextRound() {
	player1 := game.Players[0]
	player2 := game.Players[len(game.Players)-1]

	player1.Strikes = 0
	player2.Strikes = 0

	game.Question.IsProcessed = true

	questions := game.getRoundQuestions()
	if len(questions) > 0 {
		game.Round++
		game.State = "round"
		game.Answerer = nil
	} else {
		if player1.Score >= player2.Score {
			game.Answerer = player1
		} else {
			game.Answerer = player2
		}
		game.Round = 1
		game.State = "final"
	}
}

func (game *Game) sumFinalAnsweredValues() int {
	sum := 0
	for _, q := range game.FinalQuestions {
		for _, a := range q.Answers {
			if a.IsFinalAnswered {
				sum += a.Value
			}
		}
	}
	return sum
}

func (game *Game) buttonClick(playerName string) error {
	if game.State != "button" || game.Answerer != nil {
		return abstract.NothingToDo
	}
	player := game.findPlayer(playerName)
	if player == nil {
		return abstract.InvalidInputs
	}
	game.Answerer = player
	game.Intercom("button")
	return nil
}

func (game *Game) setAnswerer(playerName string) error {
	if game.State != "button" {
		return abstract.NothingToDo
	}
	player := game.findPlayer(playerName)
	if player == nil {
		return abstract.InvalidInputs
	}
	game.Answerer = player
	game.State = "answers"
	return nil
}

func (game *Game) answerQuestion(isCorrect bool, answerIndex int) error {
	if game.State != "button" && game.State != "answers" && game.State != "final_questions" {
		return abstract.NothingToDo
	}
	if game.Answerer == nil {
		return abstract.NothingToDo
	}

	answerer := game.Answerer
	opponent := game.getOpponent(answerer)

	if opponent == nil {
		return errors.New("internal server error: opponent not found")
	}

	if isCorrect {
		answer := game.findAnswer(answerIndex)
		if answer == nil {
			return abstract.InvalidInputs
		}
		if answer.IsOpened {
			return abstract.NothingToDo
		}
		answer.IsOpened = true
		if game.State == "final_questions" {
			answer.IsFinalAnswered = true
		}

		if game.State == "button" {
			game.Answerer = opponent
			game.Intercom("right")
		}
		if game.State == "answers" {
			if opponent.Strikes >= 3 || game.countUnopenedAnswers() == 0 {
				answerer.Score += game.sumOpenedAnswers()
				game.State = "answers_reveal"
			}
			game.Intercom("right")
		}
		if game.State == "final_questions" {
			game.Question.IsProcessed = true
			qs := game.getFinalQuestions()
			if len(qs) > 0 {
				game.Question = qs[0]
			} else {
				game.Question = nil
				game.State = "final_questions_reveal"
			}
		}
	} else if game.State == "button" {
		game.Answerer = opponent
		game.Intercom("wrong")
	} else if game.State == "answers" {
		answerer.Strikes++
		if opponent.Strikes >= 3 {
			opponent.Score += game.sumOpenedAnswers()
			game.State = "answers_reveal"
		} else if answerer.Strikes >= 3 {
			game.Answerer = opponent
		}
		game.Intercom("wrong")
	} else if game.State == "final_questions" {
		game.Question.IsProcessed = true
		qs := game.getFinalQuestions()
		if len(qs) > 0 {
			game.Question = qs[0]
		} else {
			game.Question = nil
			game.State = "final_questions_reveal"
		}
	}
	return nil
}
