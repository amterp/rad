package core

import ts "github.com/tree-sitter/go-tree-sitter"

type DeferBlock struct {
	DeferNode  *ts.Node
	StmtNodes  []ts.Node
	IsErrDefer bool
}

func NewDeferBlock(i *Interpreter, deferKeywordNode *ts.Node, stmtNodes []ts.Node) *DeferBlock {
	deferKeywordStr := i.sd.Src[deferKeywordNode.StartByte():deferKeywordNode.EndByte()]
	return &DeferBlock{
		DeferNode:  deferKeywordNode,
		StmtNodes:  stmtNodes,
		IsErrDefer: deferKeywordStr == "errdefer",
	}
}

func (i *Interpreter) RegisterWithExit() {
	existing := RExit
	exiting := false
	codeToExitWith := 0
	RExit = func(code int) {
		if exiting {
			// we're already exiting. if we're here again, it's probably because one of the deferred
			// statements is calling exit again (perhaps because it failed). we should keep running
			// all the deferred statements, however, and *then* exit.
			// therefore, we panic here in order to send the stack back up to where the deferred statement is being
			// invoked in the interpreter, which should be wrapped in a recover() block to catch, maybe log, and move on.
			if codeToExitWith == 0 {
				codeToExitWith = code
			}
			return
		}
		exiting = true
		codeToExitWith = code
		// todo gets executed *after* any error is printed (if error), should delay error print until after (i think?)
		i.executeDeferBlocks(code)
		existing(codeToExitWith)
	}
}

func (i *Interpreter) executeDeferBlocks(errCode int) {
	// execute backwards (LIFO)
	for j := len(i.deferBlocks) - 1; j >= 0; j-- {
		deferBlock := i.deferBlocks[j]

		if errCode == 0 && deferBlock.IsErrDefer {
			continue
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					// we only debug log. we expect the error that occurred to already have been logged.
					// we might also be here only because a deferred statement invoked a clean exit, for example, so
					// this is arguably also sometimes just standard flow.
					RP.RadDebugf("Recovered from panic in deferred statement: %v", r)
				}
			}()
			i.runBlock(deferBlock.StmtNodes)
		}()
	}
}
