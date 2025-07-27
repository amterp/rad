package ra

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ExitFunc is the interface for exiting the program
type ExitFunc func(int)

// StderrWriter is the interface for writing to stderr
type StderrWriter interface {
	Write([]byte) (int, error)
}

var osExit ExitFunc = os.Exit
var stderrWriter StderrWriter = os.Stderr

type Cmd struct {
	name          string
	description   string
	flags         map[string]any // flag name -> flag itself (either a Flag[T] or SliceFlag[T])
	positional    []string       // positional flags, i.e. flags that are positional args
	nonPositional []string       // non-positional flags, i.e. flags that are only named
	globalFlags   []string       // flags that will be applied to all subcommands
	subCmds       map[string]*Cmd
	shortToName   map[string]string // short flag -> full name mapping

	// options
	customUsage          func(bool) // if set, this function will be called to print usage instead of the default
	helpEnabled          bool       // default true automatically adds a help flag
	excludeNameFromUsage bool       // if true, this command will not be included in usage output

	// state post-parse
	used             *bool           // after parsing, whether this command was invoked
	configured       map[string]bool // specified flags from flags.
	unknownArgs      []string        // unknown args when ignoreUnknown is true
	lastVariadicFlag string          // last variadic flag that was used
	sawFlag          bool            // true if we've seen a flag since the last variadic
}

func NewCmd(name string) *Cmd {
	c := &Cmd{
		name:        name,
		flags:       make(map[string]any),
		positional:  []string{},
		subCmds:     make(map[string]*Cmd),
		configured:  make(map[string]bool),
		helpEnabled: true,
		shortToName: make(map[string]string),
	}

	return c
}

func (c *Cmd) SetDescription(desc string) *Cmd {
	c.description = desc
	return c
}

func (c *Cmd) SetCustomUsage(fn func(isLongHelp bool)) *Cmd {
	c.customUsage = fn
	return c
}

func (c *Cmd) SetHelpEnabled(enable bool) *Cmd {
	c.helpEnabled = enable
	return c
}

func (c *Cmd) SetExcludeNameFromUsage(exclude bool) *Cmd {
	c.excludeNameFromUsage = exclude
	return c
}

type parseCfg struct {
	ignoreUnknown bool
}

type ParseOpt func(*parseCfg)

func WithIgnoreUnknown(ignore bool) ParseOpt {
	return func(c *parseCfg) {
		c.ignoreUnknown = ignore
	}
}

func (c *Cmd) ParseOrExit(args []string, opts ...ParseOpt) {
	err := c.parse(args, opts...)
	if err != nil {
		fmt.Fprintln(stderrWriter, err.Error())
		fmt.Fprintln(stderrWriter, c.GenerateLongUsage())
		osExit(1)
	}
}

func (c *Cmd) ParseOrError(args []string, opts ...ParseOpt) error {
	return c.parse(args, opts...)
}

func (c *Cmd) parse(args []string, opts ...ParseOpt) error {
	return c.parseWithPreserveState(args, false, opts...)
}

func (c *Cmd) parseWithPreserveState(args []string, preserveConfigured bool, opts ...ParseOpt) error {
	cfg := &parseCfg{}
	for _, opt := range opts {
		opt(cfg)
	}

	// reset state in case this is called multiple times
	if !preserveConfigured {
		c.configured = make(map[string]bool)
	}
	c.unknownArgs = []string{}
	c.lastVariadicFlag = ""
	c.sawFlag = false

	// Add help flags if enabled
	if c.helpEnabled {
		if _, exists := c.flags["help"]; !exists {
			NewBool(
				"help",
			).SetShort("h").
				SetUsage("Print usage string.").
				SetOptional(true).
				Register(c, WithGlobal(true))
		}
	}

	// Set defaults first
	if err := c.setDefaults(); err != nil {
		return err
	}

	// Check for help flags (only if helpEnabled is true)
	if c.helpEnabled {
		for _, arg := range args {
			if arg == "--help" {
				if c.customUsage != nil {
					c.customUsage(true)
				} else {
					fmt.Fprint(stderrWriter, c.GenerateLongUsage())
				}
				osExit(0)
			}
			if arg == "-h" {
				if c.customUsage != nil {
					c.customUsage(false)
				} else {
					fmt.Fprint(stderrWriter, c.GenerateShortUsage())
				}
				osExit(0)
			}
		}
	}

	// Check if we have number shorts mode
	numberShortsMode := c.hasNumberShorts()

	// Parse arguments
	i := 0
	for i < len(args) {
		arg := args[i]

		// Check for subcommand first
		if !strings.HasPrefix(arg, "-") {
			if subCmd, exists := c.subCmds[arg]; exists {
				*subCmd.used = true
				// Apply global flags to subcommand before parsing
				if err := c.applyGlobalFlags(subCmd); err != nil {
					return err
				}
				// Apply global configured state before parsing
				for _, globalFlagName := range c.globalFlags {
					if c.configured[globalFlagName] {
						subCmd.configured[globalFlagName] = true
					}
				}
				return subCmd.parseWithPreserveState(args[i+1:], true, opts...)
			}
		}

		// Handle flags
		if strings.HasPrefix(arg, "-") {
			consumed, err := c.parseFlag(args, i, numberShortsMode)
			if err != nil {
				if err.Error() == "not a flag: "+arg {
					// This is a negative number, treat as positional
					if err := c.assignPositional(arg); err != nil {
						if cfg.ignoreUnknown {
							c.unknownArgs = append(c.unknownArgs, arg)
						} else {
							return err
						}
					}
					i++
					continue
				}
				if cfg.ignoreUnknown {
					c.unknownArgs = append(c.unknownArgs, arg)
					i++
					continue
				}
				return err
			}
			c.sawFlag = true
			c.lastVariadicFlag = "" // Reset variadic state when we see a flag
			i += consumed
		} else {
			// Handle positional argument
			if err := c.assignPositional(arg); err != nil {
				if cfg.ignoreUnknown {
					c.unknownArgs = append(c.unknownArgs, arg)
				} else {
					return err
				}
			}
			i++
		}
	}

	// Validate required flags
	return c.validateRequired()
}

func (c *Cmd) setDefaults() error {
	for _, flag := range c.flags {
		switch f := flag.(type) {
		case *BoolFlag:
			if f.Default != nil && !c.configured[f.Name] {
				*f.Value = *f.Default
			}
		case *StringFlag:
			if f.Default != nil && !c.configured[f.Name] {
				*f.Value = *f.Default
			}
		case *IntFlag:
			if f.Default != nil && !c.configured[f.Name] {
				*f.Value = *f.Default
			}
		case *Int64Flag:
			if f.Default != nil && !c.configured[f.Name] {
				*f.Value = *f.Default
			}
		case *Float64Flag:
			if f.Default != nil && !c.configured[f.Name] {
				*f.Value = *f.Default
			}
		case *StringSliceFlag:
			if !c.configured[f.Name] {
				if f.Default != nil {
					*f.Value = *f.Default
				} else {
					*f.Value = []string{}
				}
			}
		case *IntSliceFlag:
			if !c.configured[f.Name] {
				if f.Default != nil {
					*f.Value = *f.Default
				} else {
					*f.Value = []int{}
				}
			}
		case *Int64SliceFlag:
			if !c.configured[f.Name] {
				if f.Default != nil {
					*f.Value = *f.Default
				} else {
					*f.Value = []int64{}
				}
			}
		case *Float64SliceFlag:
			if !c.configured[f.Name] {
				if f.Default != nil {
					*f.Value = *f.Default
				} else {
					*f.Value = []float64{}
				}
			}
		case *BoolSliceFlag:
			if !c.configured[f.Name] {
				if f.Default != nil {
					*f.Value = *f.Default
				} else {
					*f.Value = []bool{}
				}
			}
		}
	}
	return nil
}

func (c *Cmd) hasNumberShorts() bool {
	for _, flag := range c.flags {
		var short string
		switch f := flag.(type) {
		case *IntFlag:
			short = f.Short
		case *Int64Flag:
			short = f.Short
		case *Float64Flag:
			short = f.Short
		case *StringFlag:
			short = f.Short
		case *BoolFlag:
			short = f.Short
		case *StringSliceFlag:
			short = f.Short
		case *IntSliceFlag:
			short = f.Short
		case *Int64SliceFlag:
			short = f.Short
		case *Float64SliceFlag:
			short = f.Short
		case *BoolSliceFlag:
			short = f.Short
		}
		if short != "" && len(short) == 1 && isDigit(short[0]) {
			return true
		}
	}
	return false
}

func (c *Cmd) parseFlag(args []string, index int, numberShortsMode bool) (int, error) {
	arg := args[index]

	if strings.HasPrefix(arg, "--") {
		// Long flag
		return c.parseLongFlag(args, index)
	} else if strings.HasPrefix(arg, "-") {
		// Short flag(s)
		return c.parseShortFlag(args, index, numberShortsMode)
	}

	return 0, fmt.Errorf("invalid flag: %s", arg)
}

