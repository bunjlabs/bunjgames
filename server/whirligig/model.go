package whirligig

import (
	"log"
	"math/rand"
	"os"
	"slices"
	"sync"
	"time"

	"bunjgames-server/common"

	"gopkg.in/yaml.v3"
)

type State string

const (
	StateStart              State = "start"
	StateIntro              State = "intro"
	StateQuestions          State = "questions"
	StateQuestionWhirligig  State = "question_whirligig"
	StateQuestionStart      State = "question_start"
	StateQuestionDiscussion State = "question_discussion"
	StateAnswer             State = "answer"
	StateRightAnswer        State = "right_answer"
	StateQuestionEnd        State = "question_end"
	StateEnd                State = "end"

	MaxScore = 6
)

type Question struct {
	IsProcessed bool `json:"is_processed"`

	Description string `json:"description" yaml:"description"`
	Text        string `json:"text" yaml:"text"`
	Image       string `json:"image" yaml:"image"`
	Audio       string `json:"audio" yaml:"audio"`
	Video       string `json:"video" yaml:"video"`

	AnswerDescription string `json:"answer_description" yaml:"answer_description"`
	AnswerText        string `json:"answer_text" yaml:"answer_text"`
	AnswerImage       string `json:"answer_image" yaml:"answer_image"`
	AnswerAudio       string `json:"answer_audio" yaml:"answer_audio"`
	AnswerVideo       string `json:"answer_video" yaml:"answer_video"`

	Author string `json:"author" yaml:"author"`
}

type GameItem struct {
	Name        string     `json:"name" yaml:"name"`
	Description string     `json:"description" yaml:"description"`
	Type        string     `json:"type" yaml:"type"`
	IsProcessed bool       `json:"is_processed"`
	Questions   []Question `json:"questions" yaml:"questions"`
}

func (item *GameItem) GetTime() int {
	if item.Type == "standard" {
		return 60
	}
	return 20
}

type Game struct {
	Mu sync.Mutex

	Token           string     `json:"-"`
	ConnoissScore   int        `json:"-"`
	ViewersScore    int        `json:"-"`
	CurRandomItem   *int       `json:"-"`
	CurItem         *int       `json:"-"`
	CurQuestion     *int       `json:"-"`
	State           State      `json:"-"`
	TimerPaused     bool       `json:"-"`
	TimerPausedTime int64      `json:"-"`
	TimerTime       int64      `json:"-"`
	Items           []GameItem `json:"-"`
}

func NewGame() *Game {
	return &Game{
		State:       StateStart,
		TimerPaused: true,
	}
}

func (g *Game) getCurItem() *GameItem {
	if g.CurItem == nil {
		return nil
	}
	return &g.Items[*g.CurItem]
}

func (g *Game) setTimer(seconds int) {
	if seconds > 0 {
		g.TimerPaused = false
		g.TimerPausedTime = 0
		g.TimerTime = time.Now().Add(time.Duration(seconds+2) * time.Second).UnixMilli()
	} else {
		g.TimerPaused = true
		g.TimerPausedTime = 0
		g.TimerTime = 0
	}
}

func (g *Game) clearTimer() {
	g.setTimer(0)
}

func (g *Game) ChangeScore(connoissScore, viewersScore int) {
	if connoissScore >= 0 && connoissScore <= MaxScore {
		g.ConnoissScore = connoissScore
	}
	if viewersScore >= 0 && viewersScore <= MaxScore {
		g.ViewersScore = viewersScore
	}
}

func (g *Game) ChangeTimer(paused bool) error {
	if g.State != StateQuestionDiscussion {
		return common.ErrNothingToDo
	}
	now := time.Now().UnixMilli()
	if paused && !g.TimerPaused {
		g.TimerPausedTime = now
	} else if !paused && g.TimerPaused {
		g.TimerTime += now - g.TimerPausedTime
		g.TimerPausedTime = 0
	}
	g.TimerPaused = paused
	return nil
}

func (g *Game) AnswerCorrect(isCorrect bool) error {
	if g.State != StateRightAnswer {
		return common.ErrNothingToDo
	}

	if !isCorrect {
		g.ViewersScore++
	}

	item := &g.Items[*g.CurItem]
	question := &item.Questions[*g.CurQuestion]
	question.IsProcessed = true

	if !isCorrect || *g.CurQuestion == len(item.Questions)-1 {
		if isCorrect {
			g.ConnoissScore++
		}
		item.IsProcessed = true
		g.CurRandomItem = nil
		g.CurItem = nil
		g.CurQuestion = nil
		if g.ConnoissScore == MaxScore || g.ViewersScore == MaxScore || !g.hasUnprocessedItems() {
			g.State = StateEnd
		} else {
			g.State = StateQuestionEnd
		}
	} else {
		next := *g.CurQuestion + 1
		g.CurQuestion = &next
		g.State = StateQuestionStart
	}
	return nil
}

