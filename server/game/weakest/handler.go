package weakest

import (
	"bunjgames/game/abstract"
	"bunjgames/utils"
	"errors"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func NewGame() *Game {
	game := &Game{
		State:           State{Value: "waiting_for_players", Score: 0},
		RoundState:      RoundState{Number: 1, Score: 0, Bank: 0},
		ScoreMultiplier: 1,
		Questions:       []*Question{},
		FinalQuestions:  []*Question{},
		Players:         []*Player{},
	}
	game.Initialise()
	game.Type = "weakest"
	if os.Getenv("DEBUG") == "true" {
		game.Players = []*Player{
			{Name: "PLAYER1", Active: true, FinalScore: []bool{}},
			{Name: "PLAYER2", Active: true, FinalScore: []bool{}},
			{Name: "PLAYER3", Active: true, FinalScore: []bool{}},
		}
	}
	return game
}

func (game *Game) Parse(fileStream io.Reader) error {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	data, err := io.ReadAll(fileStream)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, game); err != nil {
		return err
	}

	if len(game.Questions) == 0 {
		return abstract.InvalidInputs
	}
	if len(game.FinalQuestions) < 10 {
		return abstract.InvalidInputs
	}

	if game.ScoreMultiplier <= 0 {
		game.ScoreMultiplier = 1
	}
	return nil
}

