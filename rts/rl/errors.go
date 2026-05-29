package rl

import "fmt"

type Error string

// Tombstone rule: once an error code is in use, its number is
// reserved forever - even when retired. Renaming retired codes with
// a `_retired` suffix keeps the constant declared (so anyone reading
// an old log can grep for it) but signals at the use site that the
// code no longer fires. New codes always pick the next unused
// number in their band. Bands:
//
//	1xxxx  Syntax / parser errors    (this file's first block)
//	2xxxx  Runtime errors            (interpreter)
//	3xxxx  Type errors               (shared static + runtime)
//	4xxxx  Validation / lint errors  (static only)
//
// When adding here, also touch core/error_docs/<code>.md and the
// reference at docs-web/docs/reference/errors.md.
const (
	// 1xxxx Syntax Errors
	//
	// Most of the MISSING-node based codes (10003-10007, 10010-10017,
	// 10019) are unreachable in practice: tree-sitter's error recovery
	// emits ERROR nodes rather than MISSING nodes for these shapes, so
	// the dispatch in error_messages.go falls back to RAD10001 /
	// RAD10009. They're left in place as `_retired` markers so the
	// numbers stay reserved per the tombstone rule, and so old logs
	// remain greppable. The error_docs/<code>.md tombstones point
	// readers at the codes that do fire.
	ErrInvalidSyntax               Error = "10001"
	ErrMissingColon                Error = "10002"
	ErrMissingIdentifier_retired   Error = "10003"
	ErrMissingExpression_retired   Error = "10004"
	ErrMissingCloseParen_retired   Error = "10005"
	ErrMissingCloseBracket_retired Error = "10006"
	ErrMissingCloseBrace_retired   Error = "10007"
	ErrReservedKeyword             Error = "10008"
	ErrUnexpectedToken             Error = "10009"
	ErrMissingOpenParen_retired    Error = "10010"
	ErrMissingOpenBracket_retired  Error = "10011"
	ErrMissingOpenBrace_retired    Error = "10012"
	ErrMissingComma_retired        Error = "10013"
	ErrMissingEquals_retired       Error = "10014"
	ErrMissingArrow_retired        Error = "10015"
	ErrMissingType_retired         Error = "10016"
	ErrMissingNewline_retired      Error = "10017"
	ErrMissingIndent               Error = "10018"
	ErrUnexpectedIndent_retired    Error = "10019"
	ErrUnterminatedString          Error = "10020"
	ErrMissingOperator             Error = "10021"
	ErrKeywordMisuse               Error = "10022"

	// 2xxxx Runtime Errors
	ErrGenericRuntime       Error = "20000"
	ErrParseIntFailed       Error = "20001"
	ErrParseFloatFailed           = "20002"
	ErrFileRead                   = "20003"
	ErrFileNoPermission           = "20004"
	ErrFileNoExist                = "20005"
	ErrFileWrite                  = "20006"
	ErrAmbiguousEpoch             = "20007"
	ErrInvalidTimeUnit            = "20008"
	ErrInvalidTimeZone            = "20009"
	ErrUserInput                  = "20010"
	ErrParseJson                  = "20011"
	ErrBugTypeCheck               = "20012"
	ErrFileWalk                   = "20013"
	ErrMutualExclArgs             = "20014"
	ErrZipStrict                  = "20015"
	ErrCast                       = "20016"
	ErrNumInvalidRange            = "20017"
	ErrEmptyList                  = "20018"
	ErrArgsContradict             = "20019"
	ErrFid                        = "20020"
	ErrDecode                     = "20021"
	ErrNoStashId                  = "20022"
	ErrSleepStr                   = "20023"
	ErrInvalidRegex               = "20024"
	ErrColorizeValNotInEnum       = "20025"
	ErrStdinRead                  = "20026"
	ErrInvalidCheckDuration       = "20027"
	ErrUndefinedVariable          = "20028"
	ErrIndexOutOfBounds           = "20029"
	ErrBreakOutsideLoop           = "20030"
	ErrContinueOutsideLoop        = "20031"
	ErrNotIterable                = "20032"
	ErrUnpackMismatch             = "20033"
	ErrSwitchNoMatch              = "20034"
	ErrSwitchMultipleMatch        = "20035"
	ErrDivisionByZero             = "20036"
	ErrNegativeIndex              = "20037"
	ErrVoidValue                  = "20038"
	ErrUnsupportedOperation       = "20039"
	ErrAssertionFailed            = "20040"
	ErrKeyNotFound                = "20041"
	ErrInternalBug                = "20042"
	ErrParseDuration              = "20043"
	ErrParseDate                  = "20044"

	// 3xxxx Type Errors
	ErrTypeMismatch              Error = "30001"
	ErrInvalidTypeForOp          Error = "30002"
	ErrCannotFormat              Error = "30003"
	ErrCannotIndex               Error = "30004"
	ErrCannotAssign              Error = "30005"
	ErrInvalidArgType            Error = "30006"
	ErrWrongArgCount             Error = "30007"
	ErrCannotCompare             Error = "30008"
	ErrCannotConvert             Error = "30009"
	ErrCollectionElementMismatch Error = "30010"

	// 4xxxx Validation Errors
	ErrScientificNotationNotWholeNumber Error = "40001"
	ErrHoistedFunctionShadowsArgument   Error = "40002"
	ErrUnknownFunction                  Error = "40003"
	ErrReturnOutsideFunction            Error = "40004"
	ErrYieldOutsideFunction             Error = "40005"
	ErrInvalidAssignmentTarget          Error = "40006"
	ErrRadOptionNoEffect                Error = "40007"
	ErrDeprecatedBlockKeyword           Error = "40008"
	ErrDuplicateParameter               Error = "40009"
	ErrNonExhaustiveSwitch              Error = "40010"
	ErrDuplicateTypedDeclaration        Error = "40011"
	ErrUnreachableCase                  Error = "40012"
	ErrCaseKeyNotInDiscriminantType     Error = "40013"
)

func (e Error) String() string {
	return fmt.Sprintf("RAD%s", string(e))
}
