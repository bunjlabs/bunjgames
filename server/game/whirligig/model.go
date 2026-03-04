package whirligig

import (
	"bunjgames/game/abstract"
)

type Answer struct {
	Description string `json:"description" yaml:"description"`
	Text        string `json:"text" yaml:"text"`
	Image       string `json:"image" yaml:"image"`
	Audio       string `json:"audio" yaml:"audio"`
	Video       string `json:"video" yaml:"video"`
}

type Question struct {
	IsProcessed bool `json:"isProcessed"`

	Description string `json:"description" yaml:"description"`
	Text        string `json:"text" yaml:"text"`
	Image       string `json:"image" yaml:"image"`
	Audio       string `json:"audio" yaml:"audio"`
	Video       string `json:"video" yaml:"video"`

	Answer Answer `json:"answer" yaml:"answer"`

	Author string `json:"author" yaml:"author"`
}

type Item struct {
	Name        string     `json:"name" yaml:"name"`
	Description string     `json:"description" yaml:"description"`
	Type        string     `json:"type" yaml:"type"`
	IsProcessed bool       `json:"is_processed" yaml:"-"`
	Questions   []Question `json:"questions" yaml:"questions"`
}

type Score struct {
	Connoisseurs int `json:"connoisseurs"`
	Viewers      int `json:"viewers"`
}

type State struct {
	Value             string    `json:"value"`
	WhirligigPosition *int      `json:"whirligigPosition"`
	Item              *Item     `json:"item"`
	Question          *Question `json:"question"`
}

type Timer struct {
	Paused     bool  `json:"paused"`
	PausedTime int64 `json:"pausedTime"`
	Time       int64 `json:"time"`
}

type Game struct {
	abstract.BaseGame

	State State  `json:"state"`
	Score Score  `json:"score"`
	Timer Timer  `json:"timer"`
	Items []Item `json:"items"`
}
