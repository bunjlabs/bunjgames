package jeopardy

import (
	"bunjgames/game/abstract"
)

type Question struct {
	CustomTheme *string `json:"customTheme"`
	Text        *string `json:"text"`
	Image       *string `json:"image"`
	Audio       *string `json:"audio"`
	Video       *string `json:"video"`
	Answer      string  `json:"answer"`
	AnswerText  *string `json:"answerText"`
	AnswerImage *string `json:"answerImage"`
	AnswerAudio *string `json:"answerAudio"`
	AnswerVideo *string `json:"answerVideo"`
	Value       int     `json:"value"`
	Comment     string  `json:"comment"`
	Type        string  `json:"type"`
	IsProcessed bool    `json:"isProcessed"`
}

type Theme struct {
	Name      string      `json:"name"`
	Comment   *string     `json:"comment"`
	IsRemoved bool        `json:"isRemoved"`
	Questions []*Question `json:"questions"`
}

type Player struct {
	Name        string `json:"name"`
	Balance     int    `json:"balance"`
	FinalBet    int    `json:"finalBet"`
	FinalAnswer string `json:"finalAnswer"`
}

type Round struct {
	Number      int      `json:"number"`
	IsFinal     bool     `json:"isFinal"`
	Themes      []*Theme `json:"themes"`
	IsProcessed bool     `json:"isProcessed"`
}

type State struct {
	Value       string    `json:"value"`
	Answerer    *Player   `json:"answerer"`
	Question    *Question `json:"question"`
	QuestionBet int       `json:"-"`
}

type Game struct {
	abstract.BaseGame

	Rounds     []*Round  `json:"-"`
	Round      *Round    `json:"round"`
	RoundCount int       `json:"roundCount"`
	State      State     `json:"state"`
	Themes     []string  `json:"themes"`
	Players    []*Player `json:"players"`
}
