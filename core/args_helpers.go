package core

func TransformRadArgs(args []RadArg, transformer func(RadArg) string) []string {
	output := make([]string, len(args))
	for i, arg := range args {
		output[i] = transformer(arg)
	}
	return output
}
