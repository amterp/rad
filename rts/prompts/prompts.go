// Package prompts finds the places a script would stop and ask the user
// something: the interactive builtins, plus shell commands the author gated
// behind a confirmation.
//
// It exists so rad can tell a caller with no terminal what a script is going to
// need *before* running any of it. Answering that question after the fact is
// too late - by then the script has already done work.
package prompts

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// Kind is what a site will ask for, which determines how a supplied answer is
// parsed. Every site's kind is known statically, so an answer never has to be
// guessed at from its own shape.
type Kind int

const (
	Input Kind = iota + 1
	Confirm
	Pick
	Multipick
	ShellConfirm
)

func (k Kind) String() string {
	switch k {
	case Input:
		return "input"
	case Confirm:
		return "confirm"
	case Pick:
		return "pick"
	case Multipick:
		return "multipick"
	case ShellConfirm:
		return "shell"
	default:
		return fmt.Sprintf("Kind(%d)", int(k))
	}
}

// Site is one place a script would ask the user something.
type Site struct {
	Kind Kind
	// Fn is the builtin's name; empty for a confirm-gated shell command.
	Fn string
	// Key is how a caller addresses this site on the command line. Just the
	// line number, widened to "line.col" only when one line holds several
	// sites. Callers copy it out of rad's own output rather than building it.
	Key  string
	Line int // 1-based
	Col  int // 1-based

	// Prompt is the literal prompt text, empty when it is built at runtime.
	Prompt string
	// Options are the literal choices, nil when they are computed at runtime.
	Options []string
	// Secret marks a literal secret=true input. Those cannot be answered from
	// the command line at all: argv is readable by other processes.
	Secret bool
	// Filtered marks a pick-family call given a filter argument. Such a call
	// may narrow to a single option and never prompt, so a caller expecting
	// that should reach for --reply-na rather than inventing a value.
	Filtered bool
	// Repeats marks a site that can run more than once: it sits in a loop, or
	// in a function called from several places or used as a value. One --reply
	// answers one execution, so such a site usually needs the flag repeated -
	// and a caller told nothing would run out of answers halfway through.
	Repeats bool
	// Min and Max are a multipick's literal bounds, valid only when MinSet /
	// MaxSet is true. Knowing them lets a bad selection count fail here rather
	// than partway through the run.
	Min, Max       int64
	MinSet, MaxSet bool
	// Filter holds a pick's literal filter arguments, nil when computed. With
	// both these and Options known, rad can work out which choices survive and
	// suggest one that will actually be accepted.
	Filter []string
	// AsValue marks an interactive builtin named rather than called - `f = pick`.
	// Rad cannot see where such a value ends up, so there is no call to key an
	// answer to and --reply can never address it. It is still listed, and still
	// addressable by --reply-na, because a script that merely mentions one on a
	// path this run won't take must stay runnable.
	AsValue bool
}

// Label is what to call this site in output: the builtin's own name where there
// is one. pick_kv answers match the key the user sees rather than the value it
// returns, and calling it plain "pick" hides the distinction that matters.
func (s Site) Label() string {
	if s.Fn != "" {
		return s.Fn
	}
	return s.Kind.String()
}

// kindForBuiltin maps an interactive builtin's name to what it asks for, with a
// zero Kind for everything else. Keep in step with siteForBuiltin.
func kindForBuiltin(name string) Kind {
	switch name {
	case "input":
		return Input
	case "confirm":
		return Confirm
	case "pick", "pick_kv", "pick_from_resource":
		return Pick
	case "multipick":
		return Multipick
	}
	return 0
}

func isInteractiveBuiltin(name string) bool { return kindForBuiltin(name) != 0 }

