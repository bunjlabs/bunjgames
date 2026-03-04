package weakest

import (
	"math"
	"sort"
	"sync"
	"time"

	"bunjgames-server/common"
)

const (
	StateWaitingForPlayers = "waiting_for_players"
	StateIntro             = "intro"
	StateRound             = "round"
	StateQuestions         = "questions"
	StateWeakestChoose     = "weakest_choose"
	StateWeakestReveal     = "weakest_reveal"
	StateFinal             = "final"
	StateFinalQuestions    = "final_questions"
	StateEnd               = "end"
)

type Question struct {
	ID           int `json:"-"`
	QuestionText string
	AnswerText   string
	IsFinal      bool
	IsProcessed  bool
	IsCorrect    *bool
}

type Player struct {
	ID           int
	Name         string
	IsWeak       bool
	WeakID       *int
	RightAnswers int
	BankIncome   int
}

type Game struct {
	Mu sync.Mutex

	Token           string
	Expired         time.Time
	ScoreMultiplier int
	Score           int
	Bank            int
	TmpScore        int
	Round           int
	State           string
	Questions       []*Question
	Players         []*Player
	Question        *Question
	Answerer        *Player
	Weakest         *Player
	Strongest       *Player
	Timer           int64
	BankTimer       int64
}

func NewGame() *Game {
	return &Game{
		State:           StateWaitingForPlayers,
		Expired:         time.Now().Add(12 * time.Hour),
		Round:           1,
		ScoreMultiplier: 1,
	}
}

func (g *Game) getActivePlayers() []*Player {
	var result []*Player
	for _, p := range g.Players {
		if !p.IsWeak {
			result = append(result, p)
		}
	}
	return result
}

