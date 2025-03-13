package core

func TransformRslArgs(args []RslArg, transformer func(RslArg) string) []string {
	output := make([]string, len(args))
	for i, arg := range args {
		output[i] = transformer(arg)
	}
	return output
}