// Find returns every site reachable on this invocation, in source order.
//
// Scoping matters as much as detection: a script with five commands would
// otherwise demand answers for prompts in the four the caller didn't run. So
// the walk starts at the top-level body plus the invoked command's callback and
// follows calls from there. cmdPath is the resolved command path, nil when the
// script has no commands or none was selected.
//
// resolved is what keeps a local named `pick` from being mistaken for the
// builtin - without it this would hard-block scripts that never prompt at all.
func Find(file *rl.SourceFile, resolved *check.Resolved, cmdPath []string) []Site {
	if file == nil || resolved == nil {
		return nil
	}

	f := &finder{
		resolved:    resolved,
		entered:     make(map[*rl.FnDef]bool),
		callCount:   make(map[*rl.FnDef]int),
		usedAsValue: make(map[*rl.FnDef]bool),
		loopReached: make(map[*rl.FnDef]bool),
		callees:     make(map[*rl.FnDef][]*rl.FnDef),
		seen:        make(map[[2]int]bool),
	}

	for _, stmt := range file.Stmts {
		f.walk(stmt)
	}
	if cmd := findCmd(file.Cmds, cmdPath); cmd != nil {
		f.walkCallback(cmd.Callback)
	}

	f.markRepeats()

	sites := make([]Site, 0, len(f.found))
	for _, fs := range f.found {
		sites = append(sites, fs.site)
	}
	sort.Slice(sites, func(i, j int) bool {
		if sites[i].Line != sites[j].Line {
			return sites[i].Line < sites[j].Line
		}
		return sites[i].Col < sites[j].Col
	})
	assignKeys(sites)
	return sites
}

// foundSite pairs a site with the function it sits in, which is what lets a
// later pass decide whether the site can run more than once.
type foundSite struct {
	site Site
	// owner is the function whose body holds the site; nil for top-level code,
	// which runs exactly once.
	owner *rl.FnDef
	// inLoop records that the site sits inside a loop in its own body.
	inLoop bool
}

type finder struct {
	resolved *check.Resolved
	entered  map[*rl.FnDef]bool
	// callCount, usedAsValue and loopReached are how a function earns the
	// "runs more than once" verdict; callees carries that verdict onward to
	// everything it in turn calls.
	callCount   map[*rl.FnDef]int
	usedAsValue map[*rl.FnDef]bool
	loopReached map[*rl.FnDef]bool
	callees     map[*rl.FnDef][]*rl.FnDef
	// callTargets marks the identifiers that name a call's target, so a bare
	// reference to a function can be told apart from calling it.
	callTargets map[*rl.Identifier]bool
	// ufcsRecv maps a UFCS call to the receiver the interpreter will prepend to
	// its arguments. Without it every positional here is off by one against the
	// signature - which for pick_kv means reading the values as the labels.
	ufcsRecv map[*rl.Call]rl.Node

	found []foundSite
	seen  map[[2]int]bool

	curFn     *rl.FnDef
	loopDepth int
}

// walk visits a node and its children, skipping function definitions. Their
// bodies are visited only when something reachable reaches them, which is what
// keeps one command from demanding another command's answers.
func (f *finder) walk(node rl.Node) {
	if node == nil {
		return
	}
	if _, isFnDef := node.(*rl.FnDef); isFnDef {
		return
	}

	f.visit(node)

	// A loop's body runs many times, but the expression it iterates over runs
	// once. Walking them at the same repetition depth would tell the caller to
	// repeat an answer for a prompt that only ever fires once.
	switch v := node.(type) {
	case *rl.ForLoop:
		f.walk(v.Iter)
		f.walkRepeatedly(v.Body...)
	case *rl.WhileLoop:
		f.walkRepeatedly(append([]rl.Node{v.Condition}, v.Body...)...) // re-tested each pass
	case *rl.ListComp:
		f.walk(v.Iter)
		f.walkRepeatedly(v.Expr, v.Condition)
	case *rl.Lambda:
		// A lambda has no name to count call sites for, and the idiomatic use -
		// handing it to map or filter - runs it once per element. Treat its body
		// as repeating rather than promise a caller one answer will do.
		f.walkRepeatedly(v.Body...)
	default:
		for _, child := range node.Children() {
			f.walk(child)
		}
	}
}

func (f *finder) walkRepeatedly(nodes ...rl.Node) {
	f.loopDepth++
	for _, n := range nodes {
		f.walk(n)
	}
	f.loopDepth--
}

func (f *finder) visit(node rl.Node) {
	switch v := node.(type) {
	case *rl.VarPath:
		f.noteUfcsReceivers(v)
	case *rl.Call:
		f.visitCall(v)
	case *rl.Identifier:
		f.visitIdent(v)
	case *rl.Shell:
		if v.IsConfirm {
			f.add(Site{Kind: ShellConfirm}, v.Span())
		}
	}
}