func (g *Game) ExtraTime() error {
	if g.State != StateAnswer {
		return common.ErrNothingToDo
	}
	g.setTimer(g.getCurItem().GetTime())
	g.State = StateQuestionDiscussion
	return nil
}

func (g *Game) hasUnprocessedItems() bool {
	for _, item := range g.Items {
		if !item.IsProcessed {
			return true
		}
	}
	return false
}

func (g *Game) randomiseNextItem() (randomIdx, actualIdx int, err error) {
	n := len(g.Items)
	unprocessed := 0
	for _, item := range g.Items {
		if !item.IsProcessed {
			unprocessed++
		}
	}
	if unprocessed == 0 {
		return 0, 0, &common.BadStateError{Msg: "No items left"}
	}
	randomIdx = rand.Intn(n)
	actualIdx = randomIdx
	for g.Items[actualIdx].IsProcessed {
		actualIdx = (actualIdx + 1) % n
	}
	return randomIdx, actualIdx, nil
}

func (g *Game) NextState(fromState *State) error {
	if fromState != nil && g.State != *fromState {
		return common.ErrNothingToDo
	}
	switch g.State {
	case StateStart:
		g.State = StateIntro
	case StateIntro:
		g.State = StateQuestions
	case StateQuestions:
		randomIdx, actualIdx, err := g.randomiseNextItem()
		if err != nil {
			return err
		}
		g.CurRandomItem = &randomIdx
		g.CurItem = &actualIdx
		zero := 0
		g.CurQuestion = &zero
		g.State = StateQuestionWhirligig
	case StateQuestionWhirligig:
		g.State = StateQuestionStart
	case StateQuestionStart:
		g.setTimer(g.getCurItem().GetTime())
		g.State = StateQuestionDiscussion
	case StateQuestionDiscussion:
		g.clearTimer()
		g.State = StateAnswer
	case StateAnswer:
		g.clearTimer()
		g.State = StateRightAnswer
	case StateRightAnswer:
		return common.ErrNothingToDo
	case StateQuestionEnd:
		randomIdx, actualIdx, err := g.randomiseNextItem()
		if err != nil {
			return err
		}
		g.CurRandomItem = &randomIdx
		g.CurItem = &actualIdx
		zero := 0
		g.CurQuestion = &zero
		g.State = StateQuestionWhirligig
	case StateEnd:
		return common.ErrNothingToDo
	default:
		return &common.BadStateError{Msg: "Bad state"}
	}
	return nil
}

func (g *Game) Parse(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return &common.BadFormatError{Msg: "Cannot read file"}
	}
	log.Printf("Parsing game file %s", filename)

	var items []GameItem
	if err := yaml.Unmarshal(data, &items); err != nil {
		return &common.BadFormatError{Msg: "Cannot parse YAML"}
	}
	log.Printf("Parsed items: %+v", items)

	if len(items) == 0 {
		return &common.BadFormatError{Msg: "No items found"}
	}

	if len(items) > 13 {
		return &common.BadFormatError{Msg: "Too many items"}
	}
	g.Items = items
	return nil
}

type GameState struct {
	Token            string     `json:"token"`
	ConnoissScore    int        `json:"connoisseurs_score"`
	ViewersScore     int        `json:"viewers_score"`
	CurItem          *GameItem  `json:"cur_item"`
	CurQuestion      *Question  `json:"cur_question"`
	CurRandomItemIdx *int       `json:"cur_random_item_idx"`
	CurItemIdx       *int       `json:"cur_item_idx"`
	CurQuestionIdx   *int       `json:"cur_question_idx"`
	State            State      `json:"state"`
	Items            []GameItem `json:"items"`
	TimerPaused      bool       `json:"timer_paused"`
	TimerPausedTime  int64      `json:"timer_paused_time"`
	TimerTime        int64      `json:"timer_time"`
	Name             string     `json:"name"`
}

func (g *Game) Serialize() GameState {
	questionStates := []State{
		StateQuestionWhirligig, StateQuestionStart, StateQuestionDiscussion,
		StateAnswer, StateRightAnswer,
	}
	inQuestionState := slices.Contains(questionStates, g.State)

	var curItem *GameItem
	if inQuestionState && g.CurItem != nil {
		curItem = &g.Items[*g.CurItem]
	}

	var curQuestion *Question
	if inQuestionState && curItem != nil && g.CurQuestion != nil {
		curQuestion = &curItem.Questions[*g.CurQuestion]
	}

	return GameState{
		Token:            g.Token,
		ConnoissScore:    g.ConnoissScore,
		ViewersScore:     g.ViewersScore,
		CurItem:          curItem,
		CurQuestion:      curQuestion,
		CurRandomItemIdx: g.CurRandomItem,
		CurItemIdx:       g.CurItem,
		CurQuestionIdx:   g.CurQuestion,
		State:            g.State,
		Items:            g.Items,
		TimerPaused:      g.TimerPaused,
		TimerPausedTime:  g.TimerPausedTime,
		TimerTime:        g.TimerTime,
		Name:             "whirligig",
	}
}