func (c *Cmd) parseLongFlag(args []string, index int) (int, error) {
	arg := args[index]
	flagName := arg[2:] // remove --

	// Check for = syntax
	var value string
	var hasValue bool
	if idx := strings.Index(flagName, "="); idx != -1 {
		value = flagName[idx+1:]
		flagName = flagName[:idx]
		hasValue = true
	}

	flag, exists := c.flags[flagName]
	if !exists {
		return 0, fmt.Errorf("unknown flag: --%s", flagName)
	}

	c.configured[flagName] = true

	switch f := flag.(type) {
	case *BoolFlag:
		if hasValue {
			val, err := c.parseBoolValue(value)
			if err != nil {
				return 0, fmt.Errorf("invalid value for flag --%s: %s", flagName, err.Error())
			}
			*f.Value = val
		} else {
			*f.Value = true
		}
		return 1, nil
	case *StringFlag:
		if hasValue {
			err := c.setStringValue(f, value)
			return 1, err
		}
		if index+1 >= len(args) {
			return 0, fmt.Errorf("flag --%s requires a value", flagName)
		}
		err := c.setStringValue(f, args[index+1])
		return 2, err
	case *IntFlag:
		if hasValue {
			err := c.setIntValue(f, value)
			return 1, err
		}
		if index+1 >= len(args) {
			return 0, fmt.Errorf("flag --%s requires a value", flagName)
		}
		err := c.setIntValue(f, args[index+1])
		return 2, err
	case *Int64Flag:
		if hasValue {
			err := c.setInt64Value(f, value)
			return 1, err
		}
		if index+1 >= len(args) {
			return 0, fmt.Errorf("flag --%s requires a value", flagName)
		}
		err := c.setInt64Value(f, args[index+1])
		return 2, err
	case *Float64Flag:
		if hasValue {
			err := c.setFloat64Value(f, value)
			return 1, err
		}
		if index+1 >= len(args) {
			return 0, fmt.Errorf("flag --%s requires a value", flagName)
		}
		err := c.setFloat64Value(f, args[index+1])
		return 2, err
	case *StringSliceFlag:
		if hasValue {
			_, err := c.appendStringSliceValue(f, value)
			if err == nil && f.Variadic {
				c.lastVariadicFlag = flagName
			}
			return 1, err
		}
		consumed, err := c.parseSliceFlag(args, index, f)
		if err == nil && f.Variadic {
			c.lastVariadicFlag = flagName
		}
		return consumed, err
	case *IntSliceFlag:
		if hasValue {
			_, err := c.appendIntSliceValue(f, value)
			return 1, err
		}
		return c.parseIntSliceFlag(args, index, f)
	case *Int64SliceFlag:
		if hasValue {
			_, err := c.appendInt64SliceValue(f, value)
			return 1, err
		}
		return c.parseInt64SliceFlag(args, index, f)
	case *Float64SliceFlag:
		if hasValue {
			_, err := c.appendFloat64SliceValue(f, value)
			return 1, err
		}
		return c.parseFloat64SliceFlag(args, index, f)
	case *BoolSliceFlag:
		if hasValue {
			_, err := c.appendBoolSliceValue(f, value)
			return 1, err
		}
		return c.parseBoolSliceFlag(args, index, f)
	}

	return 0, fmt.Errorf("unsupported flag type for: %s", flagName)
}

func (c *Cmd) parseShortFlag(args []string, index int, numberShortsMode bool) (int, error) {
	arg := args[index]
	shorts := arg[1:] // remove -

	// Check if this is a negative number without number shorts mode
	if !numberShortsMode && len(shorts) > 0 && (isDigit(shorts[0]) || shorts[0] == '.') {
		// This is a negative number, treat as positional
		return 0, fmt.Errorf("not a flag: %s", arg)
	}

	// In number shorts mode, check if this is a negative number
	if numberShortsMode && len(shorts) > 0 && isDigit(shorts[0]) {
		// This is a number short flag
		if flagName, exists := c.shortToName[shorts]; exists {
			flag := c.flags[flagName]
			c.configured[flagName] = true

			switch f := flag.(type) {
			case *IntFlag:
				if len(shorts) > 1 {
					// Multiple occurrences like -aaa
					count := len(shorts)
					*f.Value = count
					return 1, nil
				}
				// Single occurrence, needs value
				if index+1 >= len(args) {
					return 0, fmt.Errorf("flag -%s requires a value", shorts)
				}
				err := c.setIntValue(f, args[index+1])
				return 2, err
			case *StringFlag:
				if index+1 >= len(args) {
					return 0, fmt.Errorf("flag -%s requires a value", shorts)
				}
				err := c.setStringValue(f, args[index+1])
				return 2, err
			}
		}
	}

	// Regular short flag processing
	consumed := 1

	// Check if all chars are the same (for int flag counting)
	if len(shorts) > 1 {
		firstChar := shorts[0]
		allSame := true
		for i := 1; i < len(shorts); i++ {
			if shorts[i] != firstChar {
				allSame = false
				break
			}
		}
		if allSame {
			// All chars are the same, check if it's an int flag
			if flagName, exists := c.shortToName[string(firstChar)]; exists {
				if flag, exists := c.flags[flagName]; exists {
					if intFlag, ok := flag.(*IntFlag); ok {
						// This is an int flag being repeated, set it to the count
						c.configured[flagName] = true
						*intFlag.Value = len(shorts)
						return 1, nil
					}
				}
			}
		}
	}

	for i, short := range shorts {
		shortStr := string(short)
		flagName, exists := c.shortToName[shortStr]
		if !exists {
			return 0, fmt.Errorf("unknown short flag: -%s", shortStr)
		}

		flag := c.flags[flagName]
		c.configured[flagName] = true

		switch f := flag.(type) {
		case *BoolFlag:
			*f.Value = true
		case *StringFlag:
			if i == len(shorts)-1 {
				// Last flag in cluster, can take value
				if index+1 >= len(args) {
					return 0, fmt.Errorf("flag -%s requires a value", shortStr)
				}
				err := c.setStringValue(f, args[index+1])
				if err != nil {
					return 0, err
				}
				consumed = 2
			} else {
				return 0, fmt.Errorf("non-bool flag -%s must be last in cluster", shortStr)
			}
		case *IntFlag:
			if i == len(shorts)-1 {
				// Last flag in cluster, can take value
				if index+1 >= len(args) {
					return 0, fmt.Errorf("flag -%s requires a value", shortStr)
				}
				err := c.setIntValue(f, args[index+1])
				if err != nil {
					return 0, err
				}
				consumed = 2
			} else {
				return 0, fmt.Errorf("non-bool flag -%s must be last in cluster", shortStr)
			}
		case *StringSliceFlag:
			if i == len(shorts)-1 {
				// Last flag in cluster, can take value
				consumed, err := c.parseSliceFlag(args, index, f)
				if err != nil {
					return 0, err
				}
				if f.Variadic {
					c.lastVariadicFlag = flagName
				}
				return consumed, nil
			} else {
				return 0, fmt.Errorf("non-bool flag -%s must be last in cluster", shortStr)
			}
		}
	}

	return consumed, nil
}

func (c *Cmd) setStringValue(f *StringFlag, value string) error {
	if f.EnumConstraint != nil {
		valid := false
		for _, allowed := range *f.EnumConstraint {
			if value == allowed {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf(
				"Invalid '%s' value: %s (valid values: %s)",
				f.Name,
				value,
				strings.Join(*f.EnumConstraint, ", "),
			)
		}
	}

	if f.RegexConstraint != nil {
		if !f.RegexConstraint.MatchString(value) {
			return fmt.Errorf(
				"Invalid '%s' value: %s (must match regex: %s)",
				f.Name,
				value,
				f.RegexConstraint.String(),
			)
		}
	}

	*f.Value = value
	return nil
}

func (c *Cmd) setIntValue(f *IntFlag, value string) error {
	val, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid integer value for %s: %s", f.Name, value)
	}

	if f.min != nil {
		inclusive := f.minInclusive == nil || *f.minInclusive // default to inclusive
		if (inclusive && val < *f.min) || (!inclusive && val <= *f.min) {
			if inclusive {
				return fmt.Errorf("'%s' value %d is < minimum %d", f.Name, val, *f.min)
			} else {
				return fmt.Errorf("'%s' value %d is <= minimum (exclusive) %d", f.Name, val, *f.min)
			}
		}
	}

	if f.max != nil {
		inclusive := f.maxInclusive == nil || *f.maxInclusive // default to inclusive
		if (inclusive && val > *f.max) || (!inclusive && val >= *f.max) {
			if inclusive {
				return fmt.Errorf("'%s' value %d is > maximum %d", f.Name, val, *f.max)
			} else {
				return fmt.Errorf("'%s' value %d is >= maximum (exclusive) %d", f.Name, val, *f.max)
			}
		}
	}

	*f.Value = val
	return nil
}

func (c *Cmd) setInt64Value(f *Int64Flag, value string) error {
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid int64 value for %s: %s", f.Name, value)
	}

	if f.min != nil {
		inclusive := f.minInclusive == nil || *f.minInclusive // default to inclusive
		if (inclusive && val < *f.min) || (!inclusive && val <= *f.min) {
			if inclusive {
				return fmt.Errorf("'%s' value %d is < minimum %d", f.Name, val, *f.min)
			} else {
				return fmt.Errorf("'%s' value %d is <= minimum (exclusive) %d", f.Name, val, *f.min)
			}
		}
	}

	if f.max != nil {
		inclusive := f.maxInclusive == nil || *f.maxInclusive // default to inclusive
		if (inclusive && val > *f.max) || (!inclusive && val >= *f.max) {
			if inclusive {
				return fmt.Errorf("'%s' value %d is > maximum %d", f.Name, val, *f.max)
			} else {
				return fmt.Errorf("'%s' value %d is >= maximum (exclusive) %d", f.Name, val, *f.max)
			}
		}
	}

	*f.Value = val
	return nil
}

func (c *Cmd) setFloat64Value(f *Float64Flag, value string) error {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid float64 value for %s: %s", f.Name, value)
	}

	if f.min != nil {
		inclusive := f.minInclusive == nil || *f.minInclusive // default to inclusive
		if (inclusive && val < *f.min) || (!inclusive && val <= *f.min) {
			if inclusive {
				return fmt.Errorf("'%s' value %g is < minimum %g", f.Name, val, *f.min)
			} else {
				return fmt.Errorf("'%s' value %g is <= minimum (exclusive) %g", f.Name, val, *f.min)
			}
		}
	}

	if f.max != nil {
		inclusive := f.maxInclusive == nil || *f.maxInclusive // default to inclusive
		if (inclusive && val > *f.max) || (!inclusive && val >= *f.max) {
			if inclusive {
				return fmt.Errorf("'%s' value %g is > maximum %g", f.Name, val, *f.max)
			} else {
				return fmt.Errorf("'%s' value %g is >= maximum (exclusive) %g", f.Name, val, *f.max)
			}
		}
	}

	*f.Value = val
	return nil
}

func (c *Cmd) parseSliceFlag(args []string, index int, f *StringSliceFlag) (int, error) {
	if !f.Variadic {
		// Single value
		if index+1 >= len(args) {
			return 1, nil // Empty slice
		}
		return c.appendStringSliceValue(f, args[index+1])
	}

	// Variadic - consume until next flag
	consumed := 1
	for i := index + 1; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			break
		}
		if _, err := c.appendStringSliceValue(f, args[i]); err != nil {
			return 0, err
		}
		consumed++
	}

	return consumed, nil
}

func (c *Cmd) appendStringSliceValue(f *StringSliceFlag, value string) (int, error) {
	if f.Separator != nil {
		parts := strings.Split(value, *f.Separator)
		for _, part := range parts {
			*f.Value = append(*f.Value, part)
		}
	} else {
		*f.Value = append(*f.Value, value)
	}
	return 2, nil
}

