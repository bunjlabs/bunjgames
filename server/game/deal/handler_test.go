package deal

import (
	"fmt"
	"strings"
	"testing"

	"bunjgames/game/abstract"

	"github.com/stretchr/testify/assert"
)

func generateYaml(presents int, givenIndices ...int) string {
	given := make(map[int]bool)
	for _, i := range givenIndices {
		given[i] = true
	}
	var b strings.Builder
	for i := range presents {
		b.WriteString(fmt.Sprintf("  - content: \"present %d\"\n", i+1))
		if given[i] {
			b.WriteString("    given: true\n")
		}
	}
	return b.String()
}

func TestParse(test *testing.T) {
	test.Parallel()

	game := NewGame()
	assert.NotEmpty(test, game.GetToken())

	err := game.Parse(strings.NewReader(generateYaml(4)))
	assert.Nil(test, err)
	assert.Equal(test, 4, len(game.Presents))
	assert.Equal(test, "round_start", game.State.Value)

	err = game.Parse(strings.NewReader(generateYaml(2)))
	assert.NotNil(test, err)
}

func TestParseGivenValidation(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(4, 0, 1, 2)))
	assert.NotNil(test, err)
}

func TestParseGiven(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(4, 0)))
	assert.Nil(test, err)
	assert.Equal(test, 3, game.availableCount())
}

func TestGame(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(4)))
	assert.Nil(test, err)

	_, err = game.ProcessCommand("next", map[string]any{"from": "round_start"})
	assert.Nil(test, err)
	assert.Equal(test, "case_select", game.State.Value)

	_, err = game.ProcessCommand("selectCase", map[string]any{"index": float64(0)})
	assert.Nil(test, err)
	assert.Equal(test, "haggle", game.State.Value)
	assert.True(test, game.Presents[0].Selected)

	for i := 1; i < 3; i++ {
		_, err = game.ProcessCommand("openCase", map[string]any{"index": float64(i)})
		assert.Nil(test, err)
	}
	assert.Equal(test, "case_swap", game.State.Value)
	assert.Equal(test, 0, game.State.SelectedIndex)
}

func TestGameFull(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(3)))
	assert.Nil(test, err)

	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_start"})
	_, _ = game.ProcessCommand("selectCase", map[string]any{"index": float64(0)})
	_, _ = game.ProcessCommand("openCase", map[string]any{"index": float64(1)})
	assert.Equal(test, "case_swap", game.State.Value)

	_, _ = game.ProcessCommand("keep", map[string]any{})
	assert.Equal(test, "reveal", game.State.Value)
	assert.Equal(test, 0, game.State.RevealedIndex)

	_, _ = game.ProcessCommand("keepGift", map[string]any{})
	assert.Equal(test, "round_end", game.State.Value)
	assert.True(test, game.Presents[0].Given)

	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_end"})
	assert.Equal(test, "round_start", game.State.Value)
}

func TestSwitch(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(4)))
	assert.Nil(test, err)

	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_start"})
	_, _ = game.ProcessCommand("selectCase", map[string]any{"index": float64(0)})
	_, _ = game.ProcessCommand("openCase", map[string]any{"index": float64(1)})
	_, _ = game.ProcessCommand("openCase", map[string]any{"index": float64(2)})
	assert.Equal(test, "case_swap", game.State.Value)
	assert.Equal(test, 0, game.State.SelectedIndex)

	_, _ = game.ProcessCommand("switch", map[string]any{})
	assert.Equal(test, "reveal", game.State.Value)
	assert.Equal(test, 3, game.State.SelectedIndex)
}

func TestWheel(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(4)))
	assert.Nil(test, err)

	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_start"})
	_, _ = game.ProcessCommand("selectCase", map[string]any{"index": float64(0)})
	_, _ = game.ProcessCommand("openCase", map[string]any{"index": float64(1)})
	_, _ = game.ProcessCommand("openCase", map[string]any{"index": float64(2)})
	_, _ = game.ProcessCommand("keep", map[string]any{})
	assert.Equal(test, "reveal", game.State.Value)

	_, _ = game.ProcessCommand("spinWheel", map[string]any{})
	assert.Equal(test, "wheel_start", game.State.Value)

	_, _ = game.ProcessCommand("setCheat", map[string]any{"index": float64(2)})
	assert.Equal(test, 2, game.State.CheatIndex)

	_, _ = game.ProcessCommand("spin", map[string]any{})
	assert.Equal(test, "wheel_spin", game.State.Value)
	assert.Equal(test, 2, game.State.WheelResultIndex)

	_, _ = game.ProcessCommand("next", map[string]any{"from": "wheel_spin"})
	assert.Equal(test, "round_end", game.State.Value)
	assert.True(test, game.Presents[2].Given)
}

func TestRevealOpensAllCases(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(4)))
	assert.Nil(test, err)

	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_start"})
	_, _ = game.ProcessCommand("selectCase", map[string]any{"index": float64(0)})
	_, _ = game.ProcessCommand("openCase", map[string]any{"index": float64(1)})
	_, _ = game.ProcessCommand("openCase", map[string]any{"index": float64(2)})
	assert.Equal(test, "case_swap", game.State.Value)

	_, _ = game.ProcessCommand("keep", map[string]any{})
	assert.Equal(test, "reveal", game.State.Value)

	for _, p := range game.Presents {
		assert.True(test, p.Opened)
	}
	assert.True(test, game.Presents[0].Selected)
}

