package deal

import (
	"bunjgames/game/abstract"
)

type Present struct {
	Content     string `json:"content" yaml:"content"`
	Description string `json:"description" yaml:"description"`
	Given       bool   `json:"given" yaml:"given"`
	Opened      bool   `json:"opened" yaml:"-"`
	Selected    bool   `json:"selected" yaml:"-"`
}

type State struct {
	Value            string `json:"value"`
	SelectedIndex    int    `json:"selectedIndex"`
	RevealedIndex    int    `json:"revealedIndex"`
	WheelResultIndex int    `json:"wheelResultIndex"`
	CheatIndex       int    `json:"cheatIndex"`
	Round            int    `json:"round"`
}

type Game struct {
	abstract.BaseGame

	State    State     `json:"state"`
	Presents []Present `json:"presents"`
}