// noteUfcsReceivers records what `x.pick(...)` will pass as the first argument.
// Visited before the walk descends to the call itself, so visitCall can read
// the arguments in the order the builtin actually receives them.
func (f *finder) noteUfcsReceivers(path *rl.VarPath) {
	recv := path.Root
	for _, seg := range path.Segments {
		call, isCall := seg.Index.(*rl.Call)
		if seg.IsUFCS && isCall {
			if f.ufcsRecv == nil {
				f.ufcsRecv = make(map[*rl.Call]rl.Node)
			}
			f.ufcsRecv[call] = recv
			// A chained call feeds the next receiver, which is then a computed
			// expression - no literal to extract, but the offset still holds.
			recv = call
			continue
		}
		recv = nil // an intervening field or index; no plain receiver to name
	}
}

func (f *finder) visitCall(call *rl.Call) {
	ident, ok := call.Func.(*rl.Identifier)
	if !ok {
		return
	}
	// Recorded before the walk reaches this identifier as a child of the call,
	// so visitIdent can tell "calls f" from "hands f around as a value".
	f.markCallTarget(ident)

	sym := f.resolved.Uses[ident]
	if sym == nil || sym.Kind != check.SymBuiltin {
		return // a call to a named function is entered by visitIdent
	}
	if site, ok := siteForBuiltin(ident.Name, call, f.positionalArgs(call)); ok {
		// Key on the call node, not the name identifier. The interpreter
		// converts the CST separately, so the two ASTs are different objects
		// and can only be matched by position - keying both sides off the
		// same node kind is what keeps that match exact.
		f.add(site, call.Span())
	}
}

func (f *finder) markCallTarget(ident *rl.Identifier) {
	if f.callTargets == nil {
		f.callTargets = make(map[*rl.Identifier]bool)
	}
	f.callTargets[ident] = true
}

// visitIdent follows every mention of a named function, not only calls to it.
// A function assigned to a variable, passed as an argument, or returned can be
// called from anywhere afterwards, and rad cannot see where it ends up. Walking
// it anyway may demand an answer for a prompt that never fires; not walking it
// lets the script run and then die at that prompt, with side effects already
// done. Only one of those is recoverable.
func (f *finder) visitIdent(ident *rl.Identifier) {
	sym := f.resolved.Uses[ident]
	if sym == nil {
		return
	}

	// An interactive builtin handed around as a value rather than called. There
	// is no call node to key an answer to, and rad cannot see where it ends up,
	// so no --reply could ever address it. Recorded so pre-flight can say that
	// outright instead of letting the run start and die at the prompt.
	if sym.Kind == check.SymBuiltin && !f.callTargets[ident] && isInteractiveBuiltin(ident.Name) {
		f.add(Site{Kind: kindForBuiltin(ident.Name), Fn: ident.Name, AsValue: true}, ident.Span())
		return
	}

	if sym.Kind != check.SymHoistedFn {
		return
	}
	fn, ok := sym.DefNode.(*rl.FnDef)
	if !ok {
		return
	}

	if f.callTargets[ident] {
		f.callCount[fn]++
	} else {
		f.usedAsValue[fn] = true
	}
	f.enterFn(fn)
}

// enterFn walks a function body once. Identity is the definition node rather
// than the name: two nested functions can share a name, and mistaking one for
// the other reports a prompt that isn't there while missing the one that is.
func (f *finder) enterFn(fn *rl.FnDef) {
	if f.curFn != nil {
		f.callees[f.curFn] = append(f.callees[f.curFn], fn)
	}
	if f.loopDepth > 0 {
		f.loopReached[fn] = true
	}

	if f.entered[fn] {
		return // also stops recursive and mutually recursive functions
	}
	f.entered[fn] = true

	// The body is walked once, at depth zero. Whether reaching it repeats is
	// settled afterwards by markRepeats, which can see every call site - this
	// walk has only seen the first.
	prevFn, prevDepth := f.curFn, f.loopDepth
	f.curFn, f.loopDepth = fn, 0
	for _, stmt := range fn.Body {
		f.walk(stmt)
	}
	f.curFn, f.loopDepth = prevFn, prevDepth
}

