package core

import (
	"fmt"
)

//
//func (c *CobraArg) InitializeOptional() {
//	if c.Arg.DefaultString != nil {
//		c.value = c.Arg.DefaultString
//	} else if c.Arg.DefaultStringArray != nil {
//		c.value = c.Arg.DefaultStringArray
//	} else if c.Arg.DefaultInt != nil {
//		c.value = c.Arg.DefaultInt
//	} else if c.Arg.DefaultIntArray != nil {
//		c.value = c.Arg.DefaultIntArray
//	} else if c.Arg.DefaultFloat != nil {
//		c.value = c.Arg.DefaultFloat
//	} else if c.Arg.DefaultFloatArray != nil {
//		c.value = c.Arg.DefaultFloatArray
//	} else if c.Arg.DefaultBool != nil {
//		c.value = c.Arg.DefaultBool
//	} else if c.Arg.DefaultBoolArray != nil {
//		c.value = c.Arg.DefaultBoolArray
//	} else {
//		c.IsNull = true
//	}
//}

func CreateFlag(arg ScriptArg) RslArg {
	apiName, argType, shorthand, description := arg.ApiName, arg.Type, "", ""
	if arg.Short != nil {
		shorthand = *arg.Short
	}
	if arg.Description != nil {
		description = *arg.Description
	}

	switch argType {
	case ArgStringT:
		defVal := ""
		if arg.DefaultString != nil {
			defVal = *arg.DefaultString
		}
		f := NewStringRadFlag(apiName, shorthand, "string", description, defVal)
		f.scriptArg = &arg
		f.Identifier = arg.Name
		return &f
	case ArgStringArrayT:
		var defVal []string
		if arg.DefaultStringArray != nil {
			defVal = *arg.DefaultStringArray
		}
		f := NewStringArrRadFlag(apiName, shorthand, "string,string", description, defVal)
		f.scriptArg = &arg
		f.Identifier = arg.Name
		return &f
	case ArgIntT:
		defVal := int64(0)
		if arg.DefaultInt != nil {
			defVal = *arg.DefaultInt
		}
		f := NewIntRadFlag(apiName, shorthand, "int", description, defVal)
		f.scriptArg = &arg
		f.Identifier = arg.Name
		return &f
	case ArgIntArrayT:
		var defVal []int64
		if arg.DefaultIntArray != nil {
			defVal = *arg.DefaultIntArray
		}
		f := NewIntArrRadFlag(apiName, shorthand, "int,int", description, defVal)
		f.scriptArg = &arg
		f.Identifier = arg.Name
		return &f
	case ArgFloatT:
		defVal := 0.0
		if arg.DefaultFloat != nil {
			defVal = *arg.DefaultFloat
		}
		f := NewFloatRadFlag(apiName, shorthand, "float", description, defVal)
		f.scriptArg = &arg
		f.Identifier = arg.Name
		return &f
	case ArgFloatArrayT:
		var defVal []float64
		if arg.DefaultFloatArray != nil {
			defVal = *arg.DefaultFloatArray
		}
		f := NewFloatArrRadFlag(apiName, shorthand, "float,float", description, defVal)
		f.scriptArg = &arg
		f.Identifier = arg.Name
		return &f
	case ArgBoolT:
		defVal := false
		if arg.DefaultBool != nil {
			defVal = *arg.DefaultBool
		}
		f := NewBoolRadFlag(apiName, shorthand, description, defVal)
		f.scriptArg = &arg
		f.Identifier = arg.Name
		return &f
	case ArgBoolArrayT:
		var defVal []bool
		if arg.DefaultBoolArray != nil {
			defVal = *arg.DefaultBoolArray
		}
		f := NewBoolArrRadFlag(apiName, shorthand, "bool,bool", description, defVal)
		f.scriptArg = &arg
		f.Identifier = arg.Name
		return &f
	default:
		RP.RadTokenErrorExit(arg.DeclarationToken, fmt.Sprintf("Unknown arg type: %v\n", argType))
		panic(UNREACHABLE)
	}
}