func (g *Game) getQuestions() []*Question {
	isFinal := g.State == StateFinalQuestions
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

func (g *Game) setTimer(seconds int) {
	g.Timer = time.Now().Add(time.Duration(seconds) * time.Second).UnixMilli()
}

func (g *Game) clearTimer() {
	g.Timer = 0
}

func (g *Game) setBankTimer(seconds int) {
	g.BankTimer = time.Now().Add(time.Duration(seconds) * time.Second).UnixMilli()
}

func (g *Game) getNextTmpScore() int {
	scores := []int{0, 1, 2, 5, 10, 15, 20, 30, 40}
	for i, s := range scores {
		if s == g.TmpScore && i+1 < len(scores) {
			return scores[i+1]
		}
	}
	return g.TmpScore
}

func (g *Game) getWeakest() *Player {
	if g.State == StateWeakestReveal {
		return g.getWeakestByVotes()
	}
	players := g.getActivePlayers()
	if len(players) == 0 {
		return nil
	}
	sort.Slice(players, func(i, j int) bool {
		if players[i].RightAnswers != players[j].RightAnswers {
			return players[i].RightAnswers < players[j].RightAnswers
		}
		return players[i].BankIncome < players[j].BankIncome
	})
	return players[0]
}

func (g *Game) getWeakestByVotes() *Player {
	players := g.getActivePlayers()
	voteCount := make(map[int]int)
	for _, p := range players {
		if p.WeakID != nil {
			voteCount[*p.WeakID]++
		}
	}

	var candidates []*Player
	for _, p := range players {
		if voteCount[p.ID] > 0 {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		ci, cj := voteCount[candidates[i].ID], voteCount[candidates[j].ID]
		if ci != cj {
			return ci > cj
		}
		if candidates[i].RightAnswers != candidates[j].RightAnswers {
			return candidates[i].RightAnswers < candidates[j].RightAnswers
		}
		return candidates[i].BankIncome < candidates[j].BankIncome
	})
	return candidates[0]
}

func (g *Game) getStrongest() *Player {
	players := g.getActivePlayers()
	if len(players) == 0 {
		return nil
	}
	sort.Slice(players, func(i, j int) bool {
		if players[i].RightAnswers != players[j].RightAnswers {
			return players[i].RightAnswers > players[j].RightAnswers
		}
		return players[i].BankIncome > players[j].BankIncome
	})
	return players[0]
}

func (g *Game) nextQuestion(answerer *Player, isCorrect *bool) {
	if g.Question != nil {
		g.Question.IsProcessed = true
		if isCorrect != nil {
			g.Question.IsCorrect = isCorrect
		}
	}

	questions := g.getQuestions()
	if len(questions) > 0 {
		g.Question = questions[0]
	} else {
		g.Question = nil
	}

	if g.Question == nil && (g.State == StateFinal || g.State == StateFinalQuestions) {
		g.Answerer = nil
		g.State = StateEnd
	} else if g.Question == nil {
		g.SaveBank(true)
		g.roundEnd()
	} else {
		if answerer != nil {
			g.Answerer = answerer
		} else if g.Answerer != nil {
			players := g.getActivePlayers()
			for i, p := range players {
				if p.ID == g.Answerer.ID {
					g.Answerer = players[(i+1)%len(players)]
					break
				}
			}
		} else if g.Strongest != nil && !g.Strongest.IsWeak {
			g.Answerer = g.Strongest
		} else {
			players := g.getActivePlayers()
			sort.Slice(players, func(i, j int) bool {
				return players[i].Name < players[j].Name
			})
			if len(players) > 0 {
				g.Answerer = players[0]
			}
		}
		g.setBankTimer(3)
	}
}

func (g *Game) roundEnd() {
	g.clearTimer()
	g.Score += g.Bank
	g.TmpScore = 0
	g.Answerer = nil
	g.Weakest = g.getWeakest()
	g.Strongest = g.getStrongest()
	g.State = StateWeakestChoose
}

func (g *Game) nextRound() {
	g.Round++
	g.Bank = 0
}

func (g *Game) NextState(fromState *string) error {
	if fromState != nil && g.State != *fromState {
		return common.ErrNothingToDo
	}
	switch g.State {
	case StateWaitingForPlayers:
		if len(g.Players) >= 3 {
			g.State = StateIntro
		} else {
			return &common.BadStateError{Msg: "Not enough players"}
		}
	case StateIntro:
		g.State = StateRound
	case StateRound:
		players := g.getActivePlayers()
		for _, p := range players {
			p.RightAnswers = 0
			p.BankIncome = 0
		}
		g.State = StateQuestions
		timerSec := 150 - (g.Round-1)*10
		g.setTimer(timerSec)
		g.nextQuestion(nil, nil)
	case StateQuestions:
		g.roundEnd()
	case StateWeakestChoose:
		return common.ErrNothingToDo
	case StateWeakestReveal:
		weakest := g.Weakest
		weakest.IsWeak = true
		g.Weakest = nil
		for _, p := range g.Players {
			if !p.IsWeak {
				p.WeakID = nil
			}
		}
		activePlayers := g.getActivePlayers()
		if len(activePlayers) > 2 {
			g.State = StateRound
			g.nextRound()
		} else {
			g.State = StateFinal
			g.nextRound()
		}
	case StateFinal:
		return common.ErrNothingToDo
	case StateFinalQuestions:
		return common.ErrNothingToDo
	case StateEnd:
		return common.ErrNothingToDo
	default:
		return &common.BadStateError{Msg: "Bad state"}
	}
	return nil
}

func (g *Game) SaveBank(force bool) error {
	if g.State != StateQuestions {
		return common.ErrNothingToDo
	}
	if !force && g.BankTimer < time.Now().UnixMilli() {
		return common.ErrNothingToDo
	}
	player := g.Answerer
	if player != nil {
		income := g.TmpScore
		if g.Bank+g.TmpScore > 40 {
			income = 40 - g.Bank
		}
		player.BankIncome += income
	}
	g.Bank += g.TmpScore
	g.TmpScore = 0
	if g.Bank >= 40 {
		g.Bank = 40
		g.roundEnd()
	}
	return nil
}

func (g *Game) AnswerCorrect(isCorrect bool) error {
	if g.State != StateQuestions && g.State != StateFinalQuestions {
		return common.ErrNothingToDo
	}

	if isCorrect {
		player := g.Answerer
		if player != nil {
			player.RightAnswers++
		}
		if g.State == StateQuestions {
			g.TmpScore = g.getNextTmpScore()
			if g.TmpScore == 40 {
				g.SaveBank(true)
				g.roundEnd()
				return nil
			}
		}
	} else if g.State == StateQuestions {
		g.TmpScore = 0
	}

	if g.State == StateFinalQuestions {
		players := g.getActivePlayers()
		if len(players) >= 2 {
			playerA := players[0]
			playerB := players[1]

			processedCount := 0
			for _, q := range g.Questions {
				if q.IsFinal && q.IsProcessed {
					processedCount++
				}
			}
			diff := int(math.Abs(float64(playerA.RightAnswers - playerB.RightAnswers)))

			if (processedCount < 10 && diff >= 3) || (processedCount == 9 && diff > 0) {
				if playerA.RightAnswers > playerB.RightAnswers {
					g.Answerer = playerA
				} else {
					g.Answerer = playerB
				}
				g.State = StateEnd
				return nil
			} else if processedCount >= 10 && processedCount%2 == 1 {
				var lastProcessed *Question
				for _, q := range g.Questions {
					if q.IsFinal && q.IsProcessed {
						lastProcessed = q
					}
				}
				if lastProcessed != nil && lastProcessed.IsCorrect != nil {
					lastCorrect := *lastProcessed.IsCorrect
					if isCorrect && !lastCorrect {
						g.State = StateEnd
						return nil
					} else if !isCorrect && lastCorrect {
						if playerA.ID != g.Answerer.ID {
							g.Answerer = playerA
						} else {
							g.Answerer = playerB
						}
						g.State = StateEnd
						return nil
					}
				}
			}
		}
	}

	ic := isCorrect
	g.nextQuestion(nil, &ic)
	return nil
}

func (g *Game) SelectWeakest(playerID, weakestID int) error {
	if g.State != StateWeakestChoose {
		return common.ErrNothingToDo
	}
	player := g.findPlayer(playerID)
	if player == nil {
		return &common.BadStateError{Msg: "Player not found"}
	}
	weakPlayer := g.findPlayer(weakestID)
	if weakPlayer == nil {
		return &common.BadStateError{Msg: "Player not found"}
	}
	if player.IsWeak || weakPlayer.IsWeak {
		return &common.BadStateError{Msg: "Cannot vote for weak player"}
	}
	player.WeakID = &weakestID

	allVoted := true
	for _, p := range g.Players {
		if !p.IsWeak && p.WeakID == nil {
			allVoted = false
			break
		}
	}
	if allVoted {
		g.State = StateWeakestReveal
		g.Weakest = g.getWeakest()
	}
	return nil
}

func (g *Game) SelectFinalAnswerer(playerID int) error {
	if g.State != StateFinal {
		return common.ErrNothingToDo
	}
	players := g.getActivePlayers()
	for _, p := range players {
		p.RightAnswers = 0
		p.BankIncome = 0
	}
	answerer := g.findPlayer(playerID)
	if answerer == nil {
		return &common.BadStateError{Msg: "Player not found"}
	}
	g.State = StateFinalQuestions
	g.nextQuestion(answerer, nil)
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

// --- Parse ---

func (g *Game) Parse(filename string) error {
	root, err := common.ParseXMLFile(filename)
	if err != nil {
		return &common.BadFormatError{Msg: "Cannot parse XML"}
	}

	questionsXML := root.Find("questions")
	if questionsXML == nil {
		return &common.BadFormatError{Msg: "No questions found"}
	}

	for _, qXML := range questionsXML.FindAll("question") {
		g.Questions = append(g.Questions, &Question{
			ID:           common.NextID(),
			QuestionText: qXML.FindText("question"),
			AnswerText:   qXML.FindText("answer"),
			IsFinal:      false,
		})
	}

	finalQuestionsXML := root.Find("final_questions")
	if finalQuestionsXML == nil {
		return &common.BadFormatError{Msg: "No final questions found"}
	}

	finalQs := finalQuestionsXML.FindAll("question")
	if len(finalQs) < 10 {
		return &common.BadFormatError{Msg: "Number of final questions must be 10 or more"}
	}
	if len(finalQs)%2 != 0 {
		return &common.BadFormatError{Msg: "Number of final questions must be even"}
	}

	for _, qXML := range finalQs {
		g.Questions = append(g.Questions, &Question{
			ID:           common.NextID(),
			QuestionText: qXML.FindText("question"),
			AnswerText:   qXML.FindText("answer"),
			IsFinal:      true,
		})
	}

	scoreMultiplierXML := root.Find("score_multiplier")
	if scoreMultiplierXML != nil {
		val := 0
		for _, c := range scoreMultiplierXML.Text() {
			if c >= '0' && c <= '9' {
				val = val*10 + int(c-'0')
			}
		}
		if val > 0 {
			g.ScoreMultiplier = val
		}
	}

	return nil
}

// --- Serialization ---

type QuestionInfoState struct {
	IsCorrect   *bool `json:"is_correct"`
	IsProcessed bool  `json:"is_processed"`
}

type QuestionState struct {
	QuestionText string `json:"question"`
	AnswerText   string `json:"answer"`
}

type PlayerState struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	IsWeak       bool   `json:"is_weak"`
	Weak         *int   `json:"weak"`
	RightAnswers int    `json:"right_answers"`
	BankIncome   int    `json:"bank_income"`
}

type GameState struct {
	Token           string              `json:"token"`
	Expired         time.Time           `json:"expired"`
	ScoreMultiplier int                 `json:"score_multiplier"`
	Score           int                 `json:"score"`
	Bank            int                 `json:"bank"`
	TmpScore        int                 `json:"tmp_score"`
	State           string              `json:"state"`
	Round           int                 `json:"round"`
	Question        *QuestionState      `json:"question"`
	Answerer        *int                `json:"answerer"`
	Weakest         *int                `json:"weakest"`
	Strongest       *int                `json:"strongest"`
	FinalQuestions  []QuestionInfoState `json:"final_questions"`
	Timer           int64               `json:"timer"`
	Players         []PlayerState       `json:"players"`
	Name            string              `json:"name"`
}

func (g *Game) Serialize() GameState {
	var question *QuestionState
	if g.Question != nil {
		question = &QuestionState{
			QuestionText: g.Question.QuestionText,
			AnswerText:   g.Question.AnswerText,
		}
	}

	var answererID, weakestID, strongestID *int
	if g.Answerer != nil {
		answererID = &g.Answerer.ID
	}
	if g.Weakest != nil {
		weakestID = &g.Weakest.ID
	}
	if g.Strongest != nil {
		strongestID = &g.Strongest.ID
	}

	var finalQuestions []QuestionInfoState
	if g.State == StateFinalQuestions {
		for _, q := range g.Questions {
			if q.IsFinal {
				finalQuestions = append(finalQuestions, QuestionInfoState{
					IsCorrect:   q.IsCorrect,
					IsProcessed: q.IsProcessed,
				})
			}
		}
	}

	playerStates := make([]PlayerState, 0, len(g.Players))
	for _, p := range g.Players {
		playerStates = append(playerStates, PlayerState{
			ID:           p.ID,
			Name:         p.Name,
			IsWeak:       p.IsWeak,
			Weak:         p.WeakID,
			RightAnswers: p.RightAnswers,
			BankIncome:   p.BankIncome,
		})
	}

	return GameState{
		Token:           g.Token,
		Expired:         g.Expired,
		ScoreMultiplier: g.ScoreMultiplier,
		Score:           g.Score,
		Bank:            g.Bank,
		TmpScore:        g.TmpScore,
		State:           g.State,
		Round:           g.Round,
		Question:        question,
		Answerer:        answererID,
		Weakest:         weakestID,
		Strongest:       strongestID,
		FinalQuestions:  finalQuestions,
		Timer:           g.Timer,
		Players:         playerStates,
		Name:            "weakest",
	}
}
