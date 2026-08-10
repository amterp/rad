package core

import (
	"fmt"
	"sort"
	"strings"

	com "github.com/amterp/rad/core/common"
	"github.com/amterp/rad/rts/rl"
)

type Env struct {
	i             *Interpreter
	Enclosing     *Env
	Vars          map[string]RadValue
	JsonFieldVars map[string]*JsonFieldVar // not pointer?
}

func NewEnv(i *Interpreter) *Env {
	return &Env{
		i:             i,
		Enclosing:     nil,
		Vars:          make(map[string]RadValue),
		JsonFieldVars: make(map[string]*JsonFieldVar),
	}
}

func (e *Env) NewChildEnv() Env {
	return Env{
		i:             e.i,
		Enclosing:     e,
		Vars:          make(map[string]RadValue),
		JsonFieldVars: make(map[string]*JsonFieldVar),
	}
}

func (e *Env) SetVar(name string, v RadValue) {
	e.SetVarUpdatingEnclosing(name, v, false)
}

func (e *Env) SetVarUpdatingEnclosing(name string, v RadValue, updateEnclosing bool) {
	e.setVar(name, v, updateEnclosing)
}

func (e *Env) GetVar(name string) (RadValue, bool) {
	if val, exists := e.Vars[name]; exists {
		return val, true
	}
	if e.Enclosing != nil {
		return e.Enclosing.GetVar(name)
	}
	return RadValue{}, false
}

func (e *Env) GetVarElseBug(i *Interpreter, node rl.Node, name string) RadValue {
	if val, exists := e.GetVar(name); exists {
		return val
	}
	i.emitError(rl.ErrInternalBug, node, "Bug: Expected variable but didn't find: "+name)
	panic(UNREACHABLE)
}

func (e *Env) SetJsonFieldVar(jsonFieldVar *JsonFieldVar) {
	e.JsonFieldVars[jsonFieldVar.Name] = jsonFieldVar
	// define empty list for json field
	e.SetVar(jsonFieldVar.Name, newRadValue(e.i, nil, NewRadList()))
}

func (e *Env) GetJsonFieldVar(name string) (*JsonFieldVar, bool) {
	if val, exists := e.JsonFieldVars[name]; exists {
		return val, true
	}
	if e.Enclosing != nil {
		return e.Enclosing.GetJsonFieldVar(name)
	}
	return nil, false
}

// setVar writes name into the frame that owns it.
//
// With updateEnclosing - compound assigns and incr/decr, which the
// converter desugars into `x = x <op> y` - that frame is the innermost
// one declaring the name, found the same way GetVar finds it. The two
// halves of `x += y` must resolve to the same binding: reading a
// function-local while writing the module-level namesake means the
// accumulator collects nothing and clobbers the outer value on its way
// past. Builtins share the root frame with module-level variables, so
// the same walk keeps `keys = []` inside a fn from replacing `keys`.
//
// Without it, the write always lands locally - that is how a plain `=`
// shadows an enclosing binding rather than reassigning it.
func (e *Env) setVar(name string, v RadValue, updateEnclosing bool) {
	target := e
	if updateEnclosing {
		for env := e; env != nil; env = env.Enclosing {
			if _, exists := env.Vars[name]; exists {
				target = env
				break
			}
		}
	}

	if v == VOID_SENTINEL {
		delete(target.Vars, name)
	} else {
		target.Vars[name] = v
	}
}

// AllVarNames returns all variable names visible from this environment.
func (e *Env) AllVarNames() []string {
	seen := make(map[string]bool)
	var names []string

	for env := e; env != nil; env = env.Enclosing {
		for name := range env.Vars {
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}

	return names
}

// FindSimilarVars finds variable names similar to the given name.
// Returns at most maxResults names, sorted by similarity.
func (e *Env) FindSimilarVars(name string, maxResults int) []string {
	type candidate struct {
		name     string
		distance int
	}

	var candidates []candidate
	allNames := e.AllVarNames()

	// Only suggest names within a reasonable edit distance
	maxDistance := len(name)/2 + 1
	if maxDistance < 2 {
		maxDistance = 2
	}

	for _, n := range allNames {
		dist := Levenshtein(name, n)
		if dist <= maxDistance && dist > 0 {
			candidates = append(candidates, candidate{n, dist})
		}
	}

	// Sort by distance, then alphabetically
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].distance != candidates[j].distance {
			return candidates[i].distance < candidates[j].distance
		}
		return candidates[i].name < candidates[j].name
	})

	// Return top results
	result := make([]string, 0, maxResults)
	for i := 0; i < len(candidates) && i < maxResults; i++ {
		result = append(result, candidates[i].name)
	}
	return result
}

func (e *Env) PrintShellExports() {
	if AlreadyExportedShellVars {
		return
	}
	AlreadyExportedShellVars = true

	keys := com.SortedKeys(e.Vars)

	printFunc := func(varName, value string) {
		RP.PrintForShellEval(fmt.Sprintf("%s=%s\n", varName, value))
	}

	for _, varName := range keys {
		if varName == "PATH" {
			// skip PATH, if royally messes up users and there doesn't
			// appear to be any legitimate use case for it
			continue
		}

		val := e.Vars[varName]
		switch coerced := val.Val.(type) {
		case RadString, int64, float64, bool:
			printFunc(varName, ToPrintable(val))
		case *RadList:
			printFunc(varName, "("+strings.Join(coerced.AsStringList(true), " ")+")")
		case *RadMap:
			// todo can do some stuff with declare -A ?
			printFunc(varName, "'"+coerced.ToString()+"'")
		case RadFn:
			// skip, doesn't make sense
		case RadNull:
			// skip, implies undefined
		default:
			RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled type for shell export: %T", val.Val))
		}
	}
}