func (game *Game) Tick(delta time.Duration) (*abstract.Command, error) {
	game.Mutex.RLock()
	defer game.Mutex.RUnlock()

	if game.State.Value == "questions" {
		game.RoundState.Time -= delta.Milliseconds()
		game.RoundState.QuestionTime -= delta.Milliseconds()
		if game.RoundState.Time < 0 {
			return &abstract.Command{"game", game}, game.roundEnd()
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
	if game.State.Value != "waiting_for_players" {
		return errors.New("game is already in progress")
	}
	player := Player{
		Name:       name,
		Active:     true,
		FinalScore: []bool{},
	}
	game.Players = append(game.Players, &player)
	return nil
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
	case "bank":
		return gameCommand, game.bank(false)
	case "answer":
		correct, ok := params["correct"].(bool)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.answer(correct)
	case "vote":
		voter, ok1 := params["voter"].(string)
		weakest, ok2 := params["weakest"].(string)
		if !ok1 || !ok2 {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.vote(voter, weakest)
	case "final_answerer":
		answerer, ok := params["answerer"].(string)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.finalAnswerer(answerer)
	default:
		return nil, abstract.UnknownMethod
	}
}

func (game *Game) nextState(fromState string) error {
	if fromState != "" && game.State.Value != fromState {
		return abstract.NothingToDo
	}

	switch game.State.Value {
	case "waiting_for_players":
		if len(game.Players) < 3 {
			return abstract.NotEnoughPlayers
		}
		game.State.Value = "intro"
	case "intro":
		game.State.Value = "round"
	case "round":
		game.cleanup()
		game.State.Value = "questions"
		game.RoundState.Time = int64((150 - (game.RoundState.Number-1)*10) * 1000)
		return game.nextQuestion(nil)
	case "questions":
		return game.roundEnd()
	case "weakest_choose":
		weakest := game.getWeakestByVotes()
		if weakest == nil {
			return errors.New("could not determine weakest player")
		}
		game.RoundState.Kicked = weakest
		game.State.Value = "weakest_reveal"
	case "weakest_reveal":
		game.RoundState.Kicked.Active = false
		if len(game.activePlayers()) > 2 {
			game.State.Value = "round"
		} else {
			game.State.Value = "final"
		}
		game.RoundState.Number++
	case "final", "final_questions", "end":
		return abstract.NothingToDo
	default:
		return abstract.InvalidInputs
	}
	return nil
}

func (game *Game) bank(force bool) error {
	if game.State.Value != "questions" || (!force && game.RoundState.QuestionTime < 0) {
		return abstract.NothingToDo
	}
	if game.RoundState.Answerer != nil {
		game.RoundState.Answerer.BankIncome += game.RoundState.Score
	}
	game.RoundState.Bank += game.RoundState.Score
	game.RoundState.Score = 0
	if game.RoundState.Bank >= 40 && !force {
		game.State.Score += game.RoundState.Bank
		game.State.Value = "weakest_choose"
	}
	return nil
}

func (game *Game) nextQuestion(answerer *Player) error {
	if game.RoundState.Question != nil {
		game.RoundState.Question.Processed = true
	}

	questions := game.Questions
	if game.State.Value == "final_questions" {
		questions = game.FinalQuestions
	}
	for _, question := range questions {
		if question.Processed {
			continue
		}
		game.RoundState.Question = question
		break
	}

	if game.RoundState.Question.Processed {
		_ = game.bank(true)
		return game.roundEnd()
	}

	if answerer != nil {
		game.RoundState.Answerer = answerer
	} else if game.RoundState.Answerer != nil {
		players := game.activePlayers()
		for index, player := range players {
			if player == game.RoundState.Answerer {
				game.RoundState.Answerer = players[(index+1)%len(players)]
				break
			}
		}
	} else if game.RoundState.Strongest != nil && game.RoundState.Strongest.Active {
		game.RoundState.Answerer = game.RoundState.Strongest
	} else {
		game.RoundState.Answerer = game.activePlayers()[0]
	}
	game.RoundState.QuestionTime = 3 * 1000
	return nil
}

func (game *Game) vote(voterName string, weakestName string) error {
	if game.State.Value != "weakest_choose" {
		return abstract.NothingToDo
	}

	players := game.activePlayers()
	var voter *Player = nil
	var weakest *Player = nil
	for _, player := range players {
		if player.Name == voterName {
			voter = player
		}
		if player.Name == weakestName {
			weakest = player
		}
	}
	if voter == nil || weakest == nil {
		return abstract.InvalidInputs
	}

	voter.Vote = weakest
	voter.VoteName = weakest.Name

	allVoted := true
	for _, player := range players {
		if player.Vote == nil {
			allVoted = false
			break
		}
	}
	if allVoted {
		game.RoundState.Kicked = game.getWeakestByVotes()
		game.State.Value = "weakest_reveal"
	}
	return nil
}

func (game *Game) answer(correct bool) error {
	if game.State.Value != "questions" && game.State.Value != "final_questions" {
		return abstract.NothingToDo
	}

	if game.State.Value == "final_questions" {
		game.RoundState.Answerer.FinalScore = append(game.RoundState.Answerer.FinalScore, correct)

		sCorrect, sTotal := game.RoundState.Strongest.finalScoreCount()
		wCorrect, wTotal := game.RoundState.Weakest.finalScoreCount()
		if (max(sTotal, wTotal) <= 5 && (sCorrect > wCorrect+(5-wTotal) || wCorrect > sCorrect+(5-sTotal))) ||
			(max(sTotal, wTotal) > 5 && sTotal == wTotal && sCorrect != wCorrect) {
			return game.roundEnd()
		}
	}

	if correct && game.State.Value != "final_questions" {
		game.RoundState.Answerer.RightAnswers++
		game.bumpScore()
		if game.RoundState.Score >= 40 {
			_ = game.bank(true)
			game.RoundState.Question.Processed = true
			return game.roundEnd()
		}
	} else {
		game.RoundState.Score = 0
	}

	return game.nextQuestion(nil)
}

func (game *Game) finalAnswerer(answererName string) error {
	if game.State.Value != "final" {
		return abstract.NothingToDo
	}

	players := game.activePlayers()
	var answerer *Player = nil
	var opponent *Player = nil
	for _, player := range players {
		if player.Name == answererName {
			answerer = player
		} else {
			opponent = player
		}
	}
	if answerer == nil || opponent == nil {
		return abstract.InvalidInputs
	}

	game.cleanup()
	game.RoundState.Strongest = answerer
	game.RoundState.Weakest = opponent
	game.State.Value = "final_questions"
	return game.nextQuestion(answerer)
}

func (game *Game) roundEnd() error {
	if game.State.Value == "final_questions" {
		sCorrect, _ := game.RoundState.Strongest.finalScoreCount()
		wCorrect, _ := game.RoundState.Weakest.finalScoreCount()
		if sCorrect < wCorrect {
			game.RoundState.Strongest = game.RoundState.Weakest
		}
		game.RoundState.Weakest = nil
		game.RoundState.Answerer = nil
		game.State.Value = "end"
		game.RoundState.Question.Processed = true
		return nil
	}

	if game.State.Value != "questions" {
		return abstract.NothingToDo
	}

	players := game.activePlayers()
	var weakest *Player = nil
	var strongest *Player = nil
	for _, player := range players {
		if strongest == nil || player.RightAnswers > strongest.RightAnswers ||
			(player.RightAnswers == strongest.RightAnswers && player.BankIncome > strongest.BankIncome) {
			strongest = player
		}
		if weakest == nil || player.RightAnswers < weakest.RightAnswers ||
			(player.RightAnswers == weakest.RightAnswers && player.BankIncome < weakest.BankIncome) {
			weakest = player
		}
	}
	game.RoundState.Weakest = weakest
	game.RoundState.Strongest = strongest

	game.State.Score += game.RoundState.Bank
	game.RoundState.Score = 0
	game.RoundState.Bank = 0
	game.State.Value = "weakest_choose"
	return nil
}

func (game *Game) getWeakestByVotes() *Player {
	players := game.activePlayers()
	voteCount := make([]int, len(players))
	for _, player := range players {
		for voteIndex, vote := range players {
			if player.Vote == vote {
				voteCount[voteIndex]++
			}
		}
	}

	weakestIndex, maxVotes := utils.Max(voteCount)
	if weakestIndex < 0 || weakestIndex >= len(players) || maxVotes == 0 {
		return nil
	}
	return players[weakestIndex]
}

func (game *Game) cleanup() {
	for _, player := range game.Players {
		player.Vote = nil
		player.VoteName = ""
		player.RightAnswers = 0
		player.BankIncome = 0
	}

	game.RoundState.Score = 0
	game.RoundState.Bank = 0

	game.RoundState.Question = nil
	game.RoundState.Time = 0
	game.RoundState.QuestionTime = 0

	game.RoundState.Answerer = nil
	game.RoundState.Weakest = nil
	//game.RoundState.Strongest = nil
	game.RoundState.Kicked = nil
}

func (game *Game) activePlayers() []*Player {
	var result []*Player
	for _, p := range game.Players {
		if p.Active {
			result = append(result, p)
		}
	}
	return result
}

func (game *Game) bumpScore() {
	switch game.RoundState.Score {
	case 0:
		game.RoundState.Score = 1
	case 1:
		game.RoundState.Score = 2
	case 2:
		game.RoundState.Score = 5
	case 5:
		game.RoundState.Score = 10
	case 10:
		game.RoundState.Score = 15
	case 15:
		game.RoundState.Score = 20
	case 20:
		game.RoundState.Score = 30
	case 30:
		game.RoundState.Score = 40
	}
}

func (player *Player) finalScoreCount() (int, int) {
	correct := 0
	total := len(player.FinalScore)

	for _, isCorrect := range player.FinalScore {
		if isCorrect {
			correct++
		}
	}

	return correct, total
}
