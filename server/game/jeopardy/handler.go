package jeopardy

import (
	"bunjgames/game/abstract"
	"bunjgames/storage"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func NewGame() *Game {
	game := &Game{
		State:   State{Value: "waiting_for_players"},
		Players: []*Player{},
	}
	game.Initialise()
	game.Type = "jeopardy"
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

	var root storage.XMLElement
	if err := xml.Unmarshal(data, &root); err != nil {
		return abstract.InvalidInputs
	}

	roundsXML := root.Find("rounds")
	if roundsXML == nil {
		return abstract.InvalidInputs
	}

	formatImageURL := func(url string) *string {
		if url == "" {
			return nil
		}
		url = strings.TrimPrefix(url, "@")
		return strPtr("Images/" + url)
	}

	formatAudioURL := func(url string) *string {
		if url == "" {
			return nil
		}
		url = strings.TrimPrefix(url, "@")
		return strPtr("Audio/" + url)
	}

	formatVideoURL := func(url string) *string {
		if url == "" {
			return nil
		}
		url = strings.TrimPrefix(url, "@")
		return strPtr("Video/" + url)
	}

	for i, roundXML := range roundsXML.FindAll("round") {
		round := &Round{
			Number: i + 1,
		}

		themesXML := roundXML.Find("themes")
		if themesXML == nil {
			game.Rounds = append(game.Rounds, round)
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
				Name:    themeName,
				Comment: themeComment,
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

				qType := "standard"
				var customTheme *string

				typeXML := qXML.Find("type")
				if typeXML != nil {
					typeName := typeXML.Attr("name")
					if typeName == "auction" {
						qType = "auction"
					} else if typeName == "cat" || typeName == "bagcat" {
						qType = "bagcat"
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

				var atoms []*storage.XMLElement
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
			round.Themes = append(round.Themes, theme)
		}
		game.Rounds = append(game.Rounds, round)
	}

	for _, round := range game.Rounds {
		maxQuestions := 0
		for _, t := range round.Themes {
			if len(t.Questions) > maxQuestions {
				maxQuestions = len(t.Questions)
			}
		}
		for _, t := range round.Themes {
			for len(t.Questions) < maxQuestions {
				t.Questions = append(t.Questions, &Question{
					Answer:      "-",
					Value:       0,
					Comment:     "-",
					Type:        "standard",
					IsProcessed: true,
				})
			}
		}
	}

	lastRound := game.Rounds[len(game.Rounds)-1]
	if len(lastRound.Themes) > 0 && len(lastRound.Themes[0].Questions) == 1 {
		lastRound.IsFinal = true
	}

	game.RoundCount = len(game.Rounds)

	return nil
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
	game.Players = append(game.Players, &Player{Name: name})
	return nil
}

func (game *Game) Tick(time.Duration) (*abstract.Command, error) {
	return nil, nil
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
	case "chooseQuestion":
		name, ok := params["question"].(string)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.chooseQuestion(name)
	case "setAnswererAndBet":
		playerName, ok1 := params["player"].(string)
		bet, ok2 := params["bet"].(float64)
		if !ok1 || !ok2 {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.setAnswererAndBet(playerName, int(bet))
	case "skipQuestion":
		return gameCommand, game.skipQuestion()
	case "buttonClick":
		playerName, ok := params["player"].(string)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.buttonClick(playerName)
	case "answer":
		isRight, ok := params["correct"].(bool)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.answerQuestion(isRight)
	case "removeFinalTheme":
		themeName, ok := params["theme"].(string)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.removeFinalTheme(themeName)
	case "finalBet":
		playerName, ok1 := params["player"].(string)
		bet, ok2 := params["bet"].(float64)
		if !ok1 || !ok2 {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.finalBet(playerName, int(bet))
	case "finalAnswer":
		playerName, ok1 := params["player"].(string)
		answer, ok2 := params["answer"].(string)
		if !ok1 || !ok2 {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.finalAnswerPlayer(playerName, answer)
	case "finalPlayerAnswer":
		isRight, ok := params["correct"].(bool)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.finalPlayerAnswer(isRight)
	case "setBalance":
		arr, ok := params["balanceList"].([]any)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		bl := make([]int, len(arr))
		for i, item := range arr {
			balance, ok := item.(float64)
			if !ok {
				return nil, abstract.InvalidInputs
			}
			bl[i] = int(balance)
		}
		game.setBalance(bl)
		return gameCommand, nil
	case "setRound":
		round, ok := params["round"].(float64)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		game.setRound(int(round))
		return gameCommand, nil
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
		if len(game.Players) >= 2 {
			game.State.Value = "intro"
		} else {
			return abstract.InvalidInputs
		}
	case "intro":
		game.State.Value = "themes_all"
		game.Themes = game.allThemeNames()
	case "themes_all":
		game.Themes = nil
		game.Round = game.nextUnprocessedRound()
		game.State.Value = "round"
	case "round":
		if game.Round != nil && game.Round.IsFinal {
			game.State.Value = "final_themes"
			count, theme := game.getNonRemovedThemes()
			if count == 1 {
				game.State.Value = "final_bets"
				game.State.Question = theme.Questions[0]
			}
		} else {
			game.State.Value = "round_themes"
		}
	case "round_themes":
		game.State.Value = "questions"
	case "questions":
		return abstract.NothingToDo
	case "question_event":
		return abstract.NothingToDo
	case "question":
		game.State.Value = "answer"
	case "answer":
		return abstract.NothingToDo
	case "question_end":
		game.processQuestionEnd()
	case "final_themes":
		return abstract.NothingToDo
	case "final_bets":
		canProceed := false
		for _, player := range game.Players {
			if player.Balance > 0 && player.FinalBet > 0 {
				canProceed = true
			}
		}
		if !canProceed {
			return errors.New("at least one player must place a bet to proceed to the final question")
		}
		game.State.Value = "final_question"
	case "final_question":
		game.State.Value = "final_answer"
	case "final_answer":
		game.nextFinalEndState()
	case "final_player_answer":
		return abstract.NothingToDo
	case "final_player_bet":
		answerer := game.State.Answerer
		answerer.FinalBet = 0
		game.State.Answerer = nil
		game.nextFinalEndState()
	default:
		return abstract.InvalidInputs
	}
	return nil
}

func (game *Game) nextUnprocessedRound() *Round {
	for _, r := range game.Rounds {
		if !r.IsProcessed {
			return r
		}
	}
	return nil
}

func (game *Game) allThemeNames() []string {
	var names []string
	for _, r := range game.Rounds {
		if r.IsFinal {
			continue
		}
		for _, t := range r.Themes {
			names = append(names, t.Name)
		}
	}
	return names
}

func (game *Game) findQuestionByName(questionName string) (*Question, *Theme) {
	if game.Round == nil {
		return nil, nil
	}
	for _, t := range game.Round.Themes {
		for _, q := range t.Questions {
			key := t.Name + ":" + strconv.Itoa(q.Value)
			if key == questionName {
				return q, t
			}
		}
	}
	return nil, nil
}

func (game *Game) findPlayer(name string) *Player {
	for _, p := range game.Players {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (game *Game) hasUnprocessedQuestions() bool {
	if game.Round == nil {
		return false
	}
	for _, t := range game.Round.Themes {
		for _, q := range t.Questions {
			if !q.IsProcessed {
				return true
			}
		}
	}
	return false
}

func (game *Game) processQuestionEnd() {
	game.State.Question.IsProcessed = true

	if game.hasUnprocessedQuestions() {
		game.State.Question = nil
		game.State.Value = "questions"
	} else {
		game.State.Question = nil
		if game.Round != nil {
			game.Round.IsProcessed = true
		}
		game.Round = game.nextUnprocessedRound()
		if game.Round == nil {
			game.State.Value = "game_end"
		} else {
			game.State.Value = "round"
		}
	}
}

func (game *Game) nextFinalEndState() {
	for _, p := range game.Players {
		if p.FinalBet > 0 {
			game.State.Answerer = p
			game.State.Value = "final_player_answer"
			return
		}
	}
	game.State.Value = "game_end"
}

func (game *Game) getNonRemovedThemes() (int, *Theme) {
	if game.Round == nil {
		return 0, nil
	}
	count := 0
	var first *Theme
	for _, t := range game.Round.Themes {
		if !t.IsRemoved {
			count++
			if first == nil {
				first = t
			}
		}
	}
	return count, first
}

func (game *Game) chooseQuestion(questionName string) error {
	q, _ := game.findQuestionByName(questionName)
	if q == nil || q.IsProcessed {
		return abstract.InvalidInputs
	}

	if q.Type == "standard" {
		game.State.Value = "question"
	} else {
		game.State.Value = "question_event"
	}
	game.State.Question = q
	return nil
}

func (game *Game) setAnswererAndBet(playerName string, bet int) error {
	if game.State.Value != "question_event" {
		return abstract.NothingToDo
	}
	if bet <= 0 {
		return abstract.InvalidInputs
	}
	player := game.findPlayer(playerName)
	if player == nil {
		return abstract.InvalidInputs
	}
	game.State.QuestionBet = bet
	game.State.Answerer = player
	game.State.Value = "question"
	return nil
}

func (game *Game) skipQuestion() error {
	if game.State.Value != "question_event" && game.State.Value != "question" && game.State.Value != "answer" {
		return abstract.NothingToDo
	}
	game.State.QuestionBet = 0
	game.State.Answerer = nil

	if game.State.Question.AnswerText != nil || game.State.Question.AnswerImage != nil ||
		game.State.Question.AnswerAudio != nil || game.State.Question.AnswerVideo != nil {
		game.State.Value = "question_end"
	} else {
		game.processQuestionEnd()
	}
	game.Intercom("skip")
	return nil
}

func (game *Game) buttonClick(playerName string) error {
	if game.State.Value != "answer" || game.State.Answerer != nil {
		return abstract.NothingToDo
	}
	if game.State.Question.Type != "standard" {
		return abstract.NothingToDo
	}
	player := game.findPlayer(playerName)
	if player == nil {
		return abstract.InvalidInputs
	}
	game.State.Answerer = player
	return nil
}

func (game *Game) answerQuestion(isRight bool) error {
	if game.State.Value != "answer" || game.State.Answerer == nil {
		return abstract.NothingToDo
	}

	questionEnd := func() {
		game.State.QuestionBet = 0
		if game.State.Question.AnswerText != nil || game.State.Question.AnswerImage != nil ||
			game.State.Question.AnswerAudio != nil || game.State.Question.AnswerVideo != nil {
			game.State.Value = "question_end"
		} else {
			game.processQuestionEnd()
		}
	}

	if isRight {
		player := game.State.Answerer
		if game.State.Question.Type != "standard" {
			player.Balance += game.State.QuestionBet
		} else {
			player.Balance += game.State.Question.Value
		}
		questionEnd()
	} else {
		player := game.State.Answerer
		if game.State.Question.Type != "standard" {
			player.Balance -= game.State.QuestionBet
		} else {
			player.Balance -= game.State.Question.Value
		}
		if game.State.Question.Type != "standard" {
			questionEnd()
		}
	}
	game.State.Answerer = nil
	return nil
}

func (game *Game) removeFinalTheme(themeName string) error {
	if game.State.Value != "final_themes" {
		return abstract.NothingToDo
	}
	if game.Round == nil {
		return abstract.InvalidInputs
	}
	var theme *Theme
	for _, t := range game.Round.Themes {
		if t.Name == themeName {
			theme = t
			break
		}
	}
	if theme == nil {
		return abstract.InvalidInputs
	}
	theme.IsRemoved = true

	count, theme := game.getNonRemovedThemes()
	if count == 1 {
		game.State.Value = "final_bets"
		game.State.Question = theme.Questions[0]
	}
	return nil
}

func (game *Game) finalBet(playerName string, bet int) error {
	if game.State.Value != "final_bets" {
		return abstract.NothingToDo
	}
	player := game.findPlayer(playerName)
	if player == nil {
		return abstract.InvalidInputs
	}
	if bet <= 0 {
		return abstract.InvalidInputs
	}
	if player.Balance < bet {
		return errors.New("bet cannot be greater than player's balance")
	}
	player.FinalBet = bet
	return nil
}

func (game *Game) finalAnswerPlayer(playerName string, answer string) error {
	if game.State.Value != "final_answer" {
		return abstract.NothingToDo
	}
	if answer == "" {
		return abstract.InvalidInputs
	}
	player := game.findPlayer(playerName)
	if player == nil {
		return abstract.InvalidInputs
	}
	player.FinalAnswer = answer
	return nil
}

func (game *Game) finalPlayerAnswer(isRight bool) error {
	if game.State.Value != "final_player_answer" {
		return abstract.NothingToDo
	}
	answerer := game.State.Answerer
	if isRight {
		answerer.Balance += answerer.FinalBet
	} else {
		answerer.Balance -= answerer.FinalBet
	}
	game.State.Value = "final_player_bet"
	return nil
}

func (game *Game) setBalance(balanceList []int) {
	for i, p := range game.Players {
		if i < len(balanceList) {
			p.Balance = balanceList[i]
		}
	}
}

func (game *Game) setRound(roundNum int) {
	for _, r := range game.Rounds {
		if r.Number == roundNum {
			r.IsProcessed = false
			for _, t := range r.Themes {
				for _, q := range t.Questions {
					q.IsProcessed = false
				}
				t.IsRemoved = false
			}
		} else if r.Number > roundNum {
			r.IsProcessed = false
		} else {
			r.IsProcessed = true
		}
	}

	game.Round = game.nextUnprocessedRound()
	if game.Round != nil {
		game.State.Value = "round"
	} else {
		game.State.Value = "game_end"
	}
	game.State.Answerer = nil
	game.State.QuestionBet = 0
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
