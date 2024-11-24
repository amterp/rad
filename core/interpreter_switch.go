package core

import (
	"github.com/scylladb/go-set/strset"
)

// todo implement fuzzy matching

type SwitchInterpreter struct {
	i          *MainInterpreter
	invocation *switchInvocation
}

func NewSwitchInterpreter(i *MainInterpreter) *SwitchInterpreter {
	return &SwitchInterpreter{i: i}
}

func (s SwitchInterpreter) RunBlock(block SwitchBlock) []interface{} {
	var discriminator Token
	if block.Discriminator != nil {
		token := *block.Discriminator
		discriminator = token
	}
	s.invocation = &switchInvocation{
		si:            &s,
		blockToken:    block.SwitchToken,
		discriminator: discriminator,
		cases:         []switchCaseInstance{},
	}

	for _, stmt := range block.Stmts {
		stmt.Accept(s)
	}

	return s.invocation.execute()
}

func (s SwitchInterpreter) RunAssignment(assignment SwitchAssignment) {
	outputs := s.RunBlock(assignment.Block)
	identifiers := assignment.Identifiers

	for i, output := range outputs {
		s.i.env.SetAndImplyType(identifiers[i], output)
	}
}

func (s SwitchInterpreter) VisitSwitchCaseSwitchStmt(switchCase SwitchCase) {
	var keys []string
	for _, key := range switchCase.Keys {
		keys = append(keys, key.Value[len(key.Value)-1].FullStringLiteral)
	}

	var values []Expr
	for _, expr := range switchCase.Values {
		values = append(values, expr)
	}

	s.invocation.cases = append(s.invocation.cases, switchCaseInstance{
		keys:   keys,
		values: values,
	})
}

func (s SwitchInterpreter) VisitSwitchDefaultSwitchStmt(switchDefault SwitchDefault) {
	var values []Expr
	for _, expr := range switchDefault.Values {
		values = append(values, expr)
	}
	s.invocation.defaultExprs = &values
}

// == switchInvocation ==

type switchCaseInstance struct {
	keys   []string
	values []Expr
}

type switchInvocation struct {
	si            *SwitchInterpreter
	blockToken    Token
	discriminator Token
	cases         []switchCaseInstance
	defaultExprs  *[]Expr
}

func (s *switchInvocation) execute() []interface{} {
	if s.discriminator == nil {
		return s.decideBasedOnStringInterpolation()
	} else {
		return s.decideBasedOnKeys()
	}
}

func (s *switchInvocation) decideBasedOnKeys() []interface{} {
	discrValueLiteral := s.si.i.env.GetByToken(s.discriminator)
	discrValueString := ToPrintable(discrValueLiteral)

	var exprs []Expr
	for _, instance := range s.cases {
		for _, key := range instance.keys {
			if key == discrValueString {
				exprs = instance.values
				break
			}
		}
	}

	if exprs == nil {
		if s.defaultExprs != nil {
			return s.evaluateExpressions(*s.defaultExprs)
		}
		s.si.i.error(s.blockToken, "No cases matched, and no default case provided")
	}

	return s.evaluateExpressions(exprs)
}

// algo:
// 1. for each option, evaluate the number of unique variables it references, and the total number
// 2. eliminate options that reference undefined variables
// 3. choose the option that references the most unique variables
// 4. tie-break with the option that references the most total variables
// 5. if there is still a tie, error
func (s *switchInvocation) decideBasedOnStringInterpolation() []interface{} {
	s.si.i.LiteralI.ShouldInterpolate = false

	// count unique and total variable counts for every case
	var numTotalReferencedVarsByCaseIndex []int
	var numUniqueReferencedVarsByCaseIndex []int
	for _, instance := range s.cases {
		var numTotalForCase = 0
		uniqueReferencedVars := strset.New()
		for _, expr := range instance.values {
			referencedVars := extractVariables(s.si.i, s.blockToken, expr)
			numTotalForCase += len(referencedVars)
			uniqueReferencedVars.Add(referencedVars...)
		}
		uniqueReferencedVarsList := uniqueReferencedVars.List()
		numUniqueForCase := len(uniqueReferencedVarsList)
		for _, varName := range uniqueReferencedVarsList {
			if !s.si.i.env.Exists(varName) {
				// todo hacky
				numTotalForCase -= 9999
				numUniqueForCase -= 9999
			}
		}
		numTotalReferencedVarsByCaseIndex = append(numTotalReferencedVarsByCaseIndex, numTotalForCase)
		numUniqueReferencedVarsByCaseIndex = append(numUniqueReferencedVarsByCaseIndex, numUniqueForCase)
	}

	// choose the case with the most unique variables, tie-break with the most total variables
	var highestCaseIndex int
	highestUnique := 0
	highestTotal := 0
	equalNumAtHighest := 0
	for index, uniqueNumReferenced := range numUniqueReferencedVarsByCaseIndex {
		numTotalReferenced := numTotalReferencedVarsByCaseIndex[index]
		if uniqueNumReferenced > highestUnique ||
			(uniqueNumReferenced == highestUnique && numTotalReferenced > highestTotal) {

			highestUnique = uniqueNumReferenced
			highestTotal = numTotalReferenced
			highestCaseIndex = index
			equalNumAtHighest = 1
		} else if uniqueNumReferenced == highestUnique && numTotalReferenced == highestTotal {
			equalNumAtHighest++
		}
	}

	// error if there are multiple options that match the same number of variables
	if equalNumAtHighest > 1 {
		s.si.i.error(s.blockToken, "Ambiguous choice block, multiple options are equal matches")
	}

	if highestTotal < -1000 { // todo hacky
		s.si.i.error(s.blockToken, "No choice block option matches all its referenced variables")
	}

	s.si.i.LiteralI.ShouldInterpolate = true

	var outputs []interface{}
	for _, expr := range s.cases[highestCaseIndex].values {
		// we end up re-evaluating the winning expressions, but in theory they should be idempotent?
		outputs = append(outputs, expr.Accept(s.si.i))
	}

	return outputs
}

func (s *switchInvocation) evaluateExpressions(exprs []Expr) []interface{} {
	var outputs []interface{}
	for _, expr := range exprs {
		outputs = append(outputs, expr.Accept(s.si.i))
	}
	return outputs
}
