package jeopardy

import (
	"encoding/xml"
	"strings"
	"sync"
	"time"

	"bunjgames/common"
)

const (
	StateWaitingForPlayers = "waiting_for_players"
	StateIntro             = "intro"
	StateThemesAll         = "themes_all"
	StateRound             = "round"
	StateRoundThemes       = "round_themes"
	StateQuestions         = "questions"
	StateQuestionEvent     = "question_event"
	StateQuestion          = "question"
	StateAnswer            = "answer"
	StateQuestionEnd       = "question_end"
	StateFinalThemes       = "final_themes"
	StateFinalBets         = "final_bets"
	StateFinalQuestion     = "final_question"
	StateFinalAnswer       = "final_answer"
	StateFinalPlayerAnswer = "final_player_answer"
	StateFinalPlayerBet    = "final_player_bet"
	StateGameEnd           = "game_end"

	TypeStandard = "standard"
	TypeAuction  = "auction"
	TypeBagCat   = "bagcat"
)

type Question struct {
	ID          int     `json:"id"`
	CustomTheme *string `json:"custom_theme"`
	Text        *string `json:"text"`
	Image       *string `json:"image"`
	Audio       *string `json:"audio"`
	Video       *string `json:"video"`
	Answer      string  `json:"answer"`
	AnswerText  *string `json:"answer_text"`
	AnswerImage *string `json:"answer_image"`
	AnswerAudio *string `json:"answer_audio"`
	AnswerVideo *string `json:"answer_video"`
	Value       int     `json:"value"`
	Comment     string  `json:"comment"`
	Type        string  `json:"type"`
	IsProcessed bool    `json:"is_processed"`
}

type Theme struct {
	ID        int         `json:"id"`
	Name      string      `json:"name"`
	Comment   *string     `json:"comment"`
	Round     int         `json:"-"`
	IsRemoved bool        `json:"is_removed"`
	Questions []*Question `json:"questions"`
}

type Player struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Balance     int    `json:"balance"`
	FinalBet    int    `json:"final_bet"`
	FinalAnswer string `json:"final_answer"`
}

type Game struct {
	Mu sync.Mutex `json:"-"`

	common.IntercomQueue

	Token       string    `json:"token"`
	Expired     time.Time `json:"expired"`
	LastRound   int       `json:"-"`
	FinalRound  int       `json:"-"`
	State       string    `json:"state"`
	Round       int       `json:"round"`
	Question    *Question `json:"question"`
	Answerer    *Player   `json:"-"`
	QuestionBet int       `json:"-"`
	Themes      []*Theme  `json:"-"`
	Players     []*Player `json:"players"`
}

func NewGame() *Game {
	return &Game{
		State:   StateWaitingForPlayers,
		Expired: time.Now().Add(12 * time.Hour),
		Round:   1,
	}
}

func (g *Game) isFinalRound() bool {
	return g.FinalRound != 0 && g.Round == g.FinalRound
}

func (g *Game) GetThemes() []*Theme {
	if g.State == StateWaitingForPlayers || g.State == StateGameEnd {
		return nil
	}
	if g.State == StateThemesAll {
		maxRound := g.LastRound
		if g.FinalRound != 0 {
			maxRound = g.LastRound - 1
		}
		var result []*Theme
		for _, t := range g.Themes {
			if t.Round >= 1 && t.Round <= maxRound {
				result = append(result, t)
			}
		}
		return result
	}
	var result []*Theme
	for _, t := range g.Themes {
		if t.Round == g.Round {
			result = append(result, t)
		}
	}
	return result
}

func (g *Game) findQuestion(id int) *Question {
	for _, t := range g.Themes {
		for _, q := range t.Questions {
			if q.ID == id {
				return q
			}
		}
	}
	return nil
}

func (g *Game) findQuestionTheme(q *Question) *Theme {
	for _, t := range g.Themes {
		for _, tq := range t.Questions {
			if tq.ID == q.ID {
				return t
			}
		}
	}
	return nil
}

