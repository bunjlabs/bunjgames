package abstract

import (
	"errors"
	"io"
	"sync"
	"time"
)

var InvalidInputs = errors.New("invalid input parameters")
var UnknownMethod = errors.New("unknown method")
var NothingToDo = errors.New("nothing to do")
var NotEnoughPlayers = errors.New("not enough players")

type Command struct {
	Type    string `json:"type"`
	Message any    `json:"message"`
}

type Game interface {
	GetToken() string
	Parse(fileStream io.Reader) error
	Tick(delta time.Duration) (*Command, error)
	RegisterPlayer(name string) error
	ProcessCommand(method string, params map[string]any) (*Command, error)
}

type BaseGame struct {
	Mutex sync.RWMutex `json:"-" yaml:"-"`

	Token string `json:"token" yaml:"-"`
	Type  string `json:"type" yaml:"-"`
}
