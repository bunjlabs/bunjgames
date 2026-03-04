package feud

import (
	"encoding/xml"
	"sync"
	"time"

	"bunjgames/common"
)

const (
	StateWaitingForPlayers    = "waiting_for_players"
	StateIntro                = "intro"
	StateRound                = "round"
	StateButton               = "button"
	StateAnswers              = "answers"
	StateAnswersReveal        = "answers_reveal"
	StateFinal                = "final"
	StateFinalQuestions       = "final_questions"
	StateFinalQuestionsReveal = "final_questions_reveal"
	StateEnd                  = "end"
)

type Answer struct {
	ID              int    `json:"id"`
	Text            string `json:"text"`
	Value           int    `json:"value"`
	IsOpened        bool   `json:"is_opened"`
	IsFinalAnswered bool   `json:"is_final_answered"`
}

type Question struct {
	ID          int       `json:"id"`
	Text        string    `json:"text"`
	IsFinal     bool      `json:"is_final"`
	IsProcessed bool      `json:"is_processed"`
	Answers     []*Answer `json:"answers"`
}

type Player struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Strikes    int    `json:"strikes"`
	Score      int    `json:"score"`
	FinalScore int    `json:"final_score"`
}

type Game struct {
	Mu sync.Mutex `json:"-"`

	common.IntercomQueue

	Token     string      `json:"token"`
	Expired   time.Time   `json:"expired"`
	Round     int         `json:"round"`
	State     string      `json:"state"`
	Questions []*Question `json:"-"`
	Players   []*Player   `json:"players"`
	Question  *Question   `json:"question"`
	Answerer  *Player     `json:"-"`
	Timer     int64       `json:"timer"`
}

func NewGame() *Game {
	return &Game{
		State:   StateWaitingForPlayers,
		Expired: time.Now().Add(12 * time.Hour),
		Round:   1,
	}
}

func (g *Game) getQuestions() []*Question {
	isFinal := g.State == StateFinal || g.State == StateFinalQuestions || g.State == StateFinalQuestionsReveal
	var result []*Question
	for _, q := range g.Questions {
		if q.IsFinal == isFinal && !q.IsProcessed {
			result = append(result, q)
		}
	}
	return result
}