func (g *Game) findPlayer(id int) *Player {
	for _, p := range g.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (g *Game) hasUnprocessedQuestions() bool {
	themes := g.GetThemes()
	for _, t := range themes {
		for _, q := range t.Questions {
			if !q.IsProcessed {
				return true
			}
		}
	}
	return false
}

func (g *Game) processQuestionEnd() {
	g.Question.IsProcessed = true

	if g.hasUnprocessedQuestions() {
		g.Question = nil
		g.State = StateQuestions
	} else {
		g.Question = nil
		g.State = StateRound
		g.Round++
		if len(g.GetThemes()) == 0 {
			g.State = StateGameEnd
		}
	}
}

func (g *Game) nextFinalEndState() {
	for _, p := range g.Players {
		if p.FinalBet > 0 {
			g.Answerer = p
			g.State = StateFinalPlayerAnswer
			return
		}
	}
	g.State = StateGameEnd
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
		g.State = StateThemesAll
	case StateThemesAll:
		g.State = StateRound
	case StateRound:
		if g.isFinalRound() {
			g.State = StateFinalThemes
			nonRemoved := g.getNonRemovedThemes()
			if len(nonRemoved) == 1 {
				g.State = StateFinalBets
				g.Question = nonRemoved[0].Questions[0]
			}
		} else {
			g.State = StateRoundThemes
		}
	case StateRoundThemes:
		g.State = StateQuestions
	case StateQuestions:
		return common.ErrNothingToDo
	case StateQuestionEvent:
		return common.ErrNothingToDo
	case StateQuestion:
		g.State = StateAnswer
	case StateAnswer:
		return common.ErrNothingToDo
	case StateQuestionEnd:
		g.processQuestionEnd()
	case StateFinalThemes:
		return common.ErrNothingToDo
	case StateFinalBets:
		if g.isFinalRound() {
			for _, p := range g.Players {
				if p.Balance > 0 && p.FinalBet <= 0 {
					return &common.BadStateError{Msg: "Wait for all bets"}
				}
			}
		}
		g.State = StateFinalQuestion
	case StateFinalQuestion:
		g.State = StateFinalAnswer
	case StateFinalAnswer:
		g.nextFinalEndState()
	case StateFinalPlayerAnswer:
		return common.ErrNothingToDo
	case StateFinalPlayerBet:
		answerer := g.Answerer
		answerer.FinalBet = 0
		g.Answerer = nil
		g.nextFinalEndState()
	default:
		return &common.BadStateError{Msg: "Bad state"}
	}
	return nil
}

func (g *Game) getNonRemovedThemes() []*Theme {
	var result []*Theme
	for _, t := range g.GetThemes() {
		if !t.IsRemoved {
			result = append(result, t)
		}
	}
	return result
}

func (g *Game) ChooseQuestion(questionID int) error {
	q := g.findQuestion(questionID)
	if q == nil || q.IsProcessed {
		return &common.BadStateError{Msg: "Question is already processed"}
	}
	theme := g.findQuestionTheme(q)
	if theme == nil {
		return &common.BadStateError{Msg: "Question not found"}
	}
	inCurrentThemes := false
	for _, t := range g.GetThemes() {
		if t.ID == theme.ID {
			inCurrentThemes = true
			break
		}
	}
	if !inCurrentThemes {
		return &common.BadStateError{Msg: "Question is already processed"}
	}

	if q.Type == TypeStandard {
		g.State = StateQuestion
	} else {
		g.State = StateQuestionEvent
	}
	g.Question = q
	return nil
}

func (g *Game) SetAnswererAndBet(playerID int, bet int) error {
	if g.State != StateQuestionEvent {
		return common.ErrNothingToDo
	}
	if bet <= 0 {
		return &common.BadStateError{Msg: "Bet must be more than 0"}
	}
	player := g.findPlayer(playerID)
	if player == nil {
		return &common.BadStateError{Msg: "Player not found"}
	}
	g.QuestionBet = bet
	g.Answerer = player
	g.State = StateQuestion
	return nil
}

func (g *Game) SkipQuestion() error {
	if g.State != StateQuestionEvent && g.State != StateQuestion && g.State != StateAnswer {
		return common.ErrNothingToDo
	}
	g.QuestionBet = 0
	g.Answerer = nil

	if g.Question.AnswerText != nil || g.Question.AnswerImage != nil ||
		g.Question.AnswerAudio != nil || g.Question.AnswerVideo != nil {
		g.State = StateQuestionEnd
	} else {
		g.processQuestionEnd()
	}
	g.QueueIntercom("skip")
	return nil
}

func (g *Game) ButtonClick(playerID int) error {
	if g.State != StateAnswer || g.Answerer != nil {
		return common.ErrNothingToDo
	}
	if g.Question.Type != TypeStandard {
		return common.ErrNothingToDo
	}
	player := g.findPlayer(playerID)
	if player == nil {
		return &common.BadStateError{Msg: "Player not found"}
	}
	g.Answerer = player
	return nil
}

