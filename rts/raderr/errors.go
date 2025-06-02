package raderr

type Error string

// todo can we avoid 5 digits?

// note to reader: I am currently very inconsistently applying these errors.
// still debating if we should them, feel free to ignore if you're implementing something.
// Note: when adding here, updating the reference!! " docs/reference/errors.md "
const (
	// RAD1xxxx Syntax Errors
	ErrInvalidSyntax Error = "RAD10001"

	// RAD2xxxx Runtime Errors
	ErrParseIntFailed   Error = "RAD20001"
	ErrParseFloatFailed       = "RAD20002"
	ErrFileRead               = "RAD20003"
	ErrFileNoPermission       = "RAD20004"
	ErrFileNoExist            = "RAD20005"
	ErrFileWrite              = "RAD20006"
	ErrAmbiguousEpoch         = "RAD20007"
	ErrInvalidTimeUnit        = "RAD20008"
	ErrInvalidTimeZone        = "RAD20009"

	// RAD3xxxx Type Errors?

	// RAD4xxxx Validation Errors?
)

func (e Error) String() string {
	return string(e)
}