// markRepeats decides which sites can run more than once. A function repeats if
// it is called from several places, handed around as a value, or reached from a
// loop - and everything it calls inherits that, however deep.
func (f *finder) markRepeats() {
	repeating := make(map[*rl.FnDef]bool)
	var queue []*rl.FnDef

	for fn := range f.entered {
		if f.callCount[fn] > 1 || f.usedAsValue[fn] || f.loopReached[fn] {
			repeating[fn] = true
			queue = append(queue, fn)
		}
	}
	for len(queue) > 0 {
		fn := queue[0]
		queue = queue[1:]
		for _, callee := range f.callees[fn] {
			if !repeating[callee] {
				repeating[callee] = true
				queue = append(queue, callee)
			}
		}
	}

	for i := range f.found {
		f.found[i].site.Repeats = f.found[i].inLoop || repeating[f.found[i].owner]
	}
}

func (f *finder) walkCallback(cb *rl.CmdCallback) {
	if cb == nil {
		return // a namespace command has no body of its own
	}
	if cb.Lambda != nil {
		// The body directly, not the lambda node: a command callback runs once,
		// unlike the lambdas handed to map and filter that walk() flags.
		for _, stmt := range cb.Lambda.Body {
			f.walk(stmt)
		}
		return
	}
	if cb.Identifier != nil {
		// `calls do_thing` invokes it exactly once. Without this it reads as a
		// function handed around as a value, and every prompt inside it gets
		// wrongly flagged as repeating.
		f.markCallTarget(cb.Identifier)
		f.visitIdent(cb.Identifier)
	}
}

func (f *finder) add(site Site, span rl.Span) {
	site.Line = span.StartRow + 1
	site.Col = span.StartCol + 1

	pos := [2]int{site.Line, site.Col}
	if f.seen[pos] {
		return
	}
	f.seen[pos] = true

	f.found = append(f.found, foundSite{site: site, owner: f.curFn, inLoop: f.loopDepth > 0})
}

// siteForBuiltin maps a call to the interactive builtin it invokes. Argument
// positions follow the signatures in docs/funcs/, which are the type checker's
// source of truth - keep them in step if a signature changes. args is already
// normalized for UFCS, so index 0 is the builtin's first parameter either way.
func siteForBuiltin(name string, call *rl.Call, args []rl.Node) (Site, bool) {
	site := Site{Fn: name}

	// Both take their prompt positionally, but rad accepts any positional by
	// name, and a caller who wrote it that way still deserves to see it. The
	// default applies only when no prompt was passed at all - one that was
	// passed but is built at runtime stays unknown rather than being reported
	// as something the user will never see.
	promptArg := func(idx int, dflt string) string {
		// idx < len(args), not hasPositional: a UFCS receiver rad can't name
		// still occupies the slot, and reporting the signature default there
		// would show a prompt the user is never going to see.
		if idx < len(args) {
			s, _ := literalString(args[idx])
			return s
		}
		if arg, ok := namedArg(call, "prompt"); ok {
			s, _ := literalString(arg)
			return s
		}
		return dflt
	}

	switch name {
	case "input":
		site.Kind = Input
		site.Prompt = promptArg(0, defaultInputPrompt)
		if arg, ok := namedArg(call, "secret"); ok {
			site.Secret, _ = literalBool(arg)
		}
	case "confirm":
		site.Kind = Confirm
		site.Prompt = promptArg(0, defaultConfirmPrompt)
	case "pick":
		site.Kind = Pick
		site.Options, _ = positionalStrings(args, 0)
		site.Filtered = hasPositional(args, 1)
		site.Filter = literalFilter(args, 1)
		site.Prompt, _ = namedString(call, "prompt")
	case "pick_kv":
		site.Kind = Pick
		site.Fn = "pick_kv"
		// The keys are what the user sees and therefore what an answer matches;
		// the values are what the call returns.
		site.Options, _ = positionalStrings(args, 0)
		site.Filtered = hasPositional(args, 2)
		site.Filter = literalFilter(args, 2)
		site.Prompt, _ = namedString(call, "prompt")
	case "pick_from_resource":
		site.Kind = Pick
		// Options live in a resource file whose path may itself be computed, so
		// they are never known here.
		site.Filtered = hasPositional(args, 1)
		site.Prompt, _ = namedString(call, "prompt")
	case "multipick":
		site.Kind = Multipick
		site.Options, _ = positionalStrings(args, 0)
		site.Prompt, _ = namedString(call, "prompt")
		site.Min, site.MinSet = namedInt(call, "min")
		site.Max, site.MaxSet = namedInt(call, "max")
	default:
		return Site{}, false
	}

	return site, true
}