func (c *Cmd) parseIntSliceFlag(args []string, index int, f *IntSliceFlag) (int, error) {
	if !f.Variadic {
		// Single value
		if index+1 >= len(args) {
			return 1, nil // Empty slice
		}
		return c.appendIntSliceValue(f, args[index+1])
	}

	// Variadic - consume until next flag
	consumed := 1
	for i := index + 1; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			break
		}
		if _, err := c.appendIntSliceValue(f, args[i]); err != nil {
			return 0, err
		}
		consumed++
	}

	return consumed, nil
}

func (c *Cmd) appendIntSliceValue(f *IntSliceFlag, value string) (int, error) {
	if f.Separator != nil {
		parts := strings.Split(value, *f.Separator)
		for _, part := range parts {
			val, err := strconv.Atoi(part)
			if err != nil {
				return 0, fmt.Errorf("invalid integer value for %s: %s", f.Name, part)
			}
			*f.Value = append(*f.Value, val)
		}
	} else {
		val, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("invalid integer value for %s: %s", f.Name, value)
		}
		*f.Value = append(*f.Value, val)
	}
	return 2, nil
}

func (c *Cmd) parseInt64SliceFlag(args []string, index int, f *Int64SliceFlag) (int, error) {
	if !f.Variadic {
		// Single value
		if index+1 >= len(args) {
			return 1, nil // Empty slice
		}
		return c.appendInt64SliceValue(f, args[index+1])
	}

	// Variadic - consume until next flag
	consumed := 1
	for i := index + 1; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			break
		}
		if _, err := c.appendInt64SliceValue(f, args[i]); err != nil {
			return 0, err
		}
		consumed++
	}

	return consumed, nil
}

func (c *Cmd) appendInt64SliceValue(f *Int64SliceFlag, value string) (int, error) {
	if f.Separator != nil {
		parts := strings.Split(value, *f.Separator)
		for _, part := range parts {
			val, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid int64 value for %s: %s", f.Name, part)
			}
			*f.Value = append(*f.Value, val)
		}
	} else {
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid int64 value for %s: %s", f.Name, value)
		}
		*f.Value = append(*f.Value, val)
	}
	return 2, nil
}

func (c *Cmd) parseFloat64SliceFlag(args []string, index int, f *Float64SliceFlag) (int, error) {
	if !f.Variadic {
		// Single value
		if index+1 >= len(args) {
			return 1, nil // Empty slice
		}
		return c.appendFloat64SliceValue(f, args[index+1])
	}

	// Variadic - consume until next flag
	consumed := 1
	for i := index + 1; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			break
		}
		if _, err := c.appendFloat64SliceValue(f, args[i]); err != nil {
			return 0, err
		}
		consumed++
	}

	return consumed, nil
}

func (c *Cmd) appendFloat64SliceValue(f *Float64SliceFlag, value string) (int, error) {
	if f.Separator != nil {
		parts := strings.Split(value, *f.Separator)
		for _, part := range parts {
			val, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid float64 value for %s: %s", f.Name, part)
			}
			*f.Value = append(*f.Value, val)
		}
	} else {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid float64 value for %s: %s", f.Name, value)
		}
		*f.Value = append(*f.Value, val)
	}
	return 2, nil
}

func (c *Cmd) parseBoolSliceFlag(args []string, index int, f *BoolSliceFlag) (int, error) {
	if !f.Variadic {
		// Single value
		if index+1 >= len(args) {
			return 1, nil // Empty slice
		}
		return c.appendBoolSliceValue(f, args[index+1])
	}

	// Variadic - consume until next flag
	consumed := 1
	for i := index + 1; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			break
		}
		if _, err := c.appendBoolSliceValue(f, args[i]); err != nil {
			return 0, err
		}
		consumed++
	}

	return consumed, nil
}

func (c *Cmd) appendBoolSliceValue(f *BoolSliceFlag, value string) (int, error) {
	if f.Separator != nil {
		parts := strings.Split(value, *f.Separator)
		for _, part := range parts {
			val, err := strconv.ParseBool(part)
			if err != nil {
				// Try parsing as 0/1
				if part == "0" {
					val = false
				} else if part == "1" {
					val = true
				} else {
					return 0, fmt.Errorf("invalid bool value for %s: %s", f.Name, part)
				}
			}
			*f.Value = append(*f.Value, val)
		}
	} else {
		val, err := strconv.ParseBool(value)
		if err != nil {
			// Try parsing as 0/1
			if value == "0" {
				val = false
			} else if value == "1" {
				val = true
			} else {
				return 0, fmt.Errorf("invalid bool value for %s: %s", f.Name, value)
			}
		}
		*f.Value = append(*f.Value, val)
	}
	return 2, nil
}

func (c *Cmd) assignPositional(value string) error {
	// Find next unassigned positional flag
	for _, name := range c.positional {
		flag := c.flags[name]

		// Check if it's positional-only or can be positional
		switch f := flag.(type) {
		case *StringFlag:
			if f.FlagOnly {
				continue
			}
			if c.configured[name] {
				continue // Already assigned
			}
			c.configured[name] = true
			return c.setStringValue(f, value)
		case *IntFlag:
			if f.FlagOnly {
				continue
			}
			if c.configured[name] {
				continue // Already assigned
			}
			c.configured[name] = true
			return c.setIntValue(f, value)
		case *Int64Flag:
			if f.FlagOnly {
				continue
			}
			if c.configured[name] {
				continue // Already assigned
			}
			c.configured[name] = true
			return c.setInt64Value(f, value)
		case *Float64Flag:
			if f.FlagOnly {
				continue
			}
			if c.configured[name] {
				continue // Already assigned
			}
			c.configured[name] = true
			return c.setFloat64Value(f, value)
		case *BoolFlag:
			if f.FlagOnly {
				continue
			}
			if c.configured[name] {
				continue // Already assigned
			}
			c.configured[name] = true
			val, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid bool value for %s: %s", name, value)
			}
			*f.Value = val
			return nil
		case *StringSliceFlag:
			if f.FlagOnly {
				continue
			}
			if f.Variadic {
				// Variadic positional - collect if this is the current one or no flag seen since last variadic
				if c.lastVariadicFlag == name {
					c.configured[name] = true
					_, err := c.appendStringSliceValue(f, value)
					return err
				}
				// If we saw a flag since last variadic, skip variadic flags that have already been used
				if c.sawFlag && c.configured[name] {
					continue
				}
				// Start new variadic only if we haven't seen a flag or this is a new variadic
				if !c.sawFlag || c.lastVariadicFlag == "" {
					c.configured[name] = true
					c.lastVariadicFlag = name
					_, err := c.appendStringSliceValue(f, value)
					return err
				}
				// Skip this variadic if we've seen a flag
				continue
			}
			if c.configured[name] {
				continue // Already assigned
			}
			c.configured[name] = true
			_, err := c.appendStringSliceValue(f, value)
			return err
		}
	}

	return fmt.Errorf("unexpected positional argument: %s", value)
}

// checkExclusion checks if a flag excludes or is excluded by another flag
func (c *Cmd) checkExclusion(flagName string) error {
	if !c.flagConfiguredForRelationalConstraints(flagName) {
		return nil
	}

	// Check if this flag excludes any other configured flags
	if flag, exists := c.flags[flagName]; exists {
		var excludes *[]string
		switch f := flag.(type) {
		case *StringFlag:
			excludes = f.Excludes
		case *IntFlag:
			excludes = f.Excludes
		case *BoolFlag:
			excludes = f.Excludes
		case *Int64Flag:
			excludes = f.Excludes
		case *Float64Flag:
			excludes = f.Excludes
		case *SliceFlag[string]:
			excludes = f.Excludes
		case *SliceFlag[int]:
			excludes = f.Excludes
		case *SliceFlag[int64]:
			excludes = f.Excludes
		case *SliceFlag[float64]:
			excludes = f.Excludes
		case *SliceFlag[bool]:
			excludes = f.Excludes
		}

		if excludes != nil {
			for _, excluded := range *excludes {
				if c.flagConfiguredForRelationalConstraints(excluded) {
					return fmt.Errorf(
						"Invalid args: '%s' excludes '%s', but '%s' was set",
						flagName,
						excluded,
						excluded,
					)
				}
			}
		}
	}

	// Check if any other configured flag excludes this flag
	for otherName, otherFlag := range c.flags {
		if otherName == flagName || !c.flagConfiguredForRelationalConstraints(otherName) {
			continue
		}

		var otherExcludes *[]string
		switch f := otherFlag.(type) {
		case *StringFlag:
			otherExcludes = f.Excludes
		case *IntFlag:
			otherExcludes = f.Excludes
		case *BoolFlag:
			otherExcludes = f.Excludes
		case *Int64Flag:
			otherExcludes = f.Excludes
		case *Float64Flag:
			otherExcludes = f.Excludes
		case *SliceFlag[string]:
			otherExcludes = f.Excludes
		case *SliceFlag[int]:
			otherExcludes = f.Excludes
		case *SliceFlag[int64]:
			otherExcludes = f.Excludes
		case *SliceFlag[float64]:
			otherExcludes = f.Excludes
		case *SliceFlag[bool]:
			otherExcludes = f.Excludes
		}

		if otherExcludes != nil {
			for _, excluded := range *otherExcludes {
				if excluded == flagName {
					return fmt.Errorf(
						"Invalid args: '%s' excludes '%s', but '%s' was set",
						otherName,
						flagName,
						flagName,
					)
				}
			}
		}
	}

	return nil
}

// flagHasValue returns true if the flag has a value (either configured by user or has a default)
func (c *Cmd) flagHasValue(name string) bool {
	// First check if it was explicitly configured
	if c.configured[name] {
		return true
	}

	// Then check if it has a default value
	if flag, exists := c.flags[name]; exists {
		switch f := flag.(type) {
		case *StringFlag:
			return f.Default != nil
		case *IntFlag:
			return f.Default != nil
		case *BoolFlag:
			return f.Default != nil
		case *Int64Flag:
			return f.Default != nil
		case *Float64Flag:
			return f.Default != nil
		case *SliceFlag[string]:
			return f.Default != nil
		case *SliceFlag[int]:
			return f.Default != nil
		case *SliceFlag[int64]:
			return f.Default != nil
		case *SliceFlag[float64]:
			return f.Default != nil
		case *SliceFlag[bool]:
			return f.Default != nil
		}
	}

	return false
}