func (g *Game) AnswerQuestion(isRight bool) error {
	if g.State != StateAnswer || g.Answerer == nil {
		return common.ErrNothingToDo
	}

	questionEnd := func() {
		g.QuestionBet = 0
		if g.Question.AnswerText != nil || g.Question.AnswerImage != nil ||
			g.Question.AnswerAudio != nil || g.Question.AnswerVideo != nil {
			g.State = StateQuestionEnd
		} else {
			g.processQuestionEnd()
		}
	}

	if isRight {
		player := g.Answerer
		if g.Question.Type != TypeStandard {
			player.Balance += g.QuestionBet
		} else {
			player.Balance += g.Question.Value
		}
		questionEnd()
	} else {
		player := g.Answerer
		if g.Question.Type != TypeStandard {
			player.Balance -= g.QuestionBet
		} else {
			player.Balance -= g.Question.Value
		}
		if g.Question.Type != TypeStandard {
			questionEnd()
		}
	}
	g.Answerer = nil
	return nil
}

func (g *Game) RemoveFinalTheme(themeID int) error {
	if g.State != StateFinalThemes {
		return common.ErrNothingToDo
	}
	var theme *Theme
	for _, t := range g.GetThemes() {
		if t.ID == themeID {
			theme = t
			break
		}
	}
	if theme == nil {
		return &common.BadStateError{Msg: "Theme not found"}
	}
	theme.IsRemoved = true

	nonRemoved := g.getNonRemovedThemes()
	if len(nonRemoved) == 1 {
		g.State = StateFinalBets
		g.Question = nonRemoved[0].Questions[0]
	}
	return nil
}

func (g *Game) FinalBet(playerID int, bet int) error {
	if g.State != StateFinalBets {
		return common.ErrNothingToDo
	}
	player := g.findPlayer(playerID)
	if player == nil {
		return &common.BadStateError{Msg: "Player not found"}
	}
	if bet <= 0 {
		return &common.BadStateError{Msg: "Bet must be more than 0"}
	}
	if player.Balance < bet {
		return &common.BadStateError{Msg: "Not enough money"}
	}
	player.FinalBet = bet
	return nil
}

func (g *Game) FinalAnswerPlayer(playerID int, answer string) error {
	if g.State != StateFinalAnswer {
		return common.ErrNothingToDo
	}
	if answer == "" {
		return &common.BadStateError{Msg: "Answer cannot be empty"}
	}
	player := g.findPlayer(playerID)
	if player == nil {
		return &common.BadStateError{Msg: "Player not found"}
	}
	player.FinalAnswer = answer
	return nil
}

func (g *Game) FinalPlayerAnswer(isRight bool) error {
	if g.State != StateFinalPlayerAnswer {
		return common.ErrNothingToDo
	}
	answerer := g.Answerer
	if isRight {
		answerer.Balance += answerer.FinalBet
	} else {
		answerer.Balance -= answerer.FinalBet
	}
	g.State = StateFinalPlayerBet
	return nil
}

func (g *Game) SetBalance(balanceList []int) {
	for i, p := range g.Players {
		if i < len(balanceList) {
			p.Balance = balanceList[i]
		}
	}
}