// These mirror the defaults in docs/funcs/input.md and confirm.md. A call that
// takes them is not "built at runtime" - rad knows exactly what it will show.
const (
	defaultInputPrompt   = "> "
	defaultConfirmPrompt = "Confirm? [Y/n] > "
)

// positionalArgs returns the call's positional arguments as the builtin will
// receive them, prepending the receiver for a UFCS call the way the
// interpreter does.
func (f *finder) positionalArgs(call *rl.Call) []rl.Node {
	recv, isUfcs := f.ufcsRecv[call]
	if !isUfcs {
		return call.Args
	}
	// recv is nil when the receiver is a field or index rather than a plain
	// expression. The slot still has to exist, or every later index shifts.
	return append([]rl.Node{recv}, call.Args...)
}

// assignKeys gives every site its command-line address, widening to line.col
// only for the lines that need it.
func assignKeys(sites []Site) {
	perLine := make(map[int]int, len(sites))
	for _, s := range sites {
		perLine[s.Line]++
	}
	for i := range sites {
		if perLine[sites[i].Line] > 1 {
			sites[i].Key = fmt.Sprintf("%d.%d", sites[i].Line, sites[i].Col)
		} else {
			sites[i].Key = strconv.Itoa(sites[i].Line)
		}
	}
}

func findCmd(cmds []*rl.CmdBlock, path []string) *rl.CmdBlock {
	var found *rl.CmdBlock
	for _, name := range path {
		next := (*rl.CmdBlock)(nil)
		for _, cmd := range cmds {
			if cmd.Name == name {
				next = cmd
				break
			}
		}
		if next == nil {
			return nil
		}
		found, cmds = next, next.SubCmds
	}
	return found
}

func hasPositional(args []rl.Node, idx int) bool {
	return idx < len(args) && args[idx] != nil
}

func positionalString(args []rl.Node, idx int) (string, bool) {
	if !hasPositional(args, idx) {
		return "", false
	}
	return literalString(args[idx])
}

func positionalStrings(args []rl.Node, idx int) ([]string, bool) {
	if !hasPositional(args, idx) {
		return nil, false
	}
	return literalStrings(args[idx])
}

func namedArg(call *rl.Call, name string) (rl.Node, bool) {
	for _, arg := range call.NamedArgs {
		if arg.Name == name {
			return arg.Value, true
		}
	}
	return nil, false
}

func namedString(call *rl.Call, name string) (string, bool) {
	arg, ok := namedArg(call, name)
	if !ok {
		return "", false
	}
	return literalString(arg)
}

// literalString accepts only non-interpolated strings: an interpolated one
// depends on runtime state, so reporting its source text would show the caller
// something the user would never see.
func literalString(node rl.Node) (string, bool) {
	lit, ok := node.(*rl.LitString)
	if !ok || !lit.Simple {
		return "", false
	}
	return lit.Value, true
}

func literalStrings(node rl.Node) ([]string, bool) {
	list, ok := node.(*rl.LitList)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(list.Elements))
	for _, el := range list.Elements {
		s, ok := literalString(el)
		if !ok {
			return nil, false // one computed entry makes the whole list unknown
		}
		out = append(out, s)
	}
	return out, true
}

func literalBool(node rl.Node) (bool, bool) {
	lit, ok := node.(*rl.LitBool)
	if !ok {
		return false, false
	}
	return lit.Value, true
}

// literalFilter reads a pick's filter argument, which the signature allows as
// either one string or a list of them. nil when it is computed at runtime.
func literalFilter(args []rl.Node, idx int) []string {
	if !hasPositional(args, idx) {
		return nil
	}
	if one, ok := literalString(args[idx]); ok {
		return []string{one}
	}
	many, _ := literalStrings(args[idx])
	return many
}

func namedInt(call *rl.Call, name string) (int64, bool) {
	arg, ok := namedArg(call, name)
	if !ok {
		return 0, false
	}
	lit, ok := arg.(*rl.LitInt)
	if !ok {
		return 0, false
	}
	return lit.Value, true
}
