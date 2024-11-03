package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	SLEEP_TITLE = "title"
)

func runSleep(i *MainInterpreter, sleepToken Token, args []interface{}, namedArgs map[string]interface{}) {
	if len(args) != 1 {
		i.error(sleepToken, SLEEP+"() takes exactly one positional argument")
	}

	validateExpectedNamedArgs(i, sleepToken, []string{SLEEP_TITLE}, namedArgs)
	parsedArgs := parseSleepArgs(namedArgs)

	switch coerced := args[0].(type) {
	case string:
		durStr := strings.Replace(coerced, " ", "", -1)

		floatVal, err := strconv.ParseFloat(durStr, 64)
		if err == nil {
			sleep(i, sleepToken, time.Duration(floatVal*1000)*time.Millisecond, parsedArgs)
			return
		}

		dur, err := time.ParseDuration(durStr)
		if err == nil {
			sleep(i, sleepToken, dur, parsedArgs)
			return
		}

		i.error(sleepToken, SLEEP+fmt.Sprintf("Invalid string argument: '%s'", args[0]))
	case int64:
		sleep(i, sleepToken, time.Duration(coerced)*time.Second, parsedArgs)
	case float64:
		sleep(i, sleepToken, time.Duration(coerced*1000)*time.Millisecond, parsedArgs)
	default:
		i.error(sleepToken, SLEEP+fmt.Sprintf("() takes an int, float, or string, got %s", TypeAsString(args[0])))
	}
}

func sleep(i *MainInterpreter, sleepToken Token, dur time.Duration, args SleepNamedArgs) {
	if dur < 0 {
		i.error(sleepToken, SLEEP+fmt.Sprintf("() cannot take a negative duration: %s", dur.String()))
	}

	if args.Title != "" {
		RP.Print(args.Title + "\n")
	}
	RSleep(dur)
}

func parseSleepArgs(args map[string]interface{}) SleepNamedArgs {
	parsedArgs := SleepNamedArgs{
		Title: "",
	}

	if title, ok := args[SLEEP_TITLE]; ok {
		parsedArgs.Title = title.(string)
	}

	return parsedArgs
}

type SleepNamedArgs struct {
	Title string
}