func TestRevealCase(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(5)))
	assert.Nil(test, err)

	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_start"})
	_, _ = game.ProcessCommand("selectCase", map[string]any{"index": float64(0)})
	assert.Equal(test, "haggle", game.State.Value)

	_, _ = game.ProcessCommand("revealCase", map[string]any{"index": float64(3)})
	assert.Equal(test, "reveal", game.State.Value)
	assert.Equal(test, 3, game.State.SelectedIndex)
	assert.True(test, game.Presents[3].Selected)
	assert.False(test, game.Presents[0].Selected)
}

func TestChooseMoney(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(5)))
	assert.Nil(test, err)

	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_start"})
	_, _ = game.ProcessCommand("selectCase", map[string]any{"index": float64(0)})
	assert.Equal(test, "haggle", game.State.Value)

	_, _ = game.ProcessCommand("chooseMoney", map[string]any{})
	assert.Equal(test, "reveal", game.State.Value)
	assert.Equal(test, -1, game.State.SelectedIndex)
	assert.True(test, game.State.ChosenMoney)
	assert.False(test, game.Presents[0].Selected)

	// Keeping money should not mark anything given.
	_, _ = game.ProcessCommand("keepGift", map[string]any{})
	assert.Equal(test, "round_end", game.State.Value)
	for _, p := range game.Presents {
		assert.False(test, p.Given)
	}
}

func TestChooseMoneyThenSpin(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(5)))
	assert.Nil(test, err)

	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_start"})
	_, _ = game.ProcessCommand("selectCase", map[string]any{"index": float64(0)})
	_, _ = game.ProcessCommand("chooseMoney", map[string]any{})
	assert.Equal(test, "reveal", game.State.Value)

	_, _ = game.ProcessCommand("spinWheel", map[string]any{})
	assert.Equal(test, "wheel_start", game.State.Value)

	_, _ = game.ProcessCommand("spin", map[string]any{})
	assert.Equal(test, "wheel_spin", game.State.Value)
	assert.True(test, game.State.WheelResultIndex >= 0)

	_, _ = game.ProcessCommand("next", map[string]any{"from": "wheel_spin"})
	assert.Equal(test, "round_end", game.State.Value)
	assert.True(test, game.Presents[game.State.WheelResultIndex].Given)
}

func TestShuffle(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(5)))
	assert.Nil(test, err)

	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_start"})
	assert.Equal(test, "case_select", game.State.Value)

	before := make([]int, len(game.Presents))
	for i, p := range game.Presents {
		before[i] = p.Order
	}

	_, _ = game.ProcessCommand("shuffle", map[string]any{})
	assert.Equal(test, "case_select", game.State.Value)

	// Orders should still be a permutation of the same values.
	after := make([]int, len(game.Presents))
	for i, p := range game.Presents {
		after[i] = p.Order
	}
	assert.ElementsMatch(test, before, after)

	// Shuffle is disabled after a case is selected.
	_, _ = game.ProcessCommand("selectCase", map[string]any{"index": float64(0)})
	assert.Equal(test, "haggle", game.State.Value)
	orders := make([]int, len(game.Presents))
	for i, p := range game.Presents {
		orders[i] = p.Order
	}
	_, err = game.ProcessCommand("shuffle", map[string]any{})
	assert.Equal(test, abstract.NothingToDo, err)
	for i, p := range game.Presents {
		assert.Equal(test, orders[i], p.Order)
	}

	_, _ = game.ProcessCommand("revealCase", map[string]any{"index": float64(0)})
	assert.Equal(test, "reveal", game.State.Value)
}

func TestGameEnd(test *testing.T) {
	test.Parallel()

	game := NewGame()
	err := game.Parse(strings.NewReader(generateYaml(3)))
	assert.Nil(test, err)

	// Round 1: eliminate one present.
	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_start"})
	_, _ = game.ProcessCommand("selectCase", map[string]any{"index": float64(0)})
	_, _ = game.ProcessCommand("openCase", map[string]any{"index": float64(1)})
	_, _ = game.ProcessCommand("keep", map[string]any{})
	_, _ = game.ProcessCommand("keepGift", map[string]any{})
	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_end"})
	assert.Equal(test, "round_start", game.State.Value)
	assert.Equal(test, 2, game.State.Round)

	// Round 2: two presents remain, so selecting goes straight to case_swap.
	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_start"})
	assert.Equal(test, "case_select", game.State.Value)

	available := game.availableIndices()
	_, _ = game.ProcessCommand("selectCase", map[string]any{"index": float64(available[0])})
	assert.Equal(test, "case_swap", game.State.Value)

	_, _ = game.ProcessCommand("keep", map[string]any{})
	_, _ = game.ProcessCommand("keepGift", map[string]any{})
	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_end"})
	assert.Equal(test, "round_start", game.State.Value)
	assert.Equal(test, 3, game.State.Round)

	// Round 3: one present remains, so selecting goes to reveal.
	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_start"})
	available = game.availableIndices()
	_, _ = game.ProcessCommand("selectCase", map[string]any{"index": float64(available[0])})
	assert.Equal(test, "reveal", game.State.Value)

	_, _ = game.ProcessCommand("keepGift", map[string]any{})
	_, _ = game.ProcessCommand("next", map[string]any{"from": "round_end"})
	assert.Equal(test, "game_end", game.State.Value)
}
