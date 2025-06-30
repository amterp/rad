package rl

import "fmt"

type Error string

// todo can we avoid 5 digits?

// note to reader: I am currently very inconsistently applying these errors.
// still debating if we should them, feel free to ignore if you're implementing something.
// Note: when adding here, updating the reference!! " docs/reference/errors.md "
const (
	// 1xxxx Syntax Errors
	ErrInvalidSyntax Error = "10001"

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

	// 3xxxx Type Errors?

	// 4xxxx Validation Errors?
)

func (e Error) String() string {
	return fmt.Sprintf("RAD%s", string(e))
}
