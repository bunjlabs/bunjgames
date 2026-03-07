package weakest

import (
	"math"
	"sort"
	"sync"
	"time"

	"bunjgames-server/common"

	"gopkg.in/yaml.v3"
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
	ID          int    `json:"id" yaml:"-"`
	Question    string `json:"question" yaml:"question"`
	Answer      string `json:"answer" yaml:"answer"`
	IsProcessed bool   `json:"is_processed" yaml:"-"`
	IsCorrect   *bool  `json:"is_correct" yaml:"-"`
}

type Player struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	IsWeak       bool   `json:"is_weak"`
	WeakID       *int   `json:"weak_id"`
	RightAnswers int    `json:"right_answers"`
	BankIncome   int    `json:"bank_income"`
}

type Game struct {
	Mu sync.Mutex `json:"-" yaml:"-"`

	Token           string      `json:"token" yaml:"-"`
	Expired         time.Time   `json:"expired" yaml:"-"`
	ScoreMultiplier int         `json:"score_multiplier" yaml:"score_multiplier"`
	Score           int         `json:"score" yaml:"-"`
	Bank            int         `json:"bank" yaml:"-"`
	TmpScore        int         `json:"tmp_score" yaml:"-"`
	Round           int         `json:"round" yaml:"-"`
	State           string      `json:"state" yaml:"-"`
	Questions       []*Question `json:"-" yaml:"questions"`
	FinalQuestions  []*Question `json:"-" yaml:"final_questions"`
	Players         []*Player   `json:"players" yaml:"-"`
	Question        *Question   `json:"question" yaml:"-"`
	Answerer        *Player     `json:"-" yaml:"-"`
	Weakest         *Player     `json:"-" yaml:"-"`
	Strongest       *Player     `json:"-" yaml:"-"`
	Timer           int64       `json:"timer" yaml:"-"`
	BankTimer       int64       `json:"bank_timer" yaml:"-"`
	Name            string      `json:"name" yaml:"-"`

	// JSON-only serialization fields
	AnswererID         *int        `json:"answerer" yaml:"-"`
	WeakestID          *int        `json:"weakest" yaml:"-"`
	StrongestID        *int        `json:"strongest" yaml:"-"`
	FinalQuestionsInfo []*Question `json:"final_questions" yaml:"-"`
}

func NewGame() *Game {
	return &Game{
		State:           StateWaitingForPlayers,
		Expired:         time.Now().Add(12 * time.Hour),
		Round:           1,
		ScoreMultiplier: 1,
		Players:         []*Player{},
		Name:            "weakest",
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
	if isFinal {
		for _, q := range g.FinalQuestions {
			if !q.IsProcessed {
				result = append(result, q)
			}
		}
	} else {
		for _, q := range g.Questions {
			if !q.IsProcessed {
				result = append(result, q)
			}
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
			for _, q := range g.FinalQuestions {
				if q.IsProcessed {
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
				for _, q := range g.FinalQuestions {
					if q.IsProcessed {
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

func (g *Game) ProcessCommand(method string, params map[string]any) error {
	switch method {
	case "next_state":
		fromState := common.OptStringParam(params, "from_state")
		return g.NextState(fromState)
	case "save_bank":
		return g.SaveBank(false)
	case "answer_correct":
		isCorrect, e := common.BoolParam(params, "is_correct")
		if e != nil {
			return e
		}
		return g.AnswerCorrect(isCorrect)
	case "select_weakest":
		pid, e1 := common.IntParam(params, "player_id")
		wid, e2 := common.IntParam(params, "weakest_id")
		if e1 != nil || e2 != nil {
			return &common.BadFormatError{Msg: "Invalid params"}
		}
		return g.SelectWeakest(pid, wid)
	case "select_final_answerer":
		pid, e := common.IntParam(params, "player_id")
		if e != nil {
			return e
		}
		return g.SelectFinalAnswerer(pid)
	default:
		return &common.BadFormatError{Msg: "Unknown method"}
	}
}

// --- Parse ---

func (g *Game) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, g); err != nil {
		return &common.BadFormatError{Msg: "Cannot parse YAML"}
	}

	if len(g.Questions) == 0 {
		return &common.BadFormatError{Msg: "No questions found"}
	}

	if len(g.FinalQuestions) < 10 {
		return &common.BadFormatError{Msg: "Number of final questions must be 10 or more"}
	}
	if len(g.FinalQuestions)%2 != 0 {
		return &common.BadFormatError{Msg: "Number of final questions must be even"}
	}

	for _, q := range g.Questions {
		q.ID = common.NextID()
	}

	for _, q := range g.FinalQuestions {
		q.ID = common.NextID()
	}

	if g.ScoreMultiplier <= 0 {
		g.ScoreMultiplier = 1
	}

	return nil
}

// --- Serialization ---

func (g *Game) Serialize() *Game {
	if g.Answerer != nil {
		g.AnswererID = &g.Answerer.ID
	} else {
		g.AnswererID = nil
	}
	if g.Weakest != nil {
		g.WeakestID = &g.Weakest.ID
	} else {
		g.WeakestID = nil
	}
	if g.Strongest != nil {
		g.StrongestID = &g.Strongest.ID
	} else {
		g.StrongestID = nil
	}

	if g.State == StateFinalQuestions {
		g.FinalQuestionsInfo = g.FinalQuestions
	} else {
		g.FinalQuestionsInfo = nil
	}

	return g
}
