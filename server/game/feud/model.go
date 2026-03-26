package feud

import (
	"bunjgames/game/abstract"
)

type Answer struct {
	Text            string `json:"text" yaml:"text"`
	Value           int    `json:"value" yaml:"value"`
	IsOpened        bool   `json:"isOpened" yaml:"-"`
	IsFinalAnswered bool   `json:"isFinalAnswered" yaml:"-"`
}

type Question struct {
	Text        string    `json:"text" yaml:"text"`
	IsProcessed bool      `json:"isProcessed" yaml:"-"`
	Answers     []*Answer `json:"answers" yaml:"answers"`
}

type Player struct {
	Name       string `json:"name"`
	Strikes    int    `json:"strikes"`
	Score      int    `json:"score"`
	FinalScore int    `json:"finalScore"`
}

type ParsedGame struct {
	Questions      []*Question `yaml:"questions"`
	FinalQuestions []*Question `yaml:"finalQuestions"`
}

type Game struct {
	abstract.BaseGame

	Round          int         `json:"round"`
	State          string      `json:"state"`
	Questions      []*Question `json:"-"`
	Players        []*Player   `json:"players"`
	Question       *Question   `json:"question"`
	Answerer       *Player     `json:"answerer"`
	FinalQuestions []*Question `json:"finalQuestions"`
	Timer          int64       `json:"timer"`
}
