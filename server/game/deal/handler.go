package deal

import (
	"bunjgames/game/abstract"
	"errors"
	"io"
	"math/rand"
	"time"

	"gopkg.in/yaml.v3"
)

func NewGame() *Game {
	game := &Game{
		State: State{
			Value:            "round_start",
			SelectedIndex:    -1,
			RevealedIndex:    -1,
			WheelResultIndex: -1,
			CheatIndex:       -1,
			Round:            1,
		},
	}
	game.Initialise()
	game.Type = "deal"
	return game
}

func (game *Game) Parse(fileStream io.Reader) error {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	data, err := io.ReadAll(fileStream)
	if err != nil {
		return abstract.InvalidInputs
	}

	var presents []Present
	if err := yaml.Unmarshal(data, &presents); err != nil {
		return errors.New("failed to parse YAML: " + err.Error())
	}

	if len(presents) < 3 {
		return errors.New("at least three presents are required")
	}

	for i := range presents {
		presents[i].Order = i
	}

	notGiven := 0
	for _, p := range presents {
		if !p.Given {
			notGiven++
		}
	}
	if notGiven < 2 {
		return errors.New("at least two presents must not be given")
	}

	game.Presents = presents
	return nil
}

func (game *Game) Tick(time.Duration) (*abstract.Command, error) {
	return nil, nil
}

func (game *Game) RegisterPlayer(string) error {
	return abstract.NothingToDo
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
	case "selectCase":
		index, ok := params["index"].(float64)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.selectCase(int(index))
	case "openCase":
		index, ok := params["index"].(float64)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.openCase(int(index))
	case "revealCase":
		index, ok := params["index"].(float64)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.revealCase(int(index))
	case "chooseMoney":
		return gameCommand, game.chooseMoney()
	case "shuffle":
		return gameCommand, game.shuffle()
	case "keep":
		return gameCommand, game.keep()
	case "switch":
		return gameCommand, game.switchCase()
	case "keepGift":
		return gameCommand, game.keepGift()
	case "spinWheel":
		return gameCommand, game.spinWheel()
	case "setCheat":
		index, ok := params["index"].(float64)
		if !ok {
			return nil, abstract.InvalidInputs
		}
		return gameCommand, game.setCheat(int(index))
	case "spin":
		return gameCommand, game.spin()
	default:
		return nil, abstract.UnknownMethod
	}
}

func (game *Game) nextState(fromState string) error {
	if fromState != "" && game.State.Value != fromState {
		return abstract.NothingToDo
	}

	switch game.State.Value {
	case "round_start":
		game.State.Value = "case_select"
		game.startRound()
	case "case_select", "haggle", "case_swap", "reveal", "wheel_start":
		return abstract.NothingToDo
	case "wheel_spin":
		game.State.Value = "round_end"
		game.enterRoundEnd()
	case "round_end":
		if game.availableCount() == 0 {
			game.State.Value = "game_end"
		} else {
			game.State.Value = "round_start"
			game.State.Round++
		}
	case "game_end":
		return abstract.NothingToDo
	default:
		return abstract.InvalidInputs
	}
	return nil
}

func (game *Game) startRound() {
	game.State.SelectedIndex = -1
	game.State.RevealedIndex = -1
	game.State.WheelResultIndex = -1
	game.State.CheatIndex = -1
	game.State.ChosenMoney = false
	for i := range game.Presents {
		game.Presents[i].Selected = false
		game.Presents[i].Opened = false
	}
	game.reshuffle()
}

func (game *Game) selectCase(index int) error {
	if game.State.Value != "case_select" {
		return abstract.NothingToDo
	}
	if index < 0 || index >= len(game.Presents) || game.Presents[index].Given {
		return abstract.InvalidInputs
	}

	game.State.SelectedIndex = index
	game.Presents[index].Selected = true
	game.Intercom("select")

	switch game.availableCount() {
	case 1:
		game.State.Value = "reveal"
		game.reveal(index)
	case 2:
		game.State.Value = "case_swap"
	default:
		game.State.Value = "haggle"
	}
	return nil
}

func (game *Game) openCase(index int) error {
	if game.State.Value != "haggle" {
		return abstract.NothingToDo
	}
	if index < 0 || index >= len(game.Presents) {
		return abstract.InvalidInputs
	}
	if game.Presents[index].Given || game.Presents[index].Selected || game.Presents[index].Opened {
		return abstract.NothingToDo
	}

	game.Presents[index].Opened = true
	game.Intercom("box_open")

	if game.unopenedCount() == 1 {
		game.State.Value = "case_swap"
	}
	return nil
}

func (game *Game) keep() error {
	if game.State.Value != "case_swap" {
		return abstract.NothingToDo
	}
	game.State.Value = "reveal"
	game.reveal(game.State.SelectedIndex)
	return nil
}

