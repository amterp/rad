package core

import (
	"fmt"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

type CobraArg struct {
	Arg    ScriptArg
	value  interface{} // should be a pointer, e.g. *string . This is to allow cobra to set the value
	IsNull bool
}

func (c *CobraArg) IsString() bool {
	return c.Arg.Type == ArgStringT
}

func (c *CobraArg) IsStringArray() bool {
	return c.Arg.Type == ArgStringArrayT
}

func (c *CobraArg) IsInt() bool {
	return c.Arg.Type == ArgIntT
}

func (c *CobraArg) IsIntArray() bool {
	return c.Arg.Type == ArgIntArrayT
}

func (c *CobraArg) IsFloat() bool {
	return c.Arg.Type == ArgFloatT
}

func (c *CobraArg) IsFloatArray() bool {
	return c.Arg.Type == ArgFloatArrayT
}

func (c *CobraArg) IsBool() bool {
	return c.Arg.Type == ArgBoolT
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
	} else if c.Arg.DefaultBoolArray != nil {
		c.value = c.Arg.DefaultBoolArray
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

func (c *CobraArg) GetBoolArray() []bool {
	return *c.value.(*[]bool)
}

func (c *CobraArg) GetMixedArray() []interface{} {
	return *c.value.(*[]interface{})
}

func (c *CobraArg) GetBool() bool {
	return *c.value.(*bool)
}

func (c *CobraArg) SetValue(arg string) {
	// do proper casting
	switch c.Arg.Type {
	case ArgStringT:
		c.value = &arg
	case ArgStringArrayT:
		// split on arg commas
		split := strings.Split(arg, ",")
		c.value = &split
	case ArgIntT:
		parsed, err := strconv.Atoi(arg)
		if err != nil {
			RP.TokenErrorExit(c.Arg.DeclarationToken, fmt.Sprintf("Expected int, but could not parse: %v\n", arg))
		}
		c.value = &parsed
	case ArgIntArrayT:
		// split on arg commas
		split := strings.Split(arg, ",")
		ints := make([]interface{}, len(split))
		for i, v := range split {
			parsed, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				RP.TokenErrorExit(c.Arg.DeclarationToken, fmt.Sprintf("Expected int, but could not parse: %v\n", arg))
			}
			ints[i] = parsed
		}
		c.value = &ints
	case ArgFloatT:
		parsed, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			RP.TokenErrorExit(c.Arg.DeclarationToken, fmt.Sprintf("Expected float, but could not parse: %v\n", arg))
		}
		c.value = &parsed
	case ArgFloatArrayT:
		// split on arg commas
		split := strings.Split(arg, ",")
		floats := make([]interface{}, len(split))
		for i, v := range split {
			parsed, err := strconv.ParseFloat(v, 64)
			if err != nil {
				RP.TokenErrorExit(c.Arg.DeclarationToken, fmt.Sprintf("Expected float, but could not parse: %v\n", arg))
			}
			floats[i] = parsed
		}
		c.value = &floats
	case ArgBoolT:
		arg = strings.ToLower(arg)
		if arg == "true" || arg == "1" {
			val := true
			c.value = &val
		} else if arg == "false" || arg == "0" {
			val := false
			c.value = &val
		} else {
			RP.TokenErrorExit(c.Arg.DeclarationToken, fmt.Sprintf("Expected bool, but could not parse: %v\n", arg))
		}
	case ArgBoolArrayT:
		// split on arg commas
		split := strings.Split(arg, ",")
		bools := make([]interface{}, len(split))
		for i, v := range split {
			v = strings.ToLower(v)
			if v == "true" || v == "1" {
				bools[i] = true
			} else if v == "false" || v == "0" {
				bools[i] = false
			} else {
				RP.TokenErrorExit(c.Arg.DeclarationToken, fmt.Sprintf("Expected bool, but could not parse: %v\n", arg))
			}
		}
	default:

	}
}

func CreateCobraArg(cmd *cobra.Command, arg ScriptArg) CobraArg {
	name, argType, flag, description := arg.ApiName, arg.Type, "", ""
	if arg.Flag != nil {
		flag = *arg.Flag
	}
	if arg.Description != nil {
		description = *arg.Description
	}

	var cobraArgValue interface{}
	switch argType {
	case ArgStringT:
		defVal := ""
		if arg.DefaultString != nil {
			defVal = *arg.DefaultString
		}
		cobraArgValue = cmd.Flags().StringP(name, flag, defVal, description)
	case ArgStringArrayT:
		var defVal []string
		if arg.DefaultStringArray != nil {
			defVal = *arg.DefaultStringArray
		}
		cobraArgValue = cmd.Flags().StringSliceP(name, flag, defVal, description)
	case ArgIntT:
		defVal := int64(0)
		if arg.DefaultInt != nil {
			defVal = *arg.DefaultInt
		}
		cobraArgValue = cmd.Flags().Int64P(name, flag, defVal, description)
	case ArgIntArrayT:
		var defVal []int64
		if arg.DefaultIntArray != nil {
			defVal = *arg.DefaultIntArray
		}
		cobraArgValue = cmd.Flags().Int64SliceP(name, flag, defVal, description)
	case ArgFloatT:
		defVal := 0.0
		if arg.DefaultFloat != nil {
			defVal = *arg.DefaultFloat
		}
		cobraArgValue = cmd.Flags().Float64P(name, flag, defVal, description)
	case ArgFloatArrayT:
		var defVal []float64
		if arg.DefaultFloatArray != nil {
			defVal = *arg.DefaultFloatArray
		}
		cobraArgValue = cmd.Flags().Float64SliceP(name, flag, defVal, description)
	case ArgBoolT:
		defVal := false
		if arg.DefaultBool != nil {
			defVal = *arg.DefaultBool
		}
		cobraArgValue = cmd.Flags().BoolP(name, flag, defVal, description)
	case ArgBoolArrayT:
		var defVal []bool
		if arg.DefaultBoolArray != nil {
			defVal = *arg.DefaultBoolArray
		}
		cobraArgValue = cmd.Flags().BoolSliceP(name, flag, defVal, description)
	default:
		RP.RadTokenErrorExit(arg.DeclarationToken, fmt.Sprintf("Unknown arg type: %v\n", argType))
	}
	cobraArg := CobraArg{Arg: arg, value: cobraArgValue}
	return cobraArg
}
