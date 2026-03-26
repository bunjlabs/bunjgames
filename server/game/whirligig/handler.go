package whirligig

import (
	"bunjgames/game/abstract"
	"errors"
	"io"
	"log"
	"math/rand"
	"time"

	"gopkg.in/yaml.v3"
)

const epsilon = int64(500)

func (item *Item) GetTime() int64 {
	if item.Type == "standard" {
		return 60*1000 + epsilon
	}
	return 20*1000 + epsilon
}

func NewGame() *Game {
	game := &Game{}
	game.Initialise()
	game.Type = "whirligig"
	game.State.Value = "start"
	return game
}

func (game *Game) Parse(fileStream io.Reader) error {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	var items []Item
	if err := yaml.NewDecoder(fileStream).Decode(&items); err != nil {
		return err
	}
	log.Printf("Parsed items: %+v", items)

	if len(items) == 0 {
		return errors.New("no items found")
	}
	if len(items) > 13 {
		return errors.New("too many items")
	}

	game.Items = items
	return nil
}

func (game *Game) Tick(delta time.Duration) (*abstract.Command, error) {
	game.Mutex.RLock()
	defer game.Mutex.RUnlock()

	if game.State.Value == "question_discussion" && !game.Timer.Paused {
		previousTime := game.Timer.Time
		game.Timer.Time -= delta.Milliseconds()
		if previousTime == game.State.Item.GetTime() {
			return &abstract.Command{"intercom", "timer_begin"}, nil
		}
		if game.State.Item.Type == "standard" && game.Timer.Time < 10*1000 && previousTime >= 10*1000 {
			return &abstract.Command{"intercom", "timer_warning"}, nil
		}
		if game.Timer.Time < 0 && previousTime >= 0 {
			return &abstract.Command{"intercom", "timer_end"}, nil
		}
		if game.Timer.Time < -epsilon {
			return &abstract.Command{"game", game}, game.nextState("question_discussion")
		}
	}
	return nil, nil
}

func (game *Game) ProcessCommand(method string, params map[string]any) (*abstract.Command, error) {
	game.Mutex.RLock()
	defer game.Mutex.RUnlock()

	gameCommand := &abstract.Command{
		Type:    "game",
		Message: game,
	}

	switch method {
	case "next":
		from, _ := params["from"].(string)
		return gameCommand, game.nextState(from)
	case "score":
		connoisseurs, connoisseursOk := params["connoisseurs"].(float64)
		viewers, viewersOk := params["viewers"].(float64)

		if !connoisseursOk || !viewersOk {
			return nil, abstract.InvalidInputs
		}
		game.Score.Connoisseurs = int(connoisseurs)
		game.Score.Viewers = int(viewers)
		return gameCommand, nil
	case "timer":
		paused, ok := params["paused"].(bool)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		game.Timer.Paused = paused
		return gameCommand, nil
	case "answer":
		correct, ok := params["correct"].(bool)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.answerCorrect(correct)
	case "extraTime":
		if game.State.Value == "answer" {
			game.State.Value = "question_start"
			return gameCommand, game.nextState("question_start")
		}
		return nil, abstract.NothingToDo
	default:
		return nil, abstract.UnknownMethod
	}
}

func (game *Game) nextState(fromState string) error {
	if fromState != "" && game.State.Value != fromState {
		return abstract.NothingToDo
	}

	switch game.State.Value {
	case "start":
		game.State.Value = "intro"
	case "intro":
		game.State.Value = "questions"
	case "questions", "question_end":
		randomIdx, item, err := game.randomiseNextItem()
		if err != nil {
			return err
		}
		game.State.WhirligigPosition = &randomIdx
		game.State.Item = item
		game.State.Question = &item.Questions[0]
		game.State.Value = "question_whirligig"
	case "question_whirligig":
		game.State.Value = "question_start"
	case "question_start":
		if game.State.Item == nil {
			return errors.New("no item selected")
		}
		game.Timer.Paused = false
		game.Timer.Time = game.State.Item.GetTime()
		game.State.Value = "question_discussion"
	case "question_discussion":
		game.State.Value = "answer"
	case "answer":
		game.State.Value = "right_answer"
	case "right_answer", "end":
		return abstract.NothingToDo
	default:
		return abstract.InvalidInputs
	}
	return nil
}

func (game *Game) randomiseNextItem() (int, *Item, error) {
	position := rand.Intn(len(game.Items))
	for _, item := range game.Items[position:] {
		if !item.IsProcessed {
			return position, &item, nil
		}
	}
	for _, item := range game.Items[:position] {
		if !item.IsProcessed {
			return position, &item, nil
		}
	}
	return 0, nil, errors.New("no items left")
}

func (game *Game) hasUnprocessedItems() bool {
	for _, item := range game.Items {
		if !item.IsProcessed {
			return true
		}
	}
	return false
}

func (game *Game) answerCorrect(isCorrect bool) error {
	if game.State.Value != "right_answer" {
		return abstract.NothingToDo
	}

	game.State.Question.IsProcessed = true
	var nextQuestion *Question = nil
	for index := range game.State.Item.Questions {
		if !game.State.Item.Questions[index].IsProcessed {
			nextQuestion = &game.State.Item.Questions[index]
			break
		}
	}
	if isCorrect && nextQuestion != nil {
		game.State.Question = nextQuestion
		game.State.Value = "question_start"
		return nil
	}

	game.State.Item = nil
	game.State.Question = nil
	game.State.WhirligigPosition = nil
	if isCorrect {
		game.Score.Connoisseurs++
	} else {
		game.Score.Viewers++
	}

	if game.Score.Connoisseurs >= 6 || game.Score.Viewers >= 6 || !game.hasUnprocessedItems() {
		game.State.Value = "end"
	} else {
		game.State.Value = "question_end"
	}
	return nil
}