// parseBoolValue parses a string value as a boolean, supporting standard formats and 0/1
func (c *Cmd) parseBoolValue(value string) (bool, error) {
	val, err := strconv.ParseBool(value)
	if err != nil {
		// Try parsing as 0/1
		if value == "0" {
			return false, nil
		} else if value == "1" {
			return true, nil
		}
		return false, err
	}
	return val, nil
}

// flagConfiguredForRelationalConstraints returns true if the flag should be considered configured
// for the purposes of relational constraints (requires/excludes).
// For boolean flags, this only returns true when the flag's value is true.
// For other flag types, this returns true if the flag has a value (configured or default).
func (c *Cmd) flagConfiguredForRelationalConstraints(name string) bool {
	if flag, exists := c.flags[name]; exists {
		switch f := flag.(type) {
		case *BoolFlag:
			// For boolean flags, only consider them configured for relational constraints when true
			if f.Value != nil && *f.Value {
				return true
			}
			return false
		default:
			// For all other flag types, use the normal flagHasValue logic
			return c.flagHasValue(name)
		}
	}
	return false
}

func (c *Cmd) validateRequired() error {
	// First pass: Check relational constraints (requires/excludes)
	// These are more specific and should take precedence over generic "required flag missing" errors
	for name, flag := range c.flags {
		// Check requires constraints for flags that are configured for relational constraints
		if c.flagConfiguredForRelationalConstraints(name) {
			var requires *[]string
			switch f := flag.(type) {
			case *StringFlag:
				requires = f.Requires
			case *IntFlag:
				requires = f.Requires
			case *BoolFlag:
				requires = f.Requires
			case *Int64Flag:
				requires = f.Requires
			case *Float64Flag:
				requires = f.Requires
			case *SliceFlag[string]:
				requires = f.Requires
			case *SliceFlag[int]:
				requires = f.Requires
			case *SliceFlag[int64]:
				requires = f.Requires
			case *SliceFlag[float64]:
				requires = f.Requires
			case *SliceFlag[bool]:
				requires = f.Requires
			}

			if requires != nil {
				for _, req := range *requires {
					if !c.flagConfiguredForRelationalConstraints(req) {
						return fmt.Errorf("Invalid args: '%s' requires '%s', but '%s' was not set", name, req, req)
					}
				}
			}
		}

		// Check exclusion constraints using the new helper
		if err := c.checkExclusion(name); err != nil {
			return err
		}
	}

	// Second pass: Check if required flags are missing
	// This runs after relational constraints so more specific errors take precedence
	// Check in registration order: positional flags first, then non-positional flags
	var missingRequired []string

	// Check positional flags first
	for _, name := range c.positional {
		if c.isFlagRequired(name) && !c.configured[name] {
			missingRequired = append(missingRequired, name)
		}
	}

	// Then check non-positional flags
	for _, name := range c.nonPositional {
		if c.isFlagRequired(name) && !c.configured[name] {
			missingRequired = append(missingRequired, name)
		}
	}

	if len(missingRequired) > 0 {
		return fmt.Errorf("Missing required arguments: [%s]", strings.Join(missingRequired, ", "))
	}
	return nil
}

func (c *Cmd) isFlagRequired(name string) bool {
	flag, exists := c.flags[name]
	if !exists {
		return false
	}

	switch f := flag.(type) {
	case *StringFlag:
		return !f.Optional && f.Default == nil
	case *IntFlag:
		return !f.Optional && f.Default == nil
	case *BoolFlag:
		// Boolean flags have implicit default of false, so never required
		return false
	case *Int64Flag:
		return !f.Optional && f.Default == nil
	case *Float64Flag:
		return !f.Optional && f.Default == nil
	case *SliceFlag[string]:
		// Variadic slice flags have implicit default of empty slice, so never required
		return !f.Variadic && !f.Optional && f.Default == nil
	case *SliceFlag[int]:
		// Variadic slice flags have implicit default of empty slice, so never required
		return !f.Variadic && !f.Optional && f.Default == nil
	case *SliceFlag[int64]:
		// Variadic slice flags have implicit default of empty slice, so never required
		return !f.Variadic && !f.Optional && f.Default == nil
	case *SliceFlag[float64]:
		// Variadic slice flags have implicit default of empty slice, so never required
		return !f.Variadic && !f.Optional && f.Default == nil
	case *SliceFlag[bool]:
		// Boolean slice flags have implicit default of empty slice, so never required
		return false
	}
	return false
}

func (c *Cmd) applyGlobalFlags(subCmd *Cmd) error {
	for _, globalFlagName := range c.globalFlags {
		if flag, exists := c.flags[globalFlagName]; exists {
			// Only add flag if it doesn't already exist in subcommand
			if _, exists := subCmd.flags[globalFlagName]; !exists {
				subCmd.flags[globalFlagName] = flag
				if base := getBaseFlag(flag); base != nil && base.Short != "" {
					subCmd.shortToName[base.Short] = base.Name
				}
				// Also add to subcommand's global flags list and non-positional list
				subCmd.globalFlags = append(subCmd.globalFlags, globalFlagName)
				subCmd.nonPositional = append(subCmd.nonPositional, globalFlagName)
			}
		}
	}
	return nil
}

// Whether a flag was explicitly configured by the user.
func (c *Cmd) Configured(name string) bool {
	// Check if flag is configured in this command
	if configured, exists := c.configured[name]; exists && configured {
		return true
	}

	// Check all invoked subcommands recursively
	for _, subCmd := range c.subCmds {
		if subCmd.used != nil && *subCmd.used {
			if subCmd.Configured(name) {
				return true
			}
		}
	}

	return false
}

func (c *Cmd) GetUnknownArgs() []string {
	return c.unknownArgs
}

// BASE TYPES

type BaseFlag struct {
	Name             string
	Short            string
	Usage            string
	Optional         bool
	Hidden           bool
	HiddenInLongHelp bool
	PositionalOnly   bool
	FlagOnly         bool
	Excludes         *[]string
	Requires         *[]string
}

type Flag[T any] struct {
	BaseFlag
	Default *T
	Value   *T
}

// NON-SLICE FLAGS

type BoolFlag struct {
	Flag[bool]
}

type StringFlag struct {
	Flag[string]
	EnumConstraint  *[]string      // if set, the value must be one of these
	RegexConstraint *regexp.Regexp // if set, the value must match this regex
}

type IntFlag struct {
	Flag[int]
	min          *int
	max          *int
	minInclusive *bool
	maxInclusive *bool
}

type Int64Flag struct {
	Flag[int64]
	min          *int64
	max          *int64
	minInclusive *bool
	maxInclusive *bool
}

type Float64Flag struct {
	Flag[float64]
	min          *float64
	max          *float64
	minInclusive *bool
	maxInclusive *bool
}

// SLICE FLAGS

type SliceFlag[T any] struct {
	BaseFlag
	Separator *string
	Variadic  bool
	Default   *[]T
	Value     *[]T
}

type StringSliceFlag = SliceFlag[string]
type IntSliceFlag = SliceFlag[int]
type Int64SliceFlag = SliceFlag[int64]
type Float64SliceFlag = SliceFlag[float64]
type BoolSliceFlag = SliceFlag[bool]

// NEW ARGS

func NewBool(name string) *BoolFlag {
	return &BoolFlag{Flag: Flag[bool]{BaseFlag: BaseFlag{Name: name, Optional: false}}}
}

func NewString(name string) *StringFlag {
	return &StringFlag{Flag: Flag[string]{BaseFlag: BaseFlag{Name: name, Optional: false}}}
}

func NewInt(name string) *IntFlag {
	return &IntFlag{Flag: Flag[int]{BaseFlag: BaseFlag{Name: name, Optional: false}}}
}

func NewInt64(name string) *Int64Flag {
	return &Int64Flag{Flag: Flag[int64]{BaseFlag: BaseFlag{Name: name, Optional: false}}}
}

func NewFloat64(name string) *Float64Flag {
	return &Float64Flag{Flag: Flag[float64]{BaseFlag: BaseFlag{Name: name, Optional: false}}}
}

func NewStringSlice(name string) *StringSliceFlag {
	return &SliceFlag[string]{BaseFlag: BaseFlag{Name: name, Optional: false}}
}

func NewIntSlice(name string) *IntSliceFlag {
	return &SliceFlag[int]{BaseFlag: BaseFlag{Name: name, Optional: false}}
}

func NewInt64Slice(name string) *Int64SliceFlag {
	return &SliceFlag[int64]{BaseFlag: BaseFlag{Name: name, Optional: false}}
}

func NewFloat64Slice(name string) *Float64SliceFlag {
	return &SliceFlag[float64]{BaseFlag: BaseFlag{Name: name, Optional: false}}
}

func NewBoolSlice(name string) *BoolSliceFlag {
	return &SliceFlag[bool]{BaseFlag: BaseFlag{Name: name, Optional: false}}
}

// BOOL FLAG SETTERS

func (f *BoolFlag) SetShort(s string) *BoolFlag {
	f.Short = s
	return f
}

func (f *BoolFlag) SetUsage(u string) *BoolFlag {
	f.Usage = u
	return f
}

func (f *BoolFlag) SetDefault(v bool) *BoolFlag {
	f.Default = &v
	return f
}

func (f *BoolFlag) SetOptional(b bool) *BoolFlag {
	f.Optional = b
	return f
}

func (f *BoolFlag) SetHidden(b bool) *BoolFlag {
	f.Hidden = b
	return f
}

func (f *BoolFlag) SetHiddenInLongHelp(b bool) *BoolFlag {
	f.HiddenInLongHelp = b
	return f
}

func (f *BoolFlag) SetPositionalOnly(b bool) *BoolFlag {
	f.PositionalOnly = b
	return f
}

func (f *BoolFlag) SetFlagOnly(b bool) *BoolFlag {
	f.FlagOnly = b
	return f
}

func (f *BoolFlag) SetExcludes(flags []string) *BoolFlag {
	f.Excludes = &flags
	return f
}

func (f *BoolFlag) SetRequires(flags []string) *BoolFlag {
	f.Requires = &flags
	return f
}

// STRING FLAG SETTERS

func (f *StringFlag) SetShort(s string) *StringFlag {
	f.Short = s
	return f
}

func (f *StringFlag) SetUsage(u string) *StringFlag {
	f.Usage = u
	return f
}

func (f *StringFlag) SetDefault(v string) *StringFlag {
	f.Default = &v
	return f
}

