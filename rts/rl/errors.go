package rl

import "fmt"

type Error string

// todo can we avoid 5 digits?

// note to reader: I am currently very inconsistently applying these errors.
// still debating if we should them, feel free to ignore if you're implementing something.
// Note: when adding here, update the reference! docs-web/docs/reference/errors.md
const (
	// 1xxxx Syntax Errors
	ErrInvalidSyntax       Error = "10001"
	ErrMissingColon        Error = "10002"
	ErrMissingIdentifier   Error = "10003"
	ErrMissingExpression   Error = "10004"
	ErrMissingCloseParen   Error = "10005"
	ErrMissingCloseBracket Error = "10006"
	ErrMissingCloseBrace   Error = "10007"
	ErrReservedKeyword     Error = "10008"
	ErrUnexpectedToken     Error = "10009"

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

	// 3xxxx Type Errors
	ErrTypeMismatch       Error = "30001"
	ErrInvalidTypeForOp   Error = "30002"
	ErrCannotFormat       Error = "30003"
	ErrCannotIndex        Error = "30004"
	ErrCannotAssign       Error = "30005"
	ErrInvalidArgType     Error = "30006"
	ErrWrongArgCount      Error = "30007"
	ErrCannotCompare      Error = "30008"
	ErrCannotConvert      Error = "30009"

	// 4xxxx Validation Errors
	ErrScientificNotationNotWholeNumber Error = "40001"
	ErrHoistedFunctionShadowsArgument   Error = "40002"
	ErrUnknownFunction                  Error = "40003"
)

func (e Error) String() string {
	return fmt.Sprintf("RAD%s", string(e))
}