func (g *Game) findPlayer(id int) *Player {
	for _, p := range g.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (g *Game) getOpponent(player *Player) *Player {
	for _, p := range g.Players {
		if p.ID != player.ID {
			return p
		}
	}
	return nil
}

func (g *Game) findAnswer(id int) *Answer {
	if g.Question == nil {
		return nil
	}
	for _, a := range g.Question.Answers {
		if a.ID == id {
			return a
		}
	}
	return nil
}

func (g *Game) sumOpenedAnswers() int {
	if g.Question == nil {
		return 0
	}
	sum := 0
	for _, a := range g.Question.Answers {
		if a.IsOpened {
			sum += a.Value
		}
	}
	return sum
}

func (g *Game) countUnopenedAnswers() int {
	if g.Question == nil {
		return 0
	}
	count := 0
	for _, a := range g.Question.Answers {
		if !a.IsOpened {
			count++
		}
	}
	return count
}

func (g *Game) lastUnopenedAnswer() *Answer {
	if g.Question == nil {
		return nil
	}
	var last *Answer
	for _, a := range g.Question.Answers {
		if !a.IsOpened {
			last = a
		}
	}
	return last
}

func (g *Game) nextRound() {
	player1 := g.Players[0]
	player2 := g.Players[len(g.Players)-1]

	player1.Strikes = 0
	player2.Strikes = 0

	g.Question.IsProcessed = true

	questions := g.getQuestions()
	if len(questions) > 0 {
		g.Round++
		g.State = StateRound
		g.Answerer = nil
	} else {
		if player1.Score >= player2.Score {
			g.Answerer = player1
		} else {
			g.Answerer = player2
		}
		g.Round = 1
		g.State = StateFinal
	}
}

func (g *Game) sumFinalAnsweredValues() int {
	sum := 0
	for _, q := range g.Questions {
		if !q.IsFinal {
			continue
		}
		for _, a := range q.Answers {
			if a.IsFinalAnswered {
				sum += a.Value
			}
		}
	}
	return sum
}

func (g *Game) NextState(fromState *string) error {
	if fromState != nil && g.State != *fromState {
		return common.ErrNothingToDo
	}
	switch g.State {
	case StateWaitingForPlayers:
		if len(g.Players) >= 2 {
			g.State = StateIntro
		} else {
			return &common.BadStateError{Msg: "Not enough players"}
		}
	case StateIntro:
		g.State = StateRound
	case StateRound:
		g.State = StateButton
		qs := g.getQuestions()
		if len(qs) > 0 {
			g.Question = qs[0]
		}
	case StateButton:
		return common.ErrNothingToDo
	case StateAnswers:
		return common.ErrNothingToDo
	case StateAnswersReveal:
		answer := g.lastUnopenedAnswer()
		if answer == nil {
			g.nextRound()
		} else {
			answer.IsOpened = true
			g.QueueIntercom("right")
		}
	case StateFinal:
		g.State = StateFinalQuestions
		qs := g.getQuestions()
		if len(qs) > 0 {
			g.Question = qs[0]
		}
	case StateFinalQuestions:
		g.State = StateFinalQuestionsReveal
	case StateFinalQuestionsReveal:
		var processedQ *Question
		for _, q := range g.Questions {
			if q.IsFinal && q.IsProcessed {
				processedQ = q
				break
			}
		}
		if processedQ == nil {
			if g.Round == 1 {
				g.State = StateFinal
				g.Round++
			} else {
				g.State = StateEnd
			}
			g.Answerer.FinalScore = g.sumFinalAnsweredValues()
			for _, q := range g.Questions {
				if q.IsFinal && q.IsProcessed {
					q.IsProcessed = false
				}
			}
			for _, q := range g.Questions {
				if q.IsFinal {
					for _, a := range q.Answers {
						if a.IsOpened {
							a.IsOpened = false
						}
					}
				}
			}
		} else {
			processedQ.IsProcessed = false
			g.QueueIntercom("right")
		}
	case StateEnd:
		return common.ErrNothingToDo
	default:
		return &common.BadStateError{Msg: "Bad state"}
	}
	return nil
}

func (g *Game) ButtonClick(playerID int) error {
	if g.State != StateButton || g.Answerer != nil {
		return common.ErrNothingToDo
	}
	player := g.findPlayer(playerID)
	if player == nil {
		return &common.BadStateError{Msg: "Player not found"}
	}
	g.Answerer = player
	g.QueueIntercom("button")
	return nil
}

func (g *Game) SetAnswerer(playerID int) error {
	if g.State != StateButton {
		return common.ErrNothingToDo
	}
	player := g.findPlayer(playerID)
	if player == nil {
		return &common.BadStateError{Msg: "Player not found"}
	}
	g.Answerer = player
	g.State = StateAnswers
	return nil
}

func (g *Game) AnswerQuestion(isCorrect bool, answerID int) error {
	if g.State != StateButton && g.State != StateAnswers && g.State != StateFinalQuestions {
		return common.ErrNothingToDo
	}
	if g.Answerer == nil {
		return common.ErrNothingToDo
	}

	answerer := g.Answerer
	opponent := g.getOpponent(answerer)

	if isCorrect {
		answer := g.findAnswer(answerID)
		if answer == nil {
			return &common.BadStateError{Msg: "Answer not found"}
		}
		if answer.IsOpened {
			return common.ErrNothingToDo
		}
		answer.IsOpened = true
		if g.State == StateFinalQuestions {
			answer.IsFinalAnswered = true
		}

		if g.State == StateButton {
			g.Answerer = opponent
			g.QueueIntercom("right")
		}
		if g.State == StateAnswers {
			if opponent.Strikes >= 3 || g.countUnopenedAnswers() == 0 {
				answerer.Score += g.sumOpenedAnswers()
				g.State = StateAnswersReveal
			}
			g.QueueIntercom("right")
		}
		if g.State == StateFinalQuestions {
			g.Question.IsProcessed = true
			qs := g.getQuestions()
			if len(qs) > 0 {
				g.Question = qs[0]
			} else {
				g.Question = nil
				g.State = StateFinalQuestionsReveal
			}
		}
	} else if g.State == StateButton {
		g.Answerer = opponent
		g.QueueIntercom("wrong")
	} else if g.State == StateAnswers {
		answerer.Strikes++
		if opponent.Strikes >= 3 {
			opponent.Score += g.sumOpenedAnswers()
			g.State = StateAnswersReveal
		} else if answerer.Strikes >= 3 {
			g.Answerer = opponent
		}
		g.QueueIntercom("wrong")
	} else if g.State == StateFinalQuestions {
		g.Question.IsProcessed = true
		qs := g.getQuestions()
		if len(qs) > 0 {
			g.Question = qs[0]
		} else {
			g.Question = nil
			g.State = StateFinalQuestionsReveal
		}
	}
	return nil
}

func (g *Game) RegisterPlayer(name string) *Player {
	for _, p := range g.Players {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (g *Game) AddPlayer(name string) *Player {
	p := &Player{
		ID:   common.NextID(),
		Name: name,
	}
	g.Players = append(g.Players, p)
	return p
}

func (g *Game) ProcessCommand(method string, params map[string]any) error {
	switch method {
	case "next_state":
		fromState := common.OptStringParam(params, "from_state")
		return g.NextState(fromState)
	case "button_click":
		pid, e := common.IntParam(params, "player_id")
		if e != nil {
			return e
		}
		return g.ButtonClick(pid)
	case "set_answerer":
		pid, e := common.IntParam(params, "player_id")
		if e != nil {
			return e
		}
		return g.SetAnswerer(pid)
	case "answer":
		isCorrect, e := common.BoolParam(params, "is_correct")
		if e != nil {
			return e
		}
		answerID := common.OptIntParam(params, "answer_id")
		return g.AnswerQuestion(isCorrect, answerID)
	default:
		return &common.BadFormatError{Msg: "Unknown method"}
	}
}

// --- Parse ---

func (g *Game) Parse(data []byte) error {
	var root common.XMLElement
	if err := xml.Unmarshal(data, &root); err != nil {
		return &common.BadFormatError{Msg: "Cannot parse XML"}
	}

	questionsXML := root.Find("questions")
	if questionsXML == nil {
		return &common.BadFormatError{Msg: "No questions found"}
	}

	roundQs := questionsXML.FindAll("question")
	if len(roundQs) == 0 {
		return &common.BadFormatError{Msg: "Game should have at least 1 round"}
	}

	for _, qXML := range roundQs {
		q := &Question{
			ID:      common.NextID(),
			Text:    qXML.FindText("text"),
			IsFinal: false,
		}
		for _, aXML := range qXML.FindAll("answer") {
			val := 0
			for _, c := range aXML.FindText("value") {
				if c >= '0' && c <= '9' {
					val = val*10 + int(c-'0')
				}
			}
			q.Answers = append(q.Answers, &Answer{
				ID:    common.NextID(),
				Text:  aXML.FindText("text"),
				Value: val,
			})
		}
		g.Questions = append(g.Questions, q)
	}

	finalQuestionsXML := root.Find("final_questions")
	if finalQuestionsXML == nil {
		return &common.BadFormatError{Msg: "No final questions found"}
	}

	finalQs := finalQuestionsXML.FindAll("question")
	if len(finalQs) != 5 {
		return &common.BadFormatError{Msg: "Game should have exactly 5 final questions"}
	}

	for _, qXML := range finalQs {
		q := &Question{
			ID:      common.NextID(),
			Text:    qXML.FindText("text"),
			IsFinal: true,
		}
		for _, aXML := range qXML.FindAll("answer") {
			val := 0
			for _, c := range aXML.FindText("value") {
				if c >= '0' && c <= '9' {
					val = val*10 + int(c-'0')
				}
			}
			q.Answers = append(q.Answers, &Answer{
				ID:    common.NextID(),
				Text:  aXML.FindText("text"),
				Value: val,
			})
		}
		g.Questions = append(g.Questions, q)
	}

	return nil
}

// --- Serialization ---

type GameState struct {
	Token          string      `json:"token"`
	Expired        time.Time   `json:"expired"`
	Round          int         `json:"round"`
	State          string      `json:"state"`
	Question       *Question   `json:"question"`
	Answerer       *int        `json:"answerer"`
	FinalQuestions []*Question `json:"final_questions"`
	Timer          int64       `json:"timer"`
	Players        []*Player   `json:"players"`
	Name           string      `json:"name"`
}

func (g *Game) Serialize() GameState {
	var answererID *int
	if g.Answerer != nil {
		answererID = &g.Answerer.ID
	}

	var finalQuestions []*Question
	if g.State == StateFinalQuestionsReveal {
		for _, q := range g.Questions {
			if q.IsFinal {
				finalQuestions = append(finalQuestions, q)
			}
		}
	}

	return GameState{
		Token:          g.Token,
		Expired:        g.Expired,
		Round:          g.Round,
		State:          g.State,
		Question:       g.Question,
		Answerer:       answererID,
		FinalQuestions: finalQuestions,
		Timer:          g.Timer,
		Players:        g.Players,
		Name:           "feud",
	}
}