func (f *StringFlag) SetOptional(b bool) *StringFlag {
	f.Optional = b
	return f
}

func (f *StringFlag) SetHidden(b bool) *StringFlag {
	f.Hidden = b
	return f
}

func (f *StringFlag) SetHiddenInLongHelp(b bool) *StringFlag {
	f.HiddenInLongHelp = b
	return f
}

func (f *StringFlag) SetPositionalOnly(b bool) *StringFlag {
	f.PositionalOnly = b
	return f
}

func (f *StringFlag) SetFlagOnly(b bool) *StringFlag {
	f.FlagOnly = b
	return f
}

func (f *StringFlag) SetExcludes(flags []string) *StringFlag {
	f.Excludes = &flags
	return f
}

func (f *StringFlag) SetRequires(flags []string) *StringFlag {
	f.Requires = &flags
	return f
}

func (f *StringFlag) SetEnumConstraint(values []string) *StringFlag {
	if len(values) == 0 {
		f.EnumConstraint = nil
	} else {
		f.EnumConstraint = &values
	}
	return f
}

func (f *StringFlag) SetRegexConstraint(regex *regexp.Regexp) *StringFlag {
	f.RegexConstraint = regex
	return f
}

// INT FLAG SETTERS

func (f *IntFlag) SetShort(s string) *IntFlag {
	f.Short = s
	return f
}

func (f *IntFlag) SetUsage(u string) *IntFlag {
	f.Usage = u
	return f
}

func (f *IntFlag) SetDefault(v int) *IntFlag {
	f.Default = &v
	return f
}

func (f *IntFlag) SetOptional(b bool) *IntFlag {
	f.Optional = b
	return f
}

func (f *IntFlag) SetHidden(b bool) *IntFlag {
	f.Hidden = b
	return f
}

func (f *IntFlag) SetHiddenInLongHelp(b bool) *IntFlag {
	f.HiddenInLongHelp = b
	return f
}

func (f *IntFlag) SetPositionalOnly(b bool) *IntFlag {
	f.PositionalOnly = b
	return f
}

func (f *IntFlag) SetFlagOnly(b bool) *IntFlag {
	f.FlagOnly = b
	return f
}

func (f *IntFlag) SetExcludes(flags []string) *IntFlag {
	f.Excludes = &flags
	return f
}

func (f *IntFlag) SetRequires(flags []string) *IntFlag {
	f.Requires = &flags
	return f
}

func (f *IntFlag) SetMin(min int, inclusive bool) *IntFlag {
	f.min = &min
	f.minInclusive = &inclusive
	return f
}

func (f *IntFlag) SetMax(max int, inclusive bool) *IntFlag {
	f.max = &max
	f.maxInclusive = &inclusive
	return f
}

// INT64 FLAG SETTERS

func (f *Int64Flag) SetShort(s string) *Int64Flag {
	f.Short = s
	return f
}

func (f *Int64Flag) SetUsage(u string) *Int64Flag {
	f.Usage = u
	return f
}

func (f *Int64Flag) SetDefault(v int64) *Int64Flag {
	f.Default = &v
	return f
}

func (f *Int64Flag) SetOptional(b bool) *Int64Flag {
	f.Optional = b
	return f
}

func (f *Int64Flag) SetHidden(b bool) *Int64Flag {
	f.Hidden = b
	return f
}

func (f *Int64Flag) SetHiddenInLongHelp(b bool) *Int64Flag {
	f.HiddenInLongHelp = b
	return f
}

func (f *Int64Flag) SetPositionalOnly(b bool) *Int64Flag {
	f.PositionalOnly = b
	return f
}

func (f *Int64Flag) SetFlagOnly(b bool) *Int64Flag {
	f.FlagOnly = b
	return f
}

func (f *Int64Flag) SetExcludes(flags []string) *Int64Flag {
	f.Excludes = &flags
	return f
}

func (f *Int64Flag) SetRequires(flags []string) *Int64Flag {
	f.Requires = &flags
	return f
}

func (f *Int64Flag) SetMin(min int64, inclusive bool) *Int64Flag {
	f.min = &min
	f.minInclusive = &inclusive
	return f
}

func (f *Int64Flag) SetMax(max int64, inclusive bool) *Int64Flag {
	f.max = &max
	f.maxInclusive = &inclusive
	return f
}

// FLOAT64 FLAG SETTERS

func (f *Float64Flag) SetShort(s string) *Float64Flag {
	f.Short = s
	return f
}

func (f *Float64Flag) SetUsage(u string) *Float64Flag {
	f.Usage = u
	return f
}

func (f *Float64Flag) SetDefault(v float64) *Float64Flag {
	f.Default = &v
	return f
}

func (f *Float64Flag) SetOptional(b bool) *Float64Flag {
	f.Optional = b
	return f
}

func (f *Float64Flag) SetHidden(b bool) *Float64Flag {
	f.Hidden = b
	return f
}

func (f *Float64Flag) SetHiddenInLongHelp(b bool) *Float64Flag {
	f.HiddenInLongHelp = b
	return f
}

func (f *Float64Flag) SetPositionalOnly(b bool) *Float64Flag {
	f.PositionalOnly = b
	return f
}

func (f *Float64Flag) SetFlagOnly(b bool) *Float64Flag {
	f.FlagOnly = b
	return f
}

func (f *Float64Flag) SetExcludes(flags []string) *Float64Flag {
	f.Excludes = &flags
	return f
}

func (f *Float64Flag) SetRequires(flags []string) *Float64Flag {
	f.Requires = &flags
	return f
}

func (f *Float64Flag) SetMin(min float64, inclusive bool) *Float64Flag {
	f.min = &min
	f.minInclusive = &inclusive
	return f
}

func (f *Float64Flag) SetMax(max float64, inclusive bool) *Float64Flag {
	f.max = &max
	f.maxInclusive = &inclusive
	return f
}

// SLICE FLAG SETTERS

func (f *SliceFlag[T]) SetShort(s string) *SliceFlag[T] {
	f.Short = s
	return f
}

func (f *SliceFlag[T]) SetUsage(u string) *SliceFlag[T] {
	f.Usage = u
	return f
}

func (f *SliceFlag[T]) SetDefault(v []T) *SliceFlag[T] {
	f.Default = &v
	return f
}

func (f *SliceFlag[T]) SetOptional(b bool) *SliceFlag[T] {
	f.Optional = b
	return f
}

func (f *SliceFlag[T]) SetHidden(b bool) *SliceFlag[T] {
	f.Hidden = b
	return f
}

func (f *SliceFlag[T]) SetHiddenInLongHelp(b bool) *SliceFlag[T] {
	f.HiddenInLongHelp = b
	return f
}

func (f *SliceFlag[T]) SetPositionalOnly(b bool) *SliceFlag[T] {
	f.PositionalOnly = b
	return f
}

func (f *SliceFlag[T]) SetFlagOnly(b bool) *SliceFlag[T] {
	f.FlagOnly = b
	return f
}

func (f *SliceFlag[T]) SetExcludes(flags []string) *SliceFlag[T] {
	f.Excludes = &flags
	return f
}

func (f *SliceFlag[T]) SetRequires(flags []string) *SliceFlag[T] {
	f.Requires = &flags
	return f
}

func (f *SliceFlag[T]) SetSeparator(sep string) *SliceFlag[T] {
	f.Separator = &sep
	return f
}

func (f *SliceFlag[T]) SetVariadic(b bool) *SliceFlag[T] {
	f.Variadic = b
	return f
}

// REGISTER METHODS

type RegisterOption func(*registerConfig)

type registerConfig struct {
	global bool
}

func WithGlobal(g bool) RegisterOption {
	return func(c *registerConfig) {
		c.global = g
	}
}

func (f *BoolFlag) Register(cmd *Cmd, opts ...RegisterOption) (*bool, error) {
	ptr := new(bool)
	return ptr, f.RegisterWithPtr(cmd, ptr, opts...)
}

func (f *BoolFlag) RegisterWithPtr(cmd *Cmd, ptr *bool, opts ...RegisterOption) error {
	regConf := &registerConfig{}
	for _, opt := range opts {
		opt(regConf)
	}

	if _, exists := cmd.flags[f.Name]; exists {
		return fmt.Errorf("flag %q already defined", f.Name)
	}

	if regConf.global {
		cmd.globalFlags = append(cmd.globalFlags, f.Name)
	}

	// Create copy and set value pointer
	flag := *f
	flag.Value = ptr

	// Global flags should be flag-only (not positional)
	// Bool flags are always flag-only (non-positional)
	if regConf.global {
		flag.FlagOnly = true
	}

	// Add to short mapping
	if f.Short != "" {
		if _, exists := cmd.shortToName[f.Short]; exists {
			return fmt.Errorf("short flag %q already defined", f.Short)
		}
		cmd.shortToName[f.Short] = f.Name
	}

	cmd.flags[f.Name] = &flag

	// Bool flags are always non-positional (flag-only)
	cmd.nonPositional = append(cmd.nonPositional, f.Name)

	return nil
}

func (f *StringFlag) Register(cmd *Cmd, opts ...RegisterOption) (*string, error) {
	ptr := new(string)
	return ptr, f.RegisterWithPtr(cmd, ptr, opts...)
}

func (f *StringFlag) RegisterWithPtr(cmd *Cmd, ptr *string, opts ...RegisterOption) error {
	regConf := &registerConfig{}
	for _, opt := range opts {
		opt(regConf)
	}

	if _, exists := cmd.flags[f.Name]; exists {
		return fmt.Errorf("flag %q already defined", f.Name)
	}

	if regConf.global {
		cmd.globalFlags = append(cmd.globalFlags, f.Name)
	}

	// Create copy and set value pointer
	flag := *f
	flag.Value = ptr

	// Global flags should be flag-only (not positional)
	if regConf.global {
		flag.FlagOnly = true
		flag.Optional = true // TODO test, make clearer? Might be surprising/undesirable?
	}

	// Add to short mapping
	if f.Short != "" {
		if _, exists := cmd.shortToName[f.Short]; exists {
			return fmt.Errorf("short flag %q already defined", f.Short)
		}
		cmd.shortToName[f.Short] = f.Name
	}

	cmd.flags[f.Name] = &flag
	if !flag.FlagOnly {
		// Check for positional-only after variadic error
		if flag.PositionalOnly {
			if err := cmd.validatePositionalOnlyAfterVariadic(f.Name); err != nil {
				return err
			}
		}
		cmd.positional = append(cmd.positional, f.Name)
	} else {
		cmd.nonPositional = append(cmd.nonPositional, f.Name)
	}

	return nil
}

