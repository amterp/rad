package cmd

import (
	"rad/core"
	"strconv"
	"strings"
)

type CobraArg struct {
	Arg   core.ScriptArg
	value interface{}
}

func (c *CobraArg) IsString() bool {
	return c.Arg.Type == core.RslString
}

func (c *CobraArg) IsStringArray() bool {
	return c.Arg.Type == core.RslStringArray
}

func (c *CobraArg) IsInt() bool {
	return c.Arg.Type == core.RslInt
}

func (c *CobraArg) IsIntArray() bool {
	return c.Arg.Type == core.RslIntArray
}

func (c *CobraArg) IsBool() bool {
	return c.Arg.Type == core.RslBool
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

func (c *CobraArg) GetBool() bool {
	return *c.value.(*bool)
}

// todo handle panics better
func (c *CobraArg) SetValue(arg string) {
	// do proper casting
	switch c.Arg.Type {
	case core.RslString:
		c.value = &arg
	case core.RslStringArray:
		// split on arg commas
		split := strings.Split(arg, ",")
		c.value = &split
	case core.RslInt:
		parsed, err := strconv.Atoi(arg)
		if err != nil {
			panic("AHH! NOT INT!")
		}
		c.value = parsed
	case core.RslIntArray:
		// split on arg commas
		split := strings.Split(arg, ",")
		ints := make([]int, len(split))
		for i, s := range split {
			parsed, err := strconv.Atoi(s)
			if err != nil {
				panic("AHH! NOT INT!")
			}
			ints[i] = parsed
		}
		c.value = &ints
	case core.RslBool:
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
