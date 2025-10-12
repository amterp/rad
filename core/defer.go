package core

import ts "github.com/tree-sitter/go-tree-sitter"

type DeferBlock struct {
	DeferNode  *ts.Node
	StmtNodes  []ts.Node
	IsErrDefer bool
}

func NewDeferBlock(i *Interpreter, deferKeywordNode *ts.Node, stmtNodes []ts.Node) *DeferBlock {
	deferKeywordStr := i.GetSrcForNode(deferKeywordNode)
	return &DeferBlock{
		DeferNode:  deferKeywordNode,
		StmtNodes:  stmtNodes,
		IsErrDefer: deferKeywordStr == "errdefer",
	}
}

func (i *Interpreter) RegisterWithExit() {
	RExit.SetExecuteDeferredStmtsFunc(i.executeDeferBlocks)
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