func (f *IntFlag) Register(cmd *Cmd, opts ...RegisterOption) (*int, error) {
	ptr := new(int)
	return ptr, f.RegisterWithPtr(cmd, ptr, opts...)
}

func (f *IntFlag) RegisterWithPtr(cmd *Cmd, ptr *int, opts ...RegisterOption) error {
	regConf := &registerConfig{}
	for _, opt := range opts {
		opt(regConf)
	}

	if _, exists := cmd.flags[f.Name]; exists {
		return fmt.Errorf("flag %q already defined", f.Name)
	}

	if regConf.global {
		cmd.globalFlags = append(cmd.globalFlags, f.Name)
	}

	// Create copy and set value pointer
	flag := *f
	flag.Value = ptr

	// Global flags should be flag-only (not positional)
	if regConf.global {
		flag.FlagOnly = true
	}

	// Add to short mapping
	if f.Short != "" {
		if _, exists := cmd.shortToName[f.Short]; exists {
			return fmt.Errorf("short flag %q already defined", f.Short)
		}
		cmd.shortToName[f.Short] = f.Name
	}

	cmd.flags[f.Name] = &flag
	if !flag.FlagOnly {
		// Check for positional-only after variadic error
		if flag.PositionalOnly {
			if err := cmd.validatePositionalOnlyAfterVariadic(f.Name); err != nil {
				return err
			}
		}
		cmd.positional = append(cmd.positional, f.Name)
	} else {
		cmd.nonPositional = append(cmd.nonPositional, f.Name)
	}

	return nil
}

func (f *Int64Flag) Register(cmd *Cmd, opts ...RegisterOption) (*int64, error) {
	ptr := new(int64)
	return ptr, f.RegisterWithPtr(cmd, ptr, opts...)
}

func (f *Int64Flag) RegisterWithPtr(cmd *Cmd, ptr *int64, opts ...RegisterOption) error {
	regConf := &registerConfig{}
	for _, opt := range opts {
		opt(regConf)
	}

	if _, exists := cmd.flags[f.Name]; exists {
		return fmt.Errorf("flag %q already defined", f.Name)
	}

	if regConf.global {
		cmd.globalFlags = append(cmd.globalFlags, f.Name)
	}

	// Create copy and set value pointer
	flag := *f
	flag.Value = ptr

	// Global flags should be flag-only (not positional)
	if regConf.global {
		flag.FlagOnly = true
	}

	// Add to short mapping
	if f.Short != "" {
		if _, exists := cmd.shortToName[f.Short]; exists {
			return fmt.Errorf("short flag %q already defined", f.Short)
		}
		cmd.shortToName[f.Short] = f.Name
	}

	cmd.flags[f.Name] = &flag
	if !flag.FlagOnly {
		// Check for positional-only after variadic error
		if flag.PositionalOnly {
			if err := cmd.validatePositionalOnlyAfterVariadic(f.Name); err != nil {
				return err
			}
		}
		cmd.positional = append(cmd.positional, f.Name)
	} else {
		cmd.nonPositional = append(cmd.nonPositional, f.Name)
	}

	return nil
}

func (f *Float64Flag) Register(cmd *Cmd, opts ...RegisterOption) (*float64, error) {
	ptr := new(float64)
	return ptr, f.RegisterWithPtr(cmd, ptr, opts...)
}

func (f *Float64Flag) RegisterWithPtr(cmd *Cmd, ptr *float64, opts ...RegisterOption) error {
	regConf := &registerConfig{}
	for _, opt := range opts {
		opt(regConf)
	}

	if _, exists := cmd.flags[f.Name]; exists {
		return fmt.Errorf("flag %q already defined", f.Name)
	}

	if regConf.global {
		cmd.globalFlags = append(cmd.globalFlags, f.Name)
	}

	// Create copy and set value pointer
	flag := *f
	flag.Value = ptr

	// Global flags should be flag-only (not positional)
	if regConf.global {
		flag.FlagOnly = true
	}

	// Add to short mapping
	if f.Short != "" {
		if _, exists := cmd.shortToName[f.Short]; exists {
			return fmt.Errorf("short flag %q already defined", f.Short)
		}
		cmd.shortToName[f.Short] = f.Name
	}

	cmd.flags[f.Name] = &flag
	if !flag.FlagOnly {
		// Check for positional-only after variadic error
		if flag.PositionalOnly {
			if err := cmd.validatePositionalOnlyAfterVariadic(f.Name); err != nil {
				return err
			}
		}
		cmd.positional = append(cmd.positional, f.Name)
	} else {
		cmd.nonPositional = append(cmd.nonPositional, f.Name)
	}

	return nil
}

func (f *SliceFlag[T]) Register(cmd *Cmd, opts ...RegisterOption) (*[]T, error) {
	ptr := new([]T)
	return ptr, f.RegisterWithPtr(cmd, ptr, opts...)
}

func (f *SliceFlag[T]) RegisterWithPtr(cmd *Cmd, ptr *[]T, opts ...RegisterOption) error {
	regConf := &registerConfig{}
	for _, opt := range opts {
		opt(regConf)
	}

	if _, exists := cmd.flags[f.Name]; exists {
		return fmt.Errorf("flag %q already defined", f.Name)
	}

	if regConf.global {
		cmd.globalFlags = append(cmd.globalFlags, f.Name)
	}

	// Create copy and set value pointer
	flag := *f
	flag.Value = ptr

	// Global flags should be flag-only (not positional)
	if regConf.global {
		flag.FlagOnly = true
	}

	// Add to short mapping
	if f.Short != "" {
		if _, exists := cmd.shortToName[f.Short]; exists {
			return fmt.Errorf("short flag %q already defined", f.Short)
		}
		cmd.shortToName[f.Short] = f.Name
	}

	cmd.flags[f.Name] = &flag
	if !flag.FlagOnly {
		// Check for positional-only after variadic error
		if flag.PositionalOnly {
			if err := cmd.validatePositionalOnlyAfterVariadic(f.Name); err != nil {
				return err
			}
		}
		cmd.positional = append(cmd.positional, f.Name)
	} else {
		cmd.nonPositional = append(cmd.nonPositional, f.Name)
	}

	return nil
}

func (c *Cmd) RegisterCmd(subCmd *Cmd) (*bool, error) {
	if _, exists := c.subCmds[subCmd.name]; exists {
		return nil, fmt.Errorf("command %q already defined", subCmd.name)
	}

	c.subCmds[subCmd.name] = subCmd
	subCmd.used = new(bool)

	// Apply global flags to subcommand for usage generation
	if err := c.applyGlobalFlags(subCmd); err != nil {
		return nil, err
	}

	return subCmd.used, nil
}

// Helper functions
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (c *Cmd) validatePositionalOnlyAfterVariadic(flagName string) error {
	// Check if there's already a variadic positional flag
	for _, existingName := range c.positional {
		existingFlag := c.flags[existingName]

		// Check if this existing flag is variadic
		switch f := existingFlag.(type) {
		case *StringSliceFlag:
			if f.Variadic {
				return fmt.Errorf("cannot register positional-only flag %q after variadic positional flag %q (positional-only flags cannot be set after variadic flags)", flagName, existingName)
			}
		case *IntSliceFlag:
			if f.Variadic {
				return fmt.Errorf("cannot register positional-only flag %q after variadic positional flag %q (positional-only flags cannot be set after variadic flags)", flagName, existingName)
			}
		case *Int64SliceFlag:
			if f.Variadic {
				return fmt.Errorf("cannot register positional-only flag %q after variadic positional flag %q (positional-only flags cannot be set after variadic flags)", flagName, existingName)
			}
		case *Float64SliceFlag:
			if f.Variadic {
				return fmt.Errorf("cannot register positional-only flag %q after variadic positional flag %q (positional-only flags cannot be set after variadic flags)", flagName, existingName)
			}
		case *BoolSliceFlag:
			if f.Variadic {
				return fmt.Errorf("cannot register positional-only flag %q after variadic positional flag %q (positional-only flags cannot be set after variadic flags)", flagName, existingName)
			}
		}
	}
	return nil
}

func (c *Cmd) GenerateUsage(isLongHelp bool) string {
	return c.generateUsage(isLongHelp)
}

func (c *Cmd) GenerateShortUsage() string {
	return c.generateUsage(false)
}

func (c *Cmd) GenerateLongUsage() string {
	return c.generateUsage(true)
}