func (game *Game) switchCase() error {
	if game.State.Value != "case_swap" {
		return abstract.NothingToDo
	}

	other := game.soleOtherIndex()
	if other < 0 {
		return abstract.InvalidInputs
	}

	game.Presents[game.State.SelectedIndex].Selected = false
	game.State.SelectedIndex = other
	game.Presents[other].Selected = true

	game.State.Value = "reveal"
	game.reveal(other)
	return nil
}

// revealCase selects the given case as the player's case and reveals it,
// used when the admin clicks a present in the list during haggle/case_swap.
func (game *Game) revealCase(index int) error {
	if game.State.Value != "haggle" && game.State.Value != "case_swap" {
		return abstract.NothingToDo
	}
	if index < 0 || index >= len(game.Presents) {
		return abstract.InvalidInputs
	}
	if game.Presents[index].Given {
		return abstract.InvalidInputs
	}

	if game.State.SelectedIndex >= 0 {
		game.Presents[game.State.SelectedIndex].Selected = false
	}
	game.State.SelectedIndex = index
	game.Presents[index].Selected = true

	game.State.Value = "reveal"
	game.reveal(index)
	return nil
}

// chooseMoney clears the case selection and reveals a "Money" outcome instead.
func (game *Game) chooseMoney() error {
	if game.State.Value != "haggle" && game.State.Value != "case_swap" {
		return abstract.NothingToDo
	}

	if game.State.SelectedIndex >= 0 {
		game.Presents[game.State.SelectedIndex].Selected = false
	}
	game.State.SelectedIndex = -1
	game.State.RevealedIndex = -1
	game.State.ChosenMoney = true

	game.State.Value = "reveal"
	return nil
}

// shuffle reorders the not-given presents without changing the game state.
// It is only allowed before a case is selected, since reshuffling afterwards
// would move already-selected or already-opened cases.
func (game *Game) shuffle() error {
	if game.State.Value != "case_select" {
		return abstract.NothingToDo
	}
	game.reshuffle()
	return nil
}

func (game *Game) reveal(index int) {
	game.State.RevealedIndex = index
	for i := range game.Presents {
		if !game.Presents[i].Given {
			game.Presents[i].Opened = true
		}
	}
}

func (game *Game) keepGift() error {
	if game.State.Value != "reveal" {
		return abstract.NothingToDo
	}
	game.State.Value = "round_end"
	game.enterRoundEnd()
	return nil
}

func (game *Game) spinWheel() error {
	if game.State.Value != "reveal" {
		return abstract.NothingToDo
	}
	game.State.Value = "wheel_start"
	return nil
}

func (game *Game) setCheat(index int) error {
	if game.State.Value != "wheel_start" {
		return abstract.NothingToDo
	}
	if index < -1 || index >= len(game.Presents) {
		return abstract.InvalidInputs
	}
	if index >= 0 && game.Presents[index].Given {
		return abstract.InvalidInputs
	}
	game.State.CheatIndex = index
	return nil
}

func (game *Game) spin() error {
	if game.State.Value != "wheel_start" {
		return abstract.NothingToDo
	}

	available := game.availableIndices()
	if len(available) == 0 {
		return abstract.InvalidInputs
	}

	result := game.State.SelectedIndex
	if game.State.CheatIndex >= 0 && !game.Presents[game.State.CheatIndex].Given {
		result = game.State.CheatIndex
	}
	// When money was chosen there is no selected case; the wheel falls on a
	// random still-available present instead.
	if result < 0 || game.Presents[result].Given {
		result = available[rand.Intn(len(available))]
	}

	game.State.WheelResultIndex = result
	game.State.Value = "wheel_spin"
	return nil
}

func (game *Game) enterRoundEnd() {
	index := game.State.WheelResultIndex
	if index < 0 {
		index = game.State.RevealedIndex
	}
	if index >= 0 && index < len(game.Presents) {
		game.Presents[index].Given = true
		game.Presents[index].Selected = false
		game.Presents[index].Opened = true
	}
}

func (game *Game) reshuffle() {
	indices := game.availableIndices()
	for i := len(indices) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		game.Presents[indices[i]].Order, game.Presents[indices[j]].Order =
			game.Presents[indices[j]].Order, game.Presents[indices[i]].Order
	}
}

func (game *Game) availableIndices() []int {
	var result []int
	for i, p := range game.Presents {
		if !p.Given {
			result = append(result, i)
		}
	}
	return result
}

func (game *Game) availableCount() int {
	return len(game.availableIndices())
}

func (game *Game) unopenedCount() int {
	count := 0
	for _, p := range game.Presents {
		if !p.Given && !p.Selected && !p.Opened {
			count++
		}
	}
	return count
}

func (game *Game) soleOtherIndex() int {
	for i, p := range game.Presents {
		if i == game.State.SelectedIndex {
			continue
		}
		if !p.Given && !p.Opened {
			return i
		}
	}
	return -1
}
