package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/amterp/rad/rts/rl"
)

// Ordered longest-first so e.g. YYYY is replaced before a hypothetical YY.
var dateFormatTokens = []struct {
	token    string
	goLayout string
}{
	{"YYYY", "2006"},
	{"MM", "01"},
	{"DD", "02"},
	{"HH", "15"},
	{"mm", "04"},
	{"ss", "05"},
}

// Auto-detect formats tried in order. More specific formats come first so we
// don't accidentally match a datetime string as just a date. hasTz marks
// formats that capture timezone from the input string.
//
// Go's .999999999 layout gracefully accepts absent fractional seconds, so
// each entry covers both fractional and non-fractional variants.
var autoDetectFormats = []struct {
	layout string
	hasTz  bool
}{
	{time.RFC3339Nano, true},                      // 2006-01-02T15:04:05[.nnn]Z or +HH:MM
	{"2006-01-02 15:04:05.999999999Z07:00", true}, // space separator with tz offset
	{"2006-01-02T15:04:05.999999999", false},      // T separator, no tz
	{"2006-01-02 15:04:05.999999999", false},      // space separator, no tz
	{"2006-01-02", false},                         // date only
}

func convertFormatToGoLayout(format string) string {
	result := format
	for _, tok := range dateFormatTokens {
		result = strings.ReplaceAll(result, tok.token, tok.goLayout)
	}
	return result
}

var FuncParseDate = BuiltInFunc{
	Name: FUNC_PARSE_DATE,
	Execute: func(f FuncInvocation) RadValue {
		dateStr := f.GetStr("_date").Plain()
		formatArg := f.GetArg("format")
		tz := f.GetStr("tz").Plain()

		// Resolve output timezone
		var location *time.Location
		if tz == "local" {
			location = RClock.Local()
		} else {
			var err error
			location, err = time.LoadLocation(tz)
			if err != nil {
				return f.ReturnErrf(rl.ErrInvalidTimeZone, "Invalid time zone '%s'", tz)
			}
		}

		if dateStr == "" {
			return f.Return(NewErrorStrf("Cannot parse an empty date string").SetCode(rl.ErrParseDate))
		}

		var parsedTime time.Time

		if !formatArg.IsNull() {
			// Explicit format: convert tokens to Go layout, parse in target tz
			format := formatArg.RequireStr(f.i, f.callNode).Plain()
			if format == "" {
				return f.Return(NewErrorStrf("Cannot parse date with an empty format string").SetCode(rl.ErrParseDate))
			}
			goLayout := convertFormatToGoLayout(format)

			t, err := time.ParseInLocation(goLayout, dateStr, location)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to parse date %q with format %q", dateStr, format)
				return f.Return(NewErrorStrf(errMsg).SetCode(rl.ErrParseDate))
			}
			parsedTime = t
		} else {
			// Auto-detect: try known unambiguous formats
			var parsed bool
			for _, af := range autoDetectFormats {
				if af.hasTz {
					t, err := time.Parse(af.layout, dateStr)
					if err == nil {
						parsedTime = t.In(location)
						parsed = true
						break
					}
				} else {
					t, err := time.ParseInLocation(af.layout, dateStr, location)
					if err == nil {
						parsedTime = t
						parsed = true
						break
					}
				}
			}

			if !parsed {
				errMsg := fmt.Sprintf(
					"Failed to parse date %q. Supported formats: YYYY-MM-DD, YYYY-MM-DDTHH:mm:ss, "+
						"YYYY-MM-DD HH:mm:ss (with optional timezone offset and fractional seconds). "+
						"Use 'format' to specify a custom format, e.g. parse_date(%q, format=\"DD/MM/YYYY\").",
					dateStr, dateStr,
				)
				return f.Return(NewErrorStrf(errMsg).SetCode(rl.ErrParseDate))
			}
		}

		return f.Return(NewTimeMap(parsedTime))
	},
}
