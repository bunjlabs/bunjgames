package abstract

import (
	"errors"
	"io"
	"sync"
)

var InvalidInputs = errors.New("invalid input parameters")
var UnknownMethod = errors.New("unknown method")
var NothingToDo = errors.New("nothing to do")

type Game interface {
	GetToken() string
	Parse(fileStream io.Reader) error
	ProcessCommand(method string, params map[string]any) (any, error)
	Serialize() any
}

type BaseGame struct {
	Mutex sync.RWMutex

	Token string `json:"token" yaml:"-"`
}
