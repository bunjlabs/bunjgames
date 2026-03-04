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

func (item *Item) GetTime() int64 {
	if item.Type == "standard" {
		return 60
	}
	return 20
}

func NewGame() *Game {
	game := &Game{}
	game.GenerateToken()
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

func (game *Game) ProcessCommand(method string, params map[string]any) (any, error) {
	game.Mutex.RLock()
	defer game.Mutex.RUnlock()

	switch method {
	case "next":
		from := params["from"].(*string)
		return nil, game.nextState(from)
	case "score":
		connoisseurs, connoisseursOk := params["connoisseurs"].(int)
		viewers, viewersOk := params["viewers"].(int)

		if !connoisseursOk || !viewersOk {
			return nil, abstract.InvalidInputs
		}
		game.Score.Connoisseurs = connoisseurs
		game.Score.Viewers = viewers
		return nil, nil
	case "timer":
		paused, ok := params["paused"].(bool)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return nil, game.changeTimer(paused)
	case "answer":
		correct, ok := params["correct"].(bool)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return nil, game.answerCorrect(correct)
	case "extraTime":
		if game.State.Value == "answer" {
			game.State.Value = "question_start"
			return nil, game.nextState(nil)
		}
		return nil, abstract.NothingToDo
	default:
		return nil, abstract.UnknownMethod
	}
}

func (game *Game) nextState(fromState *string) error {
	if fromState != nil && game.State.Value != *fromState {
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
		game.Timer.PausedTime = 0
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

func (game *Game) randomiseNextItem() (position int, item *Item, err error) {
	position = rand.Intn(len(game.Items))
	for index := range game.Items[position:] {
		if !game.Items[index].IsProcessed {
			return position, &game.Items[index], nil
		}
	}
	for index := range game.Items[:position] {
		if !game.Items[index].IsProcessed {
			return position, &game.Items[index], nil
		}
	}
	return 0, nil, errors.New("no items left")
}

func (game *Game) changeTimer(paused bool) error {
	if game.State.Value != "question_discussion" {
		return abstract.NothingToDo
	}
	now := time.Now().UnixMilli()
	if paused && !game.Timer.Paused {
		game.Timer.PausedTime = now
	} else if !paused && game.Timer.Paused {
		game.Timer.Time += now - game.Timer.PausedTime
		game.Timer.PausedTime = 0
	}
	game.Timer.Paused = paused
	return nil
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
