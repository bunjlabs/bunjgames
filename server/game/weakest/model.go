package weakest

import (
	"bunjgames/game/abstract"
)

type Question struct {
	Question  string `json:"question" yaml:"question"`
	Answer    string `json:"answer" yaml:"answer"`
	Processed bool   `json:"processed" yaml:"-"`
}

type Player struct {
	Name         string  `json:"name"`
	Active       bool    `json:"active"`
	Vote         *Player `json:"-"`
	VoteName     string  `json:"vote"`
	RightAnswers int     `json:"rightAnswers"`
	BankIncome   int     `json:"bankIncome"`
	FinalScore   []bool  `json:"finalScore" yaml:"-"`
}

type State struct {
	Value string `json:"value"`
	Score int    `json:"score"`
}

type RoundState struct {
	Number int `json:"number"`
	Score  int `json:"score"`
	Bank   int `json:"bank"`

	Question     *Question `json:"question"`
	Time         int64     `json:"time"`
	QuestionTime int64     `json:"questionTime"`

	Answerer  *Player `json:"answerer"`
	Weakest   *Player `json:"weakest"`
	Strongest *Player `json:"strongest"`
	Kicked    *Player `json:"kicked"`
}

type Game struct {
	abstract.BaseGame

	State           State       `json:"state" yaml:"-"`
	RoundState      RoundState  `json:"roundState" yaml:"-"`
	ScoreMultiplier int         `json:"scoreMultiplier" yaml:"scoreMultiplier"`
	Questions       []*Question `json:"-" yaml:"questions"`
	FinalQuestions  []*Question `json:"-" yaml:"finalQuestions"`
	Players         []*Player   `json:"players" yaml:"-"`
}