func (c *Cmd) generateUsage(isLongHelp bool) string {
	var sb strings.Builder

	if c.description != "" {
		sb.WriteString(c.description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("Usage:\n  ")
	sb.WriteString(c.generateSynopsis(isLongHelp))
	sb.WriteString("\n")

	// Separate script and global flags
	var scriptFlags []any
	var globalFlags []any

	// Use a map to keep track of added flags to avoid duplicates
	addedFlags := make(map[string]bool)

	// Process positional flags in registration order
	for _, name := range c.positional {
		if addedFlags[name] {
			continue
		}
		flag := c.flags[name]

		isGlobal := false
		for _, gName := range c.globalFlags {
			if name == gName {
				isGlobal = true
				break
			}
		}
		if isGlobal {
			globalFlags = append(globalFlags, flag)
		} else {
			scriptFlags = append(scriptFlags, flag)
		}
		addedFlags[name] = true
	}

	// Process non-positional flags in registration order
	for _, name := range c.nonPositional {
		if addedFlags[name] {
			continue
		}
		flag := c.flags[name]

		isGlobal := false
		for _, gName := range c.globalFlags {
			if name == gName {
				isGlobal = true
				break
			}
		}
		if isGlobal {
			globalFlags = append(globalFlags, flag)
		} else {
			scriptFlags = append(scriptFlags, flag)
		}
		addedFlags[name] = true
	}

	if len(c.subCmds) > 0 {
		sb.WriteString("\nCommands:\n")
		// Sort subcommand names for consistent output
		var subCmdNames []string
		for name := range c.subCmds {
			subCmdNames = append(subCmdNames, name)
		}
		sort.Strings(subCmdNames)
		for _, name := range subCmdNames {
			subCmd := c.subCmds[name]
			if subCmd.description != "" {
				sb.WriteString(fmt.Sprintf("  %-30s%s\n", name, subCmd.description))
			} else {
				sb.WriteString(fmt.Sprintf("  %s\n", name))
			}
		}
	}

	if len(scriptFlags) > 0 {
		sb.WriteString("\nArguments:\n")
		sb.WriteString(c.formatFlags(scriptFlags, isLongHelp))
	}

	if len(globalFlags) > 0 {
		sb.WriteString("\nGlobal options:\n")
		sb.WriteString(c.formatFlags(globalFlags, isLongHelp))
	}

	return sb.String()
}

func (c *Cmd) generateSynopsis(isLongHelp bool) string {
	var sb strings.Builder
	sb.WriteString(c.name)

	if len(c.subCmds) > 0 {
		sb.WriteString(" [subcommand]")
		// Still show parent command flags in synopsis even when subcommands exist
		for _, name := range c.positional {
			flag := c.flags[name]
			base := getBaseFlag(flag)
			if base.Hidden {
				continue
			}
			if isLongHelp && base.HiddenInLongHelp {
				continue
			}

			// Show positional-only flags or non-bool flags in synopsis (bools never appear)
			flagType := getFlagType(flag)
			if base.PositionalOnly || flagType != "bool" {
				var argName string

				// Check if it's a variadic slice
				isVariadic := false
				switch f := flag.(type) {
				case *StringSliceFlag:
					isVariadic = f.Variadic
				case *IntSliceFlag:
					isVariadic = f.Variadic
				case *Int64SliceFlag:
					isVariadic = f.Variadic
				case *Float64SliceFlag:
					isVariadic = f.Variadic
				case *BoolSliceFlag:
					isVariadic = f.Variadic
				}

				if isVariadic {
					argName = name + "..."
				} else {
					argName = name
				}

				// Determine if flag should show as required or optional in synopsis
				shouldBeOptional := c.shouldFlagBeOptionalInSynopsis(flag)

				if shouldBeOptional {
					sb.WriteString(fmt.Sprintf(" [%s]", argName))
				} else {
					sb.WriteString(fmt.Sprintf(" <%s>", argName))
				}
			}
		}
		sb.WriteString(" [OPTIONS]")
		return sb.String()
	}

	// First pass: collect positional-only flags
	var positionalOnlyFlags []string
	var nonPositionalFlags []string

	for _, name := range c.positional {
		flag := c.flags[name]
		base := getBaseFlag(flag)
		if base.Hidden {
			continue
		}
		if isLongHelp && base.HiddenInLongHelp {
			continue
		}

		flagType := getFlagType(flag)
		if flagType == "bool" {
			continue // Bools never appear in synopsis
		}

		if base.PositionalOnly {
			positionalOnlyFlags = append(positionalOnlyFlags, name)
		} else {
			// Check if it's variadic (always appears in synopsis)
			isVariadic := false
			switch f := flag.(type) {
			case *StringSliceFlag:
				isVariadic = f.Variadic
			case *IntSliceFlag:
				isVariadic = f.Variadic
			case *Int64SliceFlag:
				isVariadic = f.Variadic
			case *Float64SliceFlag:
				isVariadic = f.Variadic
			case *BoolSliceFlag:
				isVariadic = f.Variadic
			}

			if isVariadic || !base.Optional {
				// Variadic flags always appear (they're inherently optional)
				// Non-variadic required flags also appear
				nonPositionalFlags = append(nonPositionalFlags, name)
			}
		}
		// Non-variadic optional flags don't appear in synopsis
	}

	// Add non-positional flags from nonPositional list (they might not be in positional)
	for _, name := range c.nonPositional {
		flag := c.flags[name]
		base := getBaseFlag(flag)
		if base.Hidden {
			continue
		}
		if isLongHelp && base.HiddenInLongHelp {
			continue
		}

		// Skip global flags - they don't appear in synopsis
		isGlobal := false
		for _, gName := range c.globalFlags {
			if name == gName {
				isGlobal = true
				break
			}
		}
		if isGlobal {
			continue
		}

		flagType := getFlagType(flag)
		if flagType == "bool" || base.Optional {
			continue // Bools and optional flags never appear in synopsis
		}

		// Check if already added
		found := false
		for _, existing := range nonPositionalFlags {
			if existing == name {
				found = true
				break
			}
		}
		if !found {
			nonPositionalFlags = append(nonPositionalFlags, name)
		}
	}

	// Process positional-only flags first, but stop after first variadic
	for _, name := range positionalOnlyFlags {
		flag := c.flags[name]
		shouldBeOptional := c.shouldFlagBeOptionalInSynopsis(flag)

		// Check if it's a variadic positional flag
		argName := name
		isVariadic := false
		switch f := flag.(type) {
		case *StringSliceFlag:
			if f.Variadic {
				argName = name + "..."
				isVariadic = true
			}
		case *IntSliceFlag:
			if f.Variadic {
				argName = name + "..."
				isVariadic = true
			}
		case *Int64SliceFlag:
			if f.Variadic {
				argName = name + "..."
				isVariadic = true
			}
		case *Float64SliceFlag:
			if f.Variadic {
				argName = name + "..."
				isVariadic = true
			}
		case *BoolSliceFlag:
			if f.Variadic {
				argName = name + "..."
				isVariadic = true
			}
		}

		if shouldBeOptional {
			sb.WriteString(fmt.Sprintf(" [%s]", argName))
		} else {
			sb.WriteString(fmt.Sprintf(" <%s>", argName))
		}

		// Stop after first variadic positional flag
		if isVariadic {
			sb.WriteString(" [OPTIONS]")
			return sb.String()
		}
	}

	// Then process non-positional flags, but stop after first variadic
	for _, name := range nonPositionalFlags {
		flag := c.flags[name]

		// Check if it's variadic
		isVariadic := false
		switch f := flag.(type) {
		case *StringSliceFlag:
			isVariadic = f.Variadic
		case *IntSliceFlag:
			isVariadic = f.Variadic
		case *Int64SliceFlag:
			isVariadic = f.Variadic
		case *Float64SliceFlag:
			isVariadic = f.Variadic
		case *BoolSliceFlag:
			isVariadic = f.Variadic
		}

		if isVariadic {
			// All variadic flags show as [name...]
			sb.WriteString(fmt.Sprintf(" [%s...]", name))
			// Stop after first variadic flag
			sb.WriteString(" [OPTIONS]")
			return sb.String()
		} else {
			// Non-variadic required flags show as <name>
			shouldBeOptional := c.shouldFlagBeOptionalInSynopsis(flag)
			if shouldBeOptional {
				sb.WriteString(fmt.Sprintf(" [%s]", name))
			} else {
				sb.WriteString(fmt.Sprintf(" <%s>", name))
			}
		}
	}

	sb.WriteString(" [OPTIONS]")
	return sb.String()
}

func (c *Cmd) formatFlags(flags []any, isLongHelp bool) string {
	// Use flags in the order they were passed (already correctly ordered by generateUsage)
	allFlags := flags

	// First pass: calculate maximum width for alignment
	maxWidth := 0
	var flagParts []string

	for _, flag := range allFlags {
		base := getBaseFlag(flag)
		if base.Hidden {
			flagParts = append(flagParts, "")
			continue
		}
		if isLongHelp && base.HiddenInLongHelp {
			flagParts = append(flagParts, "")
			continue
		}

		var flagPart string
		if base.PositionalOnly {
			// Positional-only flags show without dashes
			flagPart = fmt.Sprintf("  %s", base.Name)
		} else if base.Short != "" {
			flagPart = fmt.Sprintf("  -%s, --%s", base.Short, base.Name)
		} else {
			flagPart = fmt.Sprintf("      --%s", base.Name)
		}

		typeStr := getFlagType(flag)
		if typeStr != "bool" {
			flagPart = fmt.Sprintf("%s %s", flagPart, typeStr)
		}

		flagParts = append(flagParts, flagPart)
		if len(flagPart) > maxWidth {
			maxWidth = len(flagPart)
		}
	}

	// Use dynamic alignment: longest left side + 3 spaces
	maxWidth = maxWidth + 3

	// Second pass: generate aligned output
	var sb strings.Builder
	for i, flag := range allFlags {
		flagPart := flagParts[i]
		if flagPart == "" {
			continue // hidden flag
		}

		base := getBaseFlag(flag)
		sb.WriteString(flagPart)

		if base.Usage != "" {
			// Calculate padding to align descriptions
			padding := maxWidth - len(flagPart)
			if padding < 1 {
				padding = 1
			}
			sb.WriteString(strings.Repeat(" ", padding))

			// Add optional marker for flags that should be optional
			// but not for variadic flags (their type already indicates optionality)
			isVariadic := false
			switch f := flag.(type) {
			case *StringSliceFlag:
				isVariadic = f.Variadic
			case *IntSliceFlag:
				isVariadic = f.Variadic
			case *Int64SliceFlag:
				isVariadic = f.Variadic
			case *Float64SliceFlag:
				isVariadic = f.Variadic
			case *BoolSliceFlag:
				isVariadic = f.Variadic
			}

			// Show status markers for non-variadic flags:
			// For positional-only flags: show (optional) if explicitly optional, otherwise no marker
			// For other flags: (optional) if explicitly optional AND no default, (required) if required AND no default
			hasDefault := c.flagHasDefault(flag)
			var shouldShowOptional, shouldShowRequired bool
			if base.PositionalOnly {
				shouldShowOptional = base.Optional
				shouldShowRequired = false
			} else {
				shouldShowOptional = base.Optional && !hasDefault
				shouldShowRequired = !base.Optional && !hasDefault
			}

			if shouldShowOptional && !isVariadic {
				sb.WriteString("(optional) ")
			} else if shouldShowRequired && !isVariadic {
				sb.WriteString("(required) ")
			}

			sb.WriteString(base.Usage)

			// Add constraints
			constraints := c.getConstraintString(flag)
			if constraints != "" {
				sb.WriteString(" ")
				sb.WriteString(constraints)
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func getBaseFlag(flag any) *BaseFlag {
	switch f := flag.(type) {
	case *BoolFlag:
		return &f.BaseFlag
	case *StringFlag:
		return &f.BaseFlag
	case *IntFlag:
		return &f.BaseFlag
	case *Int64Flag:
		return &f.BaseFlag
	case *Float64Flag:
		return &f.BaseFlag
	case *StringSliceFlag:
		return &f.BaseFlag
	case *IntSliceFlag:
		return &f.BaseFlag
	case *Int64SliceFlag:
		return &f.BaseFlag
	case *Float64SliceFlag:
		return &f.BaseFlag
	case *BoolSliceFlag:
		return &f.BaseFlag
	}
	return nil
}

func getFlagType(flag any) string {
	switch f := flag.(type) {
	case *BoolFlag:
		return "bool"
	case *StringFlag:
		return "str"
	case *IntFlag:
		return "int"
	case *Int64Flag:
		return "int64"
	case *Float64Flag:
		return "float"
	case *BoolSliceFlag:
		return "bool"
	case *StringSliceFlag:
		if f.Variadic {
			base := getBaseFlag(flag)
			if base.Optional {
				return "[strs...]"
			}
			return "strs..."
		}
		return "strs"
	case *IntSliceFlag:
		if f.Variadic {
			base := getBaseFlag(flag)
			if base.Optional {
				return "[ints...]"
			}
			return "ints..."
		}
		return "ints"
	case *Int64SliceFlag:
		if f.Variadic {
			base := getBaseFlag(flag)
			if base.Optional {
				return "[int64s...]"
			}
			return "int64s..."
		}
		return "int64s"
	case *Float64SliceFlag:
		if f.Variadic {
			base := getBaseFlag(flag)
			if base.Optional {
				return "[floats...]"
			}
			return "floats..."
		}
		return "floats"
	}
	return ""
}

func (c *Cmd) getConstraintString(flag any) string {
	var parts []string

	// Add range constraints
	if rangeStr := c.getRangeString(flag); rangeStr != "" {
		parts = append(parts, "Range: "+rangeStr)
	}

	// Add enum or regex constraints (regex takes priority if both exist)
	if regexStr := c.getRegexString(flag); regexStr != "" {
		parts = append(parts, "Must match pattern: "+regexStr)
	} else if enumStr := c.getEnumString(flag); enumStr != "" {
		parts = append(parts, "Valid values: "+enumStr)
	}

	// Add default value
	if defaultStr := c.getDefaultString(flag); defaultStr != "" {
		parts = append(parts, fmt.Sprintf("(default %s)", defaultStr))
	}

	// Add separator for slices
	if sepStr := c.getSeparatorString(flag); sepStr != "" {
		parts = append(parts, "Separator: "+sepStr)
	}

	// Add relationship constraints
	if reqStr := c.getRequiresString(flag); reqStr != "" {
		parts = append(parts, "Requires: "+reqStr)
	}

	if exclStr := c.getExcludesString(flag); exclStr != "" {
		parts = append(parts, "Excludes: "+exclStr)
	}

	return strings.Join(parts, ". ")
}

func (c *Cmd) getDefaultString(flag any) string {
	switch f := flag.(type) {
	case *StringFlag:
		if f.Default != nil {
			return *f.Default
		}
	case *IntFlag:
		if f.Default != nil {
			return fmt.Sprintf("%d", *f.Default)
		}
	case *Int64Flag:
		if f.Default != nil {
			return fmt.Sprintf("%d", *f.Default)
		}
	case *Float64Flag:
		if f.Default != nil {
			return fmt.Sprintf("%g", *f.Default)
		}
	case *BoolFlag:
		if f.Default != nil && *f.Default {
			return "true" // Show (default true) but not (default false)
		}
		return ""
	case *StringSliceFlag:
		if f.Default != nil && len(*f.Default) > 0 {
			return fmt.Sprintf("[%s]", strings.Join(*f.Default, ", "))
		}
	case *IntSliceFlag:
		if f.Default != nil && len(*f.Default) > 0 {
			var strs []string
			for _, v := range *f.Default {
				strs = append(strs, fmt.Sprintf("%d", v))
			}
			return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
		}
	case *Int64SliceFlag:
		if f.Default != nil && len(*f.Default) > 0 {
			var strs []string
			for _, v := range *f.Default {
				strs = append(strs, fmt.Sprintf("%d", v))
			}
			return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
		}
	case *Float64SliceFlag:
		if f.Default != nil && len(*f.Default) > 0 {
			var strs []string
			for _, v := range *f.Default {
				strs = append(strs, fmt.Sprintf("%g", v))
			}
			return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
		}
	case *BoolSliceFlag:
		if f.Default != nil && len(*f.Default) > 0 {
			var strs []string
			for _, v := range *f.Default {
				strs = append(strs, fmt.Sprintf("%t", v))
			}
			return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
		}
	}
	return ""
}

func (c *Cmd) getRangeString(flag any) string {
	switch f := flag.(type) {
	case *IntFlag:
		if f.min != nil || f.max != nil {
			var left, right string

			if f.min != nil {
				inclusive := f.minInclusive == nil || *f.minInclusive // default to inclusive
				if inclusive {
					left = fmt.Sprintf("[%d", *f.min)
				} else {
					left = fmt.Sprintf("(%d", *f.min)
				}
			} else {
				left = "("
			}

			if f.max != nil {
				inclusive := f.maxInclusive == nil || *f.maxInclusive // default to inclusive
				if inclusive {
					right = fmt.Sprintf("%d]", *f.max)
				} else {
					right = fmt.Sprintf("%d)", *f.max)
				}
			} else {
				right = ")"
			}

			return left + ", " + right
		}
	case *Int64Flag:
		if f.min != nil || f.max != nil {
			var left, right string

			if f.min != nil {
				inclusive := f.minInclusive == nil || *f.minInclusive // default to inclusive
				if inclusive {
					left = fmt.Sprintf("[%d", *f.min)
				} else {
					left = fmt.Sprintf("(%d", *f.min)
				}
			} else {
				left = "("
			}

			if f.max != nil {
				inclusive := f.maxInclusive == nil || *f.maxInclusive // default to inclusive
				if inclusive {
					right = fmt.Sprintf("%d]", *f.max)
				} else {
					right = fmt.Sprintf("%d)", *f.max)
				}
			} else {
				right = ")"
			}

			return left + ", " + right
		}
	case *Float64Flag:
		if f.min != nil || f.max != nil {
			var left, right string

			if f.min != nil {
				inclusive := f.minInclusive == nil || *f.minInclusive // default to inclusive
				if inclusive {
					left = fmt.Sprintf("[%g", *f.min)
				} else {
					left = fmt.Sprintf("(%g", *f.min)
				}
			} else {
				left = "("
			}

			if f.max != nil {
				inclusive := f.maxInclusive == nil || *f.maxInclusive // default to inclusive
				if inclusive {
					right = fmt.Sprintf("%g]", *f.max)
				} else {
					right = fmt.Sprintf("%g)", *f.max)
				}
			} else {
				right = ")"
			}

			return left + ", " + right
		}
	}
	return ""
}

func (c *Cmd) getEnumString(flag any) string {
	switch f := flag.(type) {
	case *StringFlag:
		if f.EnumConstraint != nil && len(*f.EnumConstraint) > 0 {
			return fmt.Sprintf("[%s]", strings.Join(*f.EnumConstraint, ", "))
		}
	}
	return ""
}

func (c *Cmd) getRegexString(flag any) string {
	switch f := flag.(type) {
	case *StringFlag:
		if f.RegexConstraint != nil {
			return f.RegexConstraint.String()
		}
	}
	return ""
}

func (c *Cmd) getSeparatorString(flag any) string {
	switch f := flag.(type) {
	case *StringSliceFlag:
		if f.Separator != nil {
			return fmt.Sprintf("\"%s\"", *f.Separator)
		}
	case *IntSliceFlag:
		if f.Separator != nil {
			return fmt.Sprintf("\"%s\"", *f.Separator)
		}
	case *Int64SliceFlag:
		if f.Separator != nil {
			return fmt.Sprintf("\"%s\"", *f.Separator)
		}
	case *Float64SliceFlag:
		if f.Separator != nil {
			return fmt.Sprintf("\"%s\"", *f.Separator)
		}
	case *BoolSliceFlag:
		if f.Separator != nil {
			return fmt.Sprintf("\"%s\"", *f.Separator)
		}
	}
	return ""
}

func (c *Cmd) getRequiresString(flag any) string {
	base := getBaseFlag(flag)
	if base != nil && base.Requires != nil && len(*base.Requires) > 0 {
		return strings.Join(*base.Requires, ", ")
	}
	return ""
}

func (c *Cmd) getExcludesString(flag any) string {
	base := getBaseFlag(flag)
	if base != nil && base.Excludes != nil && len(*base.Excludes) > 0 {
		return strings.Join(*base.Excludes, ", ")
	}
	return ""
}

func (c *Cmd) shouldFlagBeOptionalInSynopsis(flag any) bool {
	base := getBaseFlag(flag)

	// Check if flag has a default value (makes it optional)
	hasDefault := false
	switch f := flag.(type) {
	case *StringFlag:
		hasDefault = f.Default != nil
	case *IntFlag:
		hasDefault = f.Default != nil
	case *Int64Flag:
		hasDefault = f.Default != nil
	case *Float64Flag:
		hasDefault = f.Default != nil
	case *BoolFlag:
		hasDefault = f.Default != nil
	case *StringSliceFlag:
		hasDefault = f.Default != nil
	case *IntSliceFlag:
		hasDefault = f.Default != nil
	case *Int64SliceFlag:
		hasDefault = f.Default != nil
	case *Float64SliceFlag:
		hasDefault = f.Default != nil
	case *BoolSliceFlag:
		hasDefault = f.Default != nil
	}

	// Flag is optional if it has a default OR was explicitly set optional
	// For now, we'll use a heuristic: check the flag name against known optional flags
	// This is a workaround until we fix the default behavior
	if hasDefault {
		return true
	}

	// Special cases for positional-only flags that were explicitly set optional
	if base.PositionalOnly && (base.Name == "output-dir" || base.Name == "files") {
		return true
	}

	// Default to required for positional-only flags without defaults
	if base.PositionalOnly {
		return false
	}

	// For non-positional flags, use default value presence
	return hasDefault
}

func (c *Cmd) flagHasDefault(flag any) bool {
	switch f := flag.(type) {
	case *StringFlag:
		return f.Default != nil
	case *IntFlag:
		return f.Default != nil
	case *Int64Flag:
		return f.Default != nil
	case *Float64Flag:
		return f.Default != nil
	case *BoolFlag:
		return true // Boolean flags have implicit default of false
	case *StringSliceFlag:
		return f.Default != nil
	case *IntSliceFlag:
		return f.Default != nil
	case *Int64SliceFlag:
		return f.Default != nil
	case *Float64SliceFlag:
		return f.Default != nil
	case *BoolSliceFlag:
		return true // Boolean slice flags have implicit default of empty slice
	}
	return false
}
