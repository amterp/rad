package core

import (
	"fmt"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

type CobraArg struct {
	Arg    ScriptArg
	value  interface{} //should be a pointer, e.g. *string . This is to allow cobra to set the value
	IsNull bool
}

func (c *CobraArg) IsString() bool {
	return c.Arg.Type == RslString
}

func (c *CobraArg) IsStringArray() bool {
	return c.Arg.Type == RslStringArray
}

func (c *CobraArg) IsInt() bool {
	return c.Arg.Type == RslInt
}

func (c *CobraArg) IsIntArray() bool {
	return c.Arg.Type == RslIntArray
}

func (c *CobraArg) IsFloat() bool {
	return c.Arg.Type == RslFloat
}

func (c *CobraArg) IsFloatArray() bool {
	return c.Arg.Type == RslFloatArray
}

func (c *CobraArg) IsBool() bool {
	return c.Arg.Type == RslBool
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

func (c *CobraArg) GetInt() int {
	return *c.value.(*int)
}

func (c *CobraArg) GetIntArray() []int {
	return *c.value.(*[]int)
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

// todo handle panics better
func (c *CobraArg) SetValue(arg string) {
	// do proper casting
	switch c.Arg.Type {
	case RslString:
		c.value = &arg
	case RslStringArray:
		// split on arg commas
		split := strings.Split(arg, ",")
		c.value = &split
	case RslInt:
		parsed, err := strconv.Atoi(arg)
		if err != nil {
			panic("AHH! NOT INT: " + arg)
		}
		c.value = &parsed
	case RslIntArray:
		// split on arg commas
		split := strings.Split(arg, ",")
		ints := make([]int, len(split))
		for i, v := range split {
			parsed, err := strconv.Atoi(v)
			if err != nil {
				panic("AHH! NOT INT: " + arg)
			}
			ints[i] = parsed
		}
		c.value = &ints
	case RslFloat:
		parsed, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			panic("AHH! NOT FLOAT: " + arg)
		}
		c.value = &parsed
	case RslFloatArray:
		// split on arg commas
		split := strings.Split(arg, ",")
		floats := make([]float64, len(split))
		for i, v := range split {
			parsed, err := strconv.ParseFloat(v, 64)
			if err != nil {
				panic("AHH! NOT FLOAT: " + arg)
			}
			floats[i] = parsed
		}
		c.value = &floats
	case RslBool:
		arg = strings.ToLower(arg)
		if arg == "true" || arg == "1" {
			val := true
			c.value = &val
		} else if arg == "false" || arg == "0" {
			val := false
			c.value = &val
		} else {
			panic("AHH! NOT BOOL: " + arg)
		}
	}
}

func CreateCobraArg(cmd *cobra.Command, arg ScriptArg) CobraArg {
	name, argType, flag, description := arg.Name, arg.Type, "", ""
	if arg.Flag != nil {
		flag = *arg.Flag
	}
	if arg.Description != nil {
		description = *arg.Description
	}

	var cobraArgValue interface{}
	switch argType {
	case RslString:
		defVal := ""
		if arg.DefaultString != nil {
			defVal = *arg.DefaultString
		}
		cobraArgValue = cmd.Flags().StringP(name, flag, defVal, description)
	case RslStringArray:
		var defVal []string
		if arg.DefaultStringArray != nil {
			defVal = *arg.DefaultStringArray
		}
		cobraArgValue = cmd.Flags().StringSliceP(name, flag, defVal, description)
	case RslInt:
		defVal := 0
		if arg.DefaultInt != nil {
			defVal = *arg.DefaultInt
		}
		cobraArgValue = cmd.Flags().IntP(name, flag, defVal, description)
	case RslIntArray:
		var defVal []int
		if arg.DefaultIntArray != nil {
			defVal = *arg.DefaultIntArray
		}
		cobraArgValue = cmd.Flags().IntSliceP(name, flag, defVal, description)
	case RslFloat:
		defVal := 0.0
		if arg.DefaultFloat != nil {
			defVal = *arg.DefaultFloat
		}
		cobraArgValue = cmd.Flags().Float64P(name, flag, defVal, description)
	case RslFloatArray:
		var defVal []float64
		if arg.DefaultFloatArray != nil {
			defVal = *arg.DefaultFloatArray
		}
		cobraArgValue = cmd.Flags().Float64SliceP(name, flag, defVal, description)
	case RslBool:
		defVal := false
		if arg.DefaultBool != nil {
			defVal = *arg.DefaultBool
		}
		cobraArgValue = cmd.Flags().BoolP(name, flag, defVal, description)
	default:
		// todo better error handling
		panic(fmt.Sprintf("Unknown arg type: %v", argType))
	}
	cobraArg := CobraArg{Arg: arg, value: cobraArgValue}
	return cobraArg
}
