package core

import (
	"fmt"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

type CobraArg struct {
	printer Printer
	Arg     ScriptArg
	value   interface{} //should be a pointer, e.g. *string . This is to allow cobra to set the value
	IsNull  bool
}

func (c *CobraArg) IsString() bool {
	return c.Arg.Type == RslStringT
}

func (c *CobraArg) IsStringArray() bool {
	return c.Arg.Type == RslStringArrayT
}

func (c *CobraArg) IsInt() bool {
	return c.Arg.Type == RslIntT
}

func (c *CobraArg) IsIntArray() bool {
	return c.Arg.Type == RslIntArrayT
}

func (c *CobraArg) IsFloat() bool {
	return c.Arg.Type == RslFloatT
}

func (c *CobraArg) IsFloatArray() bool {
	return c.Arg.Type == RslFloatArrayT
}

func (c *CobraArg) IsBool() bool {
	return c.Arg.Type == RslBoolT
}

func (c *CobraArg) InitializeOptional() {
	if c.Arg.DefaultString != nil {
		c.value = c.Arg.DefaultString
	} else if c.Arg.DefaultStringArray != nil {
		c.value = c.Arg.DefaultStringArray
	} else if c.Arg.DefaultInt != nil {
		c.value = c.Arg.DefaultInt
	} else if c.Arg.DefaultIntArray != nil {
		c.value = c.Arg.DefaultIntArray
	} else if c.Arg.DefaultFloat != nil {
		c.value = c.Arg.DefaultFloat
	} else if c.Arg.DefaultFloatArray != nil {
		c.value = c.Arg.DefaultFloatArray
	} else if c.Arg.DefaultBool != nil {
		c.value = c.Arg.DefaultBool
	} else {
		c.IsNull = true
	}
}

func (c *CobraArg) GetString() string {
	return *c.value.(*string)
}

func (c *CobraArg) GetStringArray() []string {
	return *c.value.(*[]string)
}

func (c *CobraArg) GetInt() int64 {
	return *c.value.(*int64)
}

func (c *CobraArg) GetIntArray() []int64 {
	return *c.value.(*[]int64)
}

func (c *CobraArg) GetFloat() float64 {
	return *c.value.(*float64)
}

func (c *CobraArg) GetFloatArray() []float64 {
	return *c.value.(*[]float64)
}

func (c *CobraArg) GetBool() bool {
	return *c.value.(*bool)
}

func (c *CobraArg) SetValue(arg string) {
	// do proper casting
	switch c.Arg.Type {
	case RslStringT:
		c.value = &arg
	case RslStringArrayT:
		// split on arg commas
		split := strings.Split(arg, ",")
		c.value = &split
	case RslIntT:
		parsed, err := strconv.Atoi(arg)
		if err != nil {
			c.printer.TokenErrorExit(c.Arg.DeclarationToken, fmt.Sprintf("Expected int, but could not parse: %v\n", arg))
		}
		c.value = &parsed
	case RslIntArrayT:
		// split on arg commas
		split := strings.Split(arg, ",")
		ints := make([]int64, len(split))
		for i, v := range split {
			parsed, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				c.printer.TokenErrorExit(c.Arg.DeclarationToken, fmt.Sprintf("Expected int, but could not parse: %v\n", arg))
			}
			ints[i] = parsed
		}
		c.value = &ints
	case RslFloatT:
		parsed, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			c.printer.TokenErrorExit(c.Arg.DeclarationToken, fmt.Sprintf("Expected float, but could not parse: %v\n", arg))
		}
		c.value = &parsed
	case RslFloatArrayT:
		// split on arg commas
		split := strings.Split(arg, ",")
		floats := make([]float64, len(split))
		for i, v := range split {
			parsed, err := strconv.ParseFloat(v, 64)
			if err != nil {
				c.printer.TokenErrorExit(c.Arg.DeclarationToken, fmt.Sprintf("Expected float, but could not parse: %v\n", arg))
			}
			floats[i] = parsed
		}
		c.value = &floats
	case RslBoolT:
		arg = strings.ToLower(arg)
		if arg == "true" || arg == "1" {
			val := true
			c.value = &val
		} else if arg == "false" || arg == "0" {
			val := false
			c.value = &val
		} else {
			c.printer.TokenErrorExit(c.Arg.DeclarationToken, fmt.Sprintf("Expected bool, but could not parse: %v\n", arg))
		}
	}
}

func CreateCobraArg(printer Printer, cmd *cobra.Command, arg ScriptArg) CobraArg {
	name, argType, flag, description := arg.ApiName, arg.Type, "", ""
	if arg.Flag != nil {
		flag = *arg.Flag
	}
	if arg.Description != nil {
		description = *arg.Description
	}

	var cobraArgValue interface{}
	switch argType {
	case RslStringT:
		defVal := ""
		if arg.DefaultString != nil {
			defVal = *arg.DefaultString
		}
		cobraArgValue = cmd.Flags().StringP(name, flag, defVal, description)
	case RslStringArrayT:
		var defVal []string
		if arg.DefaultStringArray != nil {
			defVal = *arg.DefaultStringArray
		}
		cobraArgValue = cmd.Flags().StringSliceP(name, flag, defVal, description)
	case RslIntT:
		defVal := int64(0)
		if arg.DefaultInt != nil {
			defVal = *arg.DefaultInt
		}
		cobraArgValue = cmd.Flags().Int64P(name, flag, defVal, description)
	case RslIntArrayT:
		var defVal []int64
		if arg.DefaultIntArray != nil {
			defVal = *arg.DefaultIntArray
		}
		cobraArgValue = cmd.Flags().Int64SliceP(name, flag, defVal, description)
	case RslFloatT:
		defVal := 0.0
		if arg.DefaultFloat != nil {
			defVal = *arg.DefaultFloat
		}
		cobraArgValue = cmd.Flags().Float64P(name, flag, defVal, description)
	case RslFloatArrayT:
		var defVal []float64
		if arg.DefaultFloatArray != nil {
			defVal = *arg.DefaultFloatArray
		}
		cobraArgValue = cmd.Flags().Float64SliceP(name, flag, defVal, description)
	case RslBoolT:
		defVal := false
		if arg.DefaultBool != nil {
			defVal = *arg.DefaultBool
		}
		cobraArgValue = cmd.Flags().BoolP(name, flag, defVal, description)
	default:
		printer.RadTokenErrorExit(arg.DeclarationToken, fmt.Sprintf("Unknown arg type: %v\n", argType))
	}
	cobraArg := CobraArg{printer: printer, Arg: arg, value: cobraArgValue}
	return cobraArg
}
