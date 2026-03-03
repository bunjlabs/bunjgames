package whirligig

import (
	"math/rand"
	"sync"
	"time"

	"bunjgames-server/common"
)

const (
	StateStart              = "start"
	StateIntro              = "intro"
	StateQuestions          = "questions"
	StateQuestionWhirligig  = "question_whirligig"
	StateQuestionStart      = "question_start"
	StateQuestionDiscussion = "question_discussion"
	StateAnswer             = "answer"
	StateRightAnswer        = "right_answer"
	StateQuestionEnd        = "question_end"
	StateEnd                = "end"

	MaxScore = 6

	TypeStandard   = "standard"
	TypeBlitz      = "blitz"
	TypeSuperblitz = "superblitz"
)

type Question struct {
	Number      int     `json:"number"`
	IsProcessed bool    `json:"is_processed"`
	Description string  `json:"description"`
	Text        *string `json:"text"`
	Image       *string `json:"image"`
	Audio       *string `json:"audio"`
	Video       *string `json:"video"`

	AnswerDescription string  `json:"answer_description"`
	AnswerText        *string `json:"answer_text"`
	AnswerImage       *string `json:"answer_image"`
	AnswerAudio       *string `json:"answer_audio"`
	AnswerVideo       *string `json:"answer_video"`

	AuthorName string `json:"author_name"`
	AuthorCity string `json:"author_city"`
}

type GameItem struct {
	Number      int        `json:"number"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        string     `json:"type"`
	IsProcessed bool       `json:"is_processed"`
	Questions   []Question `json:"questions"`
}

func (item *GameItem) GetTime() int {
	if item.Type == TypeStandard {
		return 60
	}
	return 20
}

type Game struct {
	Mu sync.Mutex

	Token           string     `json:"-"`
	Expired         time.Time  `json:"-"`
	ConnoissScore   int        `json:"-"`
	ViewersScore    int        `json:"-"`
	CurRandomItem   *int       `json:"-"`
	CurItem         *int       `json:"-"`
	CurQuestion     *int       `json:"-"`
	State           string     `json:"-"`
	TimerPaused     bool       `json:"-"`
	TimerPausedTime int64      `json:"-"`
	TimerTime       int64      `json:"-"`
	Items           []GameItem `json:"-"`
}

func NewGame() *Game {
	return &Game{
		State:       StateStart,
		Expired:     time.Now().Add(12 * time.Hour),
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
			g.Expired = time.Now().Add(10 * time.Minute)
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

func (g *Game) NextState(fromState *string) error {
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
	root, err := common.ParseXMLFile(filename)
	if err != nil {
		return &common.BadFormatError{Msg: "Cannot parse XML"}
	}

	itemsXML := root.Find("items")
	if itemsXML == nil {
		return &common.BadFormatError{Msg: "No items found"}
	}

	for itemNum, itemXML := range itemsXML.FindAll("item") {
		if itemNum >= 13 {
			return &common.BadFormatError{Msg: "Too many items"}
		}
		item := GameItem{
			Number:      itemNum,
			Name:        itemXML.FindText("name"),
			Description: itemXML.FindText("description"),
			Type:        itemXML.FindText("type"),
		}

		questionsXML := itemXML.Find("questions")
		if questionsXML == nil {
			continue
		}
		for qNum, qXML := range questionsXML.FindAll("question") {
			if qNum >= 3 {
				return &common.BadFormatError{Msg: "Too many questions"}
			}
			answerXML := qXML.Find("answer")
			authorXML := qXML.Find("author")

			q := Question{
				Number:      qNum,
				Description: qXML.FindText("description"),
				Text:        strPtr(qXML.FindText("text")),
				Image:       strPtr(qXML.FindText("image")),
				Audio:       strPtr(qXML.FindText("audio")),
				Video:       strPtr(qXML.FindText("video")),
			}
			if answerXML != nil {
				q.AnswerDescription = answerXML.FindText("description")
				q.AnswerText = strPtr(answerXML.FindText("text"))
				q.AnswerImage = strPtr(answerXML.FindText("image"))
				q.AnswerAudio = strPtr(answerXML.FindText("audio"))
				q.AnswerVideo = strPtr(answerXML.FindText("video"))
			}
			if authorXML != nil {
				q.AuthorName = authorXML.FindText("name")
				q.AuthorCity = authorXML.FindText("city")
			}

			item.Questions = append(item.Questions, q)
		}
		g.Items = append(g.Items, item)
	}
	return nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// --- Serialization ---

type QuestionState struct {
	Number      int     `json:"number"`
	IsProcessed bool    `json:"is_processed"`
	Description string  `json:"description"`
	Text        *string `json:"text"`
	Image       *string `json:"image"`
	Audio       *string `json:"audio"`
	Video       *string `json:"video"`

	AnswerDescription string  `json:"answer_description"`
	AnswerText        *string `json:"answer_text"`
	AnswerImage       *string `json:"answer_image"`
	AnswerAudio       *string `json:"answer_audio"`
	AnswerVideo       *string `json:"answer_video"`

	AuthorName string `json:"author_name"`
	AuthorCity string `json:"author_city"`
}

type ItemState struct {
	Number      int             `json:"number"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
	IsProcessed bool            `json:"is_processed"`
	Questions   []QuestionState `json:"questions"`
}

