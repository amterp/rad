package main

import "github.com/amterp/rad/core"

func main() {
	runner := core.NewRadRunner(core.RunnerInput{})
	runner.Run()
}