func (g *Game) SetRound(round int) {
	g.Round = round
	themes := g.GetThemes()
	if len(themes) > 0 {
		g.State = StateRound
	} else {
		g.State = StateGameEnd
	}
	g.Answerer = nil
	g.QuestionBet = 0
	for _, t := range themes {
		for _, q := range t.Questions {
			if q.IsProcessed {
				q.IsProcessed = false
			}
		}
		if t.IsRemoved {
			t.IsRemoved = false
		}
	}
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

// --- Parse ---

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (g *Game) ProcessCommand(method string, params map[string]any) error {
	switch method {
	case "next_state":
		fromState := common.OptStringParam(params, "from_state")
		return g.NextState(fromState)
	case "choose_question":
		qid, e := common.IntParam(params, "question_id")
		if e != nil {
			return e
		}
		return g.ChooseQuestion(qid)
	case "set_answerer_and_bet":
		pid, e1 := common.IntParam(params, "player_id")
		bet, e2 := common.IntParam(params, "bet")
		if e1 != nil || e2 != nil {
			return &common.BadFormatError{Msg: "Invalid params"}
		}
		return g.SetAnswererAndBet(pid, bet)
	case "skip_question":
		return g.SkipQuestion()
	case "button_click":
		pid, e := common.IntParam(params, "player_id")
		if e != nil {
			return e
		}
		return g.ButtonClick(pid)
	case "answer":
		isRight, e := common.BoolParam(params, "is_right")
		if e != nil {
			return e
		}
		return g.AnswerQuestion(isRight)
	case "remove_final_theme":
		tid, e := common.IntParam(params, "theme_id")
		if e != nil {
			return e
		}
		return g.RemoveFinalTheme(tid)
	case "final_bet":
		pid, e1 := common.IntParam(params, "player_id")
		bet, e2 := common.IntParam(params, "bet")
		if e1 != nil || e2 != nil {
			return &common.BadFormatError{Msg: "Invalid params"}
		}
		return g.FinalBet(pid, bet)
	case "final_answer":
		pid, e1 := common.IntParam(params, "player_id")
		answer, e2 := common.StringParam(params, "answer")
		if e1 != nil || e2 != nil {
			return &common.BadFormatError{Msg: "Invalid params"}
		}
		return g.FinalAnswerPlayer(pid, answer)
	case "final_player_answer":
		isRight, e := common.BoolParam(params, "is_right")
		if e != nil {
			return e
		}
		return g.FinalPlayerAnswer(isRight)
	case "set_balance":
		bl, e := common.IntSliceParam(params, "balance_list")
		if e != nil {
			return e
		}
		g.SetBalance(bl)
		return nil
	case "set_round":
		round, e := common.IntParam(params, "round")
		if e != nil {
			return e
		}
		g.SetRound(round)
		return nil
	default:
		return &common.BadFormatError{Msg: "Unknown method"}
	}
}

func (g *Game) Parse(data []byte) error {
	var root common.XMLElement
	if err := xml.Unmarshal(data, &root); err != nil {
		return &common.BadFormatError{Msg: "Cannot parse XML"}
	}

	roundsXML := root.Find("rounds")
	if roundsXML == nil {
		return &common.BadFormatError{Msg: "No rounds found"}
	}

	lastRound := 0

	formatImageURL := func(url string) *string {
		if url == "" {
			return nil
		}
		url = strings.TrimPrefix(url, "@")
		return strPtr("/Images/" + url)
	}

	formatAudioURL := func(url string) *string {
		if url == "" {
			return nil
		}
		url = strings.TrimPrefix(url, "@")
		return strPtr("/Audio/" + url)
	}

	formatVideoURL := func(url string) *string {
		if url == "" {
			return nil
		}
		url = strings.TrimPrefix(url, "@")
		return strPtr("/Video/" + url)
	}

	for i, roundXML := range roundsXML.FindAll("round") {
		lastRound = i + 1
		themesXML := roundXML.Find("themes")
		if themesXML == nil {
			continue
		}
		for _, themeXML := range themesXML.FindAll("theme") {
			themeName := themeXML.Attr("name")

			var themeComment *string
			infoXML := themeXML.Find("info")
			if infoXML != nil {
				commentsXML := infoXML.Find("comments")
				if commentsXML != nil {
					themeComment = strPtr(commentsXML.Text())
				}
			}

			questionsXML := themeXML.Find("questions")
			if questionsXML == nil {
				continue
			}

			theme := &Theme{
				ID:      common.NextID(),
				Name:    themeName,
				Comment: themeComment,
				Round:   i + 1,
			}

			questionXMLs := questionsXML.FindAll("question")
			if len(questionXMLs) > 8 {
				questionXMLs = questionXMLs[:8]
			}

			for _, qXML := range questionXMLs {
				price := 0
				if p := qXML.Attr("price"); p != "" {
					for _, c := range p {
						if c >= '0' && c <= '9' {
							price = price*10 + int(c-'0')
						}
					}
				}

				qType := TypeStandard
				var customTheme *string

				typeXML := qXML.Find("type")
				if typeXML != nil {
					typeName := typeXML.Attr("name")
					if typeName == "auction" {
						qType = TypeAuction
					} else if typeName == "cat" || typeName == "bagcat" {
						qType = TypeBagCat
						for _, param := range typeXML.FindAll("param") {
							if param.Attr("name") == "theme" {
								customTheme = strPtr(param.Text())
							}
						}
					}
				}
				if customTheme == nil || *customTheme == "" {
					customTheme = &themeName
				}

				var text string
				var image, audio, video *string

				markerFlag := false
				var postText, postImage, postAudio, postVideo *string

				var atoms []*common.XMLElement
				scenarioXML := qXML.Find("scenario")
				if scenarioXML != nil {
					atoms = scenarioXML.FindAll("atom")
				} else {
					paramsXML := qXML.Find("params")
					if paramsXML != nil {
						paramXML := paramsXML.Find("param")
						if paramXML != nil {
							atoms = paramXML.FindAll("item")
						}
					}
				}

				for _, atom := range atoms {
					atomType := atom.Attr("type")
					atomText := atom.Text()
					switch atomType {
					case "image":
						if markerFlag {
							postImage = &atomText
						} else {
							image = &atomText
						}
					case "voice", "audio":
						if markerFlag {
							postAudio = &atomText
						} else {
							audio = &atomText
						}
					case "video":
						if markerFlag {
							postVideo = &atomText
						} else {
							video = &atomText
						}
					case "marker":
						markerFlag = true
					default:
						if atomText != "" {
							if markerFlag {
								postText = &atomText
							} else {
								text = atomText
							}
						}
					}
				}

				var rightAnswer string
				rightXML := qXML.Find("right")
				if rightXML != nil {
					for _, ansXML := range rightXML.FindAll("answer") {
						if ansXML.Text() != "" {
							rightAnswer += ansXML.Text() + "   "
						}
					}
					rightAnswer = strings.TrimSpace(rightAnswer)
				}

				var comment string
				infoXML := qXML.Find("info")
				if infoXML != nil {
					commentsXML := infoXML.Find("comments")
					if commentsXML != nil {
						comment = commentsXML.Text()
					}
				}

				q := &Question{
					ID:          common.NextID(),
					CustomTheme: customTheme,
					Text:        strPtr(text),
					Image:       formatImageURL(ptrVal(image)),
					Audio:       formatAudioURL(ptrVal(audio)),
					Video:       formatVideoURL(ptrVal(video)),
					AnswerText:  postText,
					AnswerImage: formatImageURL(ptrVal(postImage)),
					AnswerAudio: formatAudioURL(ptrVal(postAudio)),
					AnswerVideo: formatVideoURL(ptrVal(postVideo)),
					Value:       price,
					Answer:      rightAnswer,
					Comment:     comment,
					Type:        qType,
				}
				theme.Questions = append(theme.Questions, q)
			}
			g.Themes = append(g.Themes, theme)
		}
	}

	g.LastRound = lastRound

	for round := 1; round <= lastRound; round++ {
		maxQuestions := 0
		for _, t := range g.Themes {
			if t.Round == round && len(t.Questions) > maxQuestions {
				maxQuestions = len(t.Questions)
			}
		}
		for _, t := range g.Themes {
			if t.Round == round {
				for len(t.Questions) < maxQuestions {
					t.Questions = append(t.Questions, &Question{
						ID:          common.NextID(),
						Answer:      "-",
						Value:       0,
						Comment:     "-",
						Type:        TypeStandard,
						IsProcessed: true,
					})
				}
			}
		}
	}

	lastRoundThemes := g.themesForRound(lastRound)
	if len(lastRoundThemes) > 0 && len(lastRoundThemes[0].Questions) == 1 {
		g.FinalRound = lastRound
	}

	return nil
}

func ptrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (g *Game) themesForRound(round int) []*Theme {
	var result []*Theme
	for _, t := range g.Themes {
		if t.Round == round {
			result = append(result, t)
		}
	}
	return result
}

// --- Serialization ---

type GameState struct {
	Token        string    `json:"token"`
	Expired      time.Time `json:"expired"`
	Round        int       `json:"round"`
	RoundsCount  int       `json:"rounds_count"`
	IsFinalRound bool      `json:"is_final_round"`
	State        string    `json:"state"`
	Question     *Question `json:"question"`
	Themes       []*Theme  `json:"themes"`
	Players      []*Player `json:"players"`
	Answerer     *int      `json:"answerer"`
	Name         string    `json:"name"`
}

func (g *Game) Serialize() GameState {
	var answererID *int
	if g.Answerer != nil {
		answererID = &g.Answerer.ID
	}

	return GameState{
		Token:        g.Token,
		Expired:      g.Expired,
		Round:        g.Round,
		RoundsCount:  g.LastRound,
		IsFinalRound: g.isFinalRound(),
		State:        g.State,
		Question:     g.Question,
		Themes:       g.GetThemes(),
		Players:      g.Players,
		Answerer:     answererID,
		Name:         "jeopardy",
	}
}
