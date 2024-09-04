package core

import (
	"strconv"
	"strings"
)

type CobraArg struct {
	Arg   ScriptArg
	value interface{} //should be a pointer, e.g. *string . This is to allow cobra to set the value
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

func (c *CobraArg) SetDefaultIfPresent() {
	if c.Arg.DefaultString != nil {
		c.value = c.Arg.DefaultString
	}
	if c.Arg.DefaultStringArray != nil {
		c.value = c.Arg.DefaultStringArray
	}
	if c.Arg.DefaultInt != nil {
		c.value = c.Arg.DefaultInt
	}
	if c.Arg.DefaultIntArray != nil {
		c.value = c.Arg.DefaultIntArray
	}
	if c.Arg.DefaultFloat != nil {
		c.value = c.Arg.DefaultFloat
	}
	if c.Arg.DefaultFloatArray != nil {
		c.value = c.Arg.DefaultFloatArray
	}
	if c.Arg.DefaultBool != nil {
		c.value = c.Arg.DefaultBool
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
			panic("AHH! NOT INT!")
		}
		c.value = &parsed
	case RslIntArray:
		// split on arg commas
		split := strings.Split(arg, ",")
		ints := make([]int, len(split))
		for i, v := range split {
			parsed, err := strconv.Atoi(v)
			if err != nil {
				panic("AHH! NOT INT!")
			}
			ints[i] = parsed
		}
		c.value = &ints
	case RslFloat:
		parsed, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			panic("AHH! NOT FLOAT!")
		}
		c.value = &parsed
	case RslFloatArray:
		// split on arg commas
		split := strings.Split(arg, ",")
		floats := make([]float64, len(split))
		for i, v := range split {
			parsed, err := strconv.ParseFloat(v, 64)
			if err != nil {
				panic("AHH! NOT FLOAT!")
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
			panic("AHH! NOT BOOL!")
		}
	}
}