type GameState struct {
	Token            string         `json:"token"`
	Expired          time.Time      `json:"expired"`
	ConnoissScore    int            `json:"connoisseurs_score"`
	ViewersScore     int            `json:"viewers_score"`
	CurItem          *ItemState     `json:"cur_item"`
	CurQuestion      *QuestionState `json:"cur_question"`
	CurRandomItemIdx *int           `json:"cur_random_item_idx"`
	CurItemIdx       *int           `json:"cur_item_idx"`
	CurQuestionIdx   *int           `json:"cur_question_idx"`
	State            string         `json:"state"`
	Items            []ItemState    `json:"items"`
	TimerPaused      bool           `json:"timer_paused"`
	TimerPausedTime  int64          `json:"timer_paused_time"`
	TimerTime        int64          `json:"timer_time"`
	Name             string         `json:"name"`
}

func serializeQuestion(q *Question) QuestionState {
	return QuestionState{
		Number:            q.Number,
		IsProcessed:       q.IsProcessed,
		Description:       q.Description,
		Text:              q.Text,
		Image:             q.Image,
		Audio:             q.Audio,
		Video:             q.Video,
		AnswerDescription: q.AnswerDescription,
		AnswerText:        q.AnswerText,
		AnswerImage:       q.AnswerImage,
		AnswerAudio:       q.AnswerAudio,
		AnswerVideo:       q.AnswerVideo,
		AuthorName:        q.AuthorName,
		AuthorCity:        q.AuthorCity,
	}
}

func serializeItem(item *GameItem) ItemState {
	qs := make([]QuestionState, len(item.Questions))
	for i := range item.Questions {
		qs[i] = serializeQuestion(&item.Questions[i])
	}
	return ItemState{
		Number:      item.Number,
		Name:        item.Name,
		Description: item.Description,
		Type:        item.Type,
		IsProcessed: item.IsProcessed,
		Questions:   qs,
	}
}

func (g *Game) Serialize() GameState {
	questionStates := []string{
		StateQuestionWhirligig, StateQuestionStart, StateQuestionDiscussion,
		StateAnswer, StateRightAnswer,
	}
	inQuestionState := false
	for _, s := range questionStates {
		if g.State == s {
			inQuestionState = true
			break
		}
	}

	var curItemState *ItemState
	var curQuestionState *QuestionState
	if inQuestionState && g.CurItem != nil {
		is := serializeItem(&g.Items[*g.CurItem])
		curItemState = &is
		if g.CurQuestion != nil {
			qs := serializeQuestion(&g.Items[*g.CurItem].Questions[*g.CurQuestion])
			curQuestionState = &qs
		}
	}

	items := make([]ItemState, len(g.Items))
	for i := range g.Items {
		items[i] = serializeItem(&g.Items[i])
	}

	return GameState{
		Token:            g.Token,
		Expired:          g.Expired,
		ConnoissScore:    g.ConnoissScore,
		ViewersScore:     g.ViewersScore,
		CurItem:          curItemState,
		CurQuestion:      curQuestionState,
		CurRandomItemIdx: g.CurRandomItem,
		CurItemIdx:       g.CurItem,
		CurQuestionIdx:   g.CurQuestion,
		State:            g.State,
		Items:            items,
		TimerPaused:      g.TimerPaused,
		TimerPausedTime:  g.TimerPausedTime,
		TimerTime:        g.TimerTime,
		Name:             "whirligig",
	}
}
