package adaptor

import (
	"errors"
	"fmt"
	"strings"

	world "github.com/sapphire-ai-dev/sapphire-world"
)

// if the agent ever needs to connect to multiple worlds simultaneously, it can connect to this adaptor
// which would in turn connect to all required worlds on the agent's behalf
type adaptorWorld struct {
	actors     map[int]*actor      // actorId -> actor
	cycleFuncs map[int]func()      // actorId -> cycle function
	children   map[int]world.World // child world id -> child world
}

func (w *adaptorWorld) registerChild() int {
	child := world.GetWorld()
	if child == nil {
		panic(errWorldNotFound)
	}

	for childWorldId, existingChild := range w.children {
		if existingChild == child {
			return childWorldId
		}
	}

	childWorldId := world.NewUnitId()
	w.children[childWorldId] = child
	return childWorldId
}

func (w *adaptorWorld) Name() string {
	var childrenNames []string
	for _, child := range w.children {
		childrenNames = append(childrenNames, child.Name())
	}
	return fmt.Sprintf("adaptor: [%s]", strings.Join(childrenNames, ", "))
}

func (w *adaptorWorld) Reset() {
	w.actors = map[int]*actor{}
	w.cycleFuncs = map[int]func(){}
	w.children = map[int]world.World{}
}

func (w *adaptorWorld) Tick() {
	for _, f := range w.cycleFuncs {
		f()
	}
}

var (
	errInvalidArgs   = errors.New("invalid args")
	errActorNotFound = errors.New("actor not found")
	errWorldNotFound = errors.New("world not found")
)

func (w *adaptorWorld) NewActor(args ...any) (int, []*world.ActionInterface) {
	if len(args) > 1 {
		panic(errInvalidArgs)
	}

	argsMap, ok := map[int][]any{}, false
	if len(args) == 1 {
		if argsMap, ok = args[0].(map[int][]any); !ok {
			panic(errInvalidArgs)
		}
	}

	actorId := w.newActor()
	return actorId, w.actors[actorId].collectActionInterfaces(argsMap)
}

func (w *adaptorWorld) Register(actorId int, cycle func()) {
	if _, seen := w.actors[actorId]; !seen {
		panic(errActorNotFound)
	}

	w.cycleFuncs[actorId] = cycle
}

func (w *adaptorWorld) Look(actorId int) []*world.Image {
	if _, seen := w.actors[actorId]; !seen {
		return []*world.Image{}
	}

	return w.actors[actorId].look()
}

func (w *adaptorWorld) Feel(actorId int) []*world.Touch {
	if _, seen := w.actors[actorId]; !seen {
		return []*world.Touch{}
	}

	return w.actors[actorId].feel()
}

const (
	CmdTypeLocal = iota
	CmdTypeChild
)

func (w *adaptorWorld) Cmd(args ...any) {
	if len(args) < 1 {
		panic(errInvalidArgs)
	}

	cmdType, cmdTypeOk := args[0].(int)
	if !cmdTypeOk {
		panic(errInvalidArgs)
	}

	if cmdType == CmdTypeLocal {
		w.CmdLocal(args[1:])
	} else if cmdType == CmdTypeChild {
		if len(args) < 2 {
			panic(errInvalidArgs)
		}

		worldId, worldIdOk := args[1].(int)
		if !worldIdOk {
			panic(errInvalidArgs)
		}

		if child, seen := w.children[worldId]; !seen {
			panic(errWorldNotFound)
		} else {
			child.Cmd(args[2:]...)
		}
	} else {
		panic(errInvalidArgs)
	}
}

func (w *adaptorWorld) CmdLocal(args ...any) {

}

func newAdaptorWorld() *adaptorWorld {
	result := &adaptorWorld{}
	result.Reset()
	return result
}

var tempSingleton *adaptorWorld

func InitStart() {
    tempSingleton = newAdaptorWorld()
}

// Proxy the currently registered world and return the newly created child world id
func Proxy() int {
    if tempSingleton == nil {
        panic(errWorldNotFound)
    }

    return tempSingleton.registerChild()
}

func InitComplete() {
    if tempSingleton == nil {
        panic(errWorldNotFound)
    }

    world.SetWorld(tempSingleton)
}
