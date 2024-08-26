package core

type ScriptArg struct {
	Name        string
	Flag        *string
	Type        RslTypeEnum
	Description *string
	IsOptional  bool
	// first check the Type and IsOptional, then get the value
	DefaultString      *string
	DefaultStringArray *[]string
	DefaultInt         *int
	DefaultIntArray    *[]int
	DefaultBool        *bool
}
