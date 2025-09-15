package testing

import (
	"fmt"
	"testing"
)

func Test_Args_Variadic_BasicStringCollection(t *testing.T) {
	script := `
args:
	*files str

print("files: {files}")
`
	setupAndRunCode(t, script, "file1.txt", "file2.txt", "file3.txt", "--color=never")
	expected := `files: [ "file1.txt", "file2.txt", "file3.txt" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_BasicIntegerCollection(t *testing.T) {
	script := `
args:
	*numbers int

print("numbers: {numbers}")
`
	setupAndRunCode(t, script, "1", "2", "3", "42", "--color=never")
	expected := `numbers: [ 1, 2, 3, 42 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_BasicFloatCollection(t *testing.T) {
	script := `
args:
	*values float

print("values: {values}")
`
	setupAndRunCode(t, script, "1.5", "2.7", "3.14", "--color=never")
	expected := `values: [ 1.5, 2.7, 3.14 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_WithRequiredArg(t *testing.T) {
	script := `
args:
	command str
	*options str

print("command: {command}")
print("options: {options}")
`
	setupAndRunCode(t, script, "build", "opt1", "opt2", "opt3", "--color=never")
	expected := `command: build
options: [ "opt1", "opt2", "opt3" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_WithFlags(t *testing.T) {
	script := `
args:
	*files str
	verbose v bool

print("files: {files}")
print("verbose: {verbose}")
`
	setupAndRunCode(t, script, "file1.txt", "file2.txt", "--verbose", "--color=never")
	expected := `files: [ "file1.txt", "file2.txt" ]
verbose: true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_FlagsBeforeVariadic(t *testing.T) {
	script := `
args:
	*files str
	verbose v bool

print("files: {files}")
print("verbose: {verbose}")
`
	setupAndRunCode(t, script, "--verbose", "file1.txt", "file2.txt", "--color=never")
	expected := `files: [ "file1.txt", "file2.txt" ]
verbose: true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_MultipleVariadicSections(t *testing.T) {
	script := `
args:
	command str
	*options1 str
	flag f bool
	*options2 str

print("command: {command}")
print("options1: {options1}")
print("flag: {flag}")
print("options2: {options2}")
`
	setupAndRunCode(t, script, "cmd", "opt1", "opt2", "--flag", "opt3", "opt4", "--color=never")
	expected := `command: cmd
options1: [ "opt1", "opt2" ]
flag: true
options2: [ "opt3", "opt4" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_MultipleVariadicWithEmptyFirst(t *testing.T) {
	script := `
args:
	command str
	*options1 str
	flag f bool
	*options2 str

print("command: {command}")
print("options1: {options1}")
print("flag: {flag}")
print("options2: {options2}")
`
	setupAndRunCode(t, script, "cmd", "--flag", "opt3", "opt4", "--color=never")
	expected := `command: cmd
options1: [ "opt3", "opt4" ]
flag: true
options2: [ ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_MultipleVariadicWithEmptySecond(t *testing.T) {
	script := `
args:
	command str
	*options1 str
	flag f bool
	*options2 str

print("command: {command}")
print("options1: {options1}")
print("flag: {flag}")
print("options2: {options2}")
`
	setupAndRunCode(t, script, "cmd", "opt1", "opt2", "--flag", "--color=never")
	expected := `command: cmd
options1: [ "opt1", "opt2" ]
flag: true
options2: [ ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_ThreeVariadicSections(t *testing.T) {
	script := `
args:
	*section1 str
	flag1 f bool
	*section2 int
	flag2 g bool
	*section3 str

print("section1: {section1}")
print("flag1: {flag1}")
print("section2: {section2}")
print("flag2: {flag2}")
print("section3: {section3}")
`
	setupAndRunCode(t, script, "a", "b", "--flag1", "1", "2", "--flag2", "x", "y", "--color=never")
	expected := `section1: [ "a", "b" ]
flag1: true
section2: [ 1, 2 ]
flag2: true
section3: [ "x", "y" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_EmptyCollection(t *testing.T) {
	script := `
args:
	*files str
	verbose v bool

print("files: {files}")
print("verbose: {verbose}")
`
	setupAndRunCode(t, script, "--verbose", "--color=never")
	expected := `files: [ ]
verbose: true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_InvalidIntegerType(t *testing.T) {
	script := `
args:
	*numbers int
`
	setupAndRunCode(t, script, "1", "not-a-number", "3", "--color=never")
	expected := `invalid int64 value for numbers: not-a-number

Usage:
  TestCase [numbers...] [OPTIONS]

Script args:
      --numbers ints

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Args_Variadic_InvalidFloatType(t *testing.T) {
	script := `
args:
	*values float
`
	setupAndRunCode(t, script, "1.5", "invalid-float", "--color=never")
	expected := `invalid float64 value for values: invalid-float

Usage:
  TestCase [values...] [OPTIONS]

Script args:
      --values [floats...]

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Args_Variadic_FlagClustering(t *testing.T) {
	script := `
args:
	*options str
	verbose v bool
	debug d bool
	quiet q bool

print("options: {options}")
print("verbose: {verbose}")
print("debug: {debug}")
print("quiet: {quiet}")
`
	setupAndRunCode(t, script, "opt1", "opt2", "-vdq", "--color=never")
	expected := `options: [ "opt1", "opt2" ]
verbose: true
debug: true
quiet: true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_HelpGeneration(t *testing.T) {
	script := `
args:
	command str
	*options str
	verbose v bool
`
	setupAndRunCode(t, script, "-h", "--color=never")
	expected := `Usage:
  TestCase <command> [options...] [OPTIONS]

Script args:
      --command str
      --options [strs...]
  -v, --verbose
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertExitCode(t, 0)
}

func Test_Args_Variadic_HelpGenerationMultipleVariadic(t *testing.T) {
	script := `
args:
	command str
	*options1 str
	flag f bool
	*options2 str
`
	setupAndRunCode(t, script, "-h", "--color=never")
	expected := `Usage:
  TestCase <command> [options1...] [OPTIONS]

Script args:
      --command str
      --options1 [strs...]
      --options2 [strs...]
  -f, --flag
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertExitCode(t, 0)
}

func Test_Args_Variadic_WithDefaultValue(t *testing.T) {
	script := `
args:
	*files str = ["default.txt"]

print("files: {files}")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `files: [ "default.txt" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_OverrideDefaultValue(t *testing.T) {
	script := `
args:
	*files str = ["default.txt"]

print("files: {files}")
`
	setupAndRunCode(t, script, "custom1.txt", "custom2.txt", "--color=never")
	expected := `files: [ "custom1.txt", "custom2.txt" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_DefaultIntValues(t *testing.T) {
	script := `
args:
	*numbers int = [1, 2, 3]

print("numbers: {numbers}")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `numbers: [ 1, 2, 3 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_ComplexIntegration(t *testing.T) {
	script := `
args:
	command str
	*input_files str
	output_file o str = "output.txt"
	*compile_flags str
	verbose v bool
	threads t int = 1

print("command: {command}")
print("input_files: {input_files}")
print("output_file: {output_file}")
print("compile_flags: {compile_flags}")
print("verbose: {verbose}")
print("threads: {threads}")
`
	setupAndRunCode(t, script, "build", "src1.c", "src2.c", "--output-file", "result.bin", "opt1", "opt2", "--verbose", "--threads", "4", "--color=never")
	expected := `command: build
input_files: [ "src1.c", "src2.c" ]
output_file: result.bin
compile_flags: [ "opt1", "opt2" ]
verbose: true
threads: 4
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_LargeCollection(t *testing.T) {
	script := `
args:
	*items str

print("count: {len(items)}")
`
	// Generate a large list of arguments
	args := make([]string, 101) // 100 items + color flag
	for i := 0; i < 100; i++ {
		args[i] = fmt.Sprintf("item%d", i)
	}
	args[100] = "--color=never"

	setupAndRunCode(t, script, args...)
	expected := `count: 100
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_NestedWithOptionalArgs(t *testing.T) {
	script := `
args:
	command str
	*files str
	output o str = "out.txt"
	*excludes str
	verbose v bool

print("command: {command}")
print("files: {files}")
print("output: {output}")
print("excludes: {excludes}")
print("verbose: {verbose}")
`
	setupAndRunCode(t, script, "process", "file1.txt", "file2.txt", "--output", "result.txt", "*.tmp", "*.log", "--verbose", "--color=never")
	expected := `command: process
files: [ "file1.txt", "file2.txt" ]
output: result.txt
excludes: [ "*.tmp", "*.log" ]
verbose: true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_AllFlagsNoPositional(t *testing.T) {
	script := `
args:
	*files str
	verbose v bool
	debug d bool

print("files: {files}")
print("verbose: {verbose}")
print("debug: {debug}")
`
	setupAndRunCode(t, script, "--verbose", "--debug", "--color=never")
	expected := `files: [ ]
verbose: true
debug: true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Variadic_WithNullableOptional(t *testing.T) {
	script := `
args:
	*files str
	config c str?

print("files: {files}")
print("config: {config}")
`
	setupAndRunCode(t, script, "file1.txt", "file2.txt", "--color=never")
	expected := `files: [ "file1.txt", "file2.txt" ]
config: null
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
