package ra

import (
	"regexp"
	"strings"
	"testing"

	"github.com/amterp/color"

	"github.com/stretchr/testify/assert"
)

func init() {
	color.NoColor = true
}

func Test_Usage_SimpleCommand(t *testing.T) {
	cmd := NewCmd("myapp")
	cmd.SetDescription("A simple application for testing")

	flag1, err := NewString("config").
		SetShort("c").
		SetUsage("Configuration file path.").
		SetDefault("/etc/myapp.conf").
		Register(cmd)
	assert.NoError(t, err)

	flag2, err := NewBool("verbose").
		SetShort("v").
		SetUsage("Enable verbose output.").
		Register(cmd)
	assert.NoError(t, err)

	flag3, err := NewInt("timeout").
		SetUsage("Request timeout in seconds.").
		SetMin(1, true).
		SetMax(300, true).
		SetDefault(30).
		Register(cmd)
	assert.NoError(t, err)

	_ = flag1
	_ = flag2
	_ = flag3

	usage := cmd.GenerateUsage(false)

	expected := `A simple application for testing

Usage:
  myapp [config] [timeout] [OPTIONS]

Arguments:
  -c, --config str    Configuration file path. (default /etc/myapp.conf)
      --timeout int   Request timeout in seconds. Range: [1, 300]. (default 30)
  -v, --verbose       Enable verbose output.
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(usage))
}

func Test_Usage_WithPositionalArgs(t *testing.T) {
	cmd := NewCmd("processor")
	cmd.SetDescription("Process input files")

	// Positional arguments
	input, err := NewString("input-file").
		SetUsage("Path to input file").
		SetPositionalOnly(true).
		Register(cmd)
	assert.NoError(t, err)

	output, err := NewString("output-dir").
		SetUsage("Output directory").
		SetPositionalOnly(true).
		SetOptional(true).
		Register(cmd)
	assert.NoError(t, err)

	// Named options
	format, err := NewString("format").
		SetShort("f").
		SetUsage("Output format").
		SetEnumConstraint([]string{"json", "yaml", "xml"}).
		SetDefault("json").
		Register(cmd)
	assert.NoError(t, err)

	concurrent, err := NewInt("jobs").
		SetShort("j").
		SetUsage("Number of concurrent jobs").
		SetMin(1, true).
		SetMax(16, false).
		Register(cmd)
	assert.NoError(t, err)

	_ = input
	_ = output
	_ = format
	_ = concurrent

	usage := cmd.GenerateUsage(false)

	expected := `Process input files

Usage:
  processor <input-file> [output-dir] [format] <jobs> [OPTIONS]

Arguments:
  input-file str     Path to input file
  output-dir str     (optional) Output directory
  -f, --format str   Output format Valid values: [json, yaml, xml]. (default json)
  -j, --jobs int     (required) Number of concurrent jobs Range: [1, 16)
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(usage))
}

func Test_Usage_WithSubcommands(t *testing.T) {
	cmd := NewCmd("git-like")
	cmd.SetDescription("A git-like version control system")

	// Global options
	verbose, err := NewBool("verbose").
		SetShort("v").
		SetUsage("Enable verbose output").
		Register(cmd)
	assert.NoError(t, err)

	config, err := NewString("config").
		SetUsage("Path to config file").
		SetDefault("~/.gitlike/config").
		Register(cmd)
	assert.NoError(t, err)

	// Add subcommand
	addCmd := NewCmd("add")
	addCmd.SetDescription("Add files to the staging area")

	addAll, err := NewBool("all").
		SetShort("A").
		SetUsage("Add all modified files").
		Register(addCmd)
	assert.NoError(t, err)

	addFiles, err := NewStringSlice("files").
		SetUsage("Files to add").
		SetPositionalOnly(true).
		SetVariadic(true).
		Register(addCmd)
	assert.NoError(t, err)

	// Commit subcommand
	commitCmd := NewCmd("commit")
	commitCmd.SetDescription("Record changes to the repository")

	message, err := NewString("message").
		SetShort("m").
		SetUsage("Commit message").
		SetRegexConstraint(regexp.MustCompile(`^.{1,72}$`)).
		Register(commitCmd)
	assert.NoError(t, err)

	amend, err := NewBool("amend").
		SetUsage("Amend the previous commit").
		SetExcludes([]string{"message"}).
		Register(commitCmd)
	assert.NoError(t, err)

	_, err = cmd.RegisterCmd(addCmd)
	assert.NoError(t, err)
	_, err = cmd.RegisterCmd(commitCmd)
	assert.NoError(t, err)

	_ = verbose
	_ = config
	_ = addAll
	_ = addFiles
	_ = message
	_ = amend

	usage := cmd.GenerateUsage(false)

	expected := `A git-like version control system

Usage:
  git-like [subcommand] [config] [OPTIONS]

Commands:
  add                           Add files to the staging area
  commit                        Record changes to the repository

Arguments:
      --config str   Path to config file (default ~/.gitlike/config)
  -v, --verbose      Enable verbose output
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(usage))
}

func Test_Usage_SubcommandWithArgs(t *testing.T) {
	cmd := NewCmd("git-like")
	addCmd := NewCmd("add")
	addCmd.SetDescription("Add files to the staging area")

	// Add subcommand options
	all, err := NewBool("all").
		SetShort("A").
		SetUsage("Add all modified files").
		SetExcludes([]string{"force"}).
		Register(addCmd)
	assert.NoError(t, err)

	force, err := NewBool("force").
		SetShort("f").
		SetUsage("Force add ignored files").
		Register(addCmd)
	assert.NoError(t, err)

	files, err := NewStringSlice("files").
		SetUsage("Files to add").
		SetPositionalOnly(true).
		SetVariadic(true).
		SetOptional(true).
		Register(addCmd)
	assert.NoError(t, err)

	_, err = cmd.RegisterCmd(addCmd)
	assert.NoError(t, err)

	_ = all
	_ = force
	_ = files

	usage := addCmd.GenerateUsage(false)

	expected := `Add files to the staging area

Usage:
  add [files...] [OPTIONS]

Arguments:
  files [strs...]   Files to add
  -A, --all         Add all modified files Excludes: force
  -f, --force       Force add ignored files
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(usage))
}

func Test_Usage_ComplexConstraints(t *testing.T) {
	cmd := NewCmd("complex")
	cmd.SetDescription("Command with various constraint types")

	// Range constraints
	port, err := NewInt("port").
		SetShort("p").
		SetUsage("Server port").
		SetMin(1024, true).
		SetMax(65535, true).
		SetDefault(8080).
		Register(cmd)
	assert.NoError(t, err)

	rate, err := NewFloat64("rate").
		SetUsage("Processing rate").
		SetMin(0.0, false).
		SetMax(100.0, true).
		Register(cmd)
	assert.NoError(t, err)

	unbounded, err := NewInt("retries").
		SetUsage("Number of retries").
		SetMin(0, true).
		Register(cmd)
	assert.NoError(t, err)

	// Enum constraint
	format, err := NewString("format").
		SetUsage("Output format").
		SetEnumConstraint([]string{"json", "yaml", "xml", "csv"}).
		SetDefault("json").
		Register(cmd)
	assert.NoError(t, err)

	// Regex constraint
	name, err := NewString("name").
		SetUsage("Resource name").
		SetRegexConstraint(regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)).
		Register(cmd)
	assert.NoError(t, err)

	// Slice with custom separator
	tags, err := NewStringSlice("tags").
		SetUsage("Resource tags").
		SetSeparator("|").
		SetRequires([]string{"name"}).
		Register(cmd)
	assert.NoError(t, err)

	// Multiple relationships
	debug, err := NewBool("debug").
		SetUsage("Enable debug mode").
		SetExcludes([]string{"quiet", "silent"}).
		Register(cmd)
	assert.NoError(t, err)

	quiet, err := NewBool("quiet").
		SetShort("q").
		SetUsage("Suppress output").
		SetExcludes([]string{"debug", "verbose"}).
		Register(cmd)
	assert.NoError(t, err)

	_ = port
	_ = rate
	_ = unbounded
	_ = format
	_ = name
	_ = tags
	_ = debug
	_ = quiet

	usage := cmd.GenerateUsage(false)

	expected := `Command with various constraint types

Usage:
  complex [port] <rate> <retries> [format] <name> <tags> [OPTIONS]

Arguments:
  -p, --port int      Server port Range: [1024, 65535]. (default 8080)
      --rate float    (required) Processing rate Range: (0, 100]
      --retries int   (required) Number of retries Range: [0, )
      --format str    Output format Valid values: [json, yaml, xml, csv]. (default json)
      --name str      (required) Resource name Must match pattern: ^[a-zA-Z][a-zA-Z0-9_-]*$
      --tags strs     (required) Resource tags Separator: "|". Requires: name
      --debug         Enable debug mode Excludes: quiet, silent
  -q, --quiet         Suppress output Excludes: debug, verbose
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(usage))
}

func Test_Usage_MixedPositionalAndOptions(t *testing.T) {
	cmd := NewCmd("deploy")
	cmd.SetDescription("Deploy application to target environment")

	// Required positional
	app, err := NewString("app-name").
		SetUsage("Application name").
		SetPositionalOnly(true).
		Register(cmd)
	assert.NoError(t, err)

	// Optional positional
	env, err := NewString("environment").
		SetUsage("Target environment").
		SetPositionalOnly(true).
		SetOptional(true).
		SetDefault("staging").
		Register(cmd)
	assert.NoError(t, err)

	// Named options
	force, err := NewBool("force").
		SetShort("f").
		SetUsage("Force deployment even if checks fail").
		Register(cmd)
	assert.NoError(t, err)

	timeout, err := NewInt("timeout").
		SetUsage("Deployment timeout").
		SetMin(30, true).
		SetDefault(300).
		Register(cmd)
	assert.NoError(t, err)

	rollback, err := NewBool("no-rollback").
		SetUsage("Disable automatic rollback on failure").
		SetExcludes([]string{"force"}).
		Register(cmd)
	assert.NoError(t, err)

	_ = app
	_ = env
	_ = force
	_ = timeout
	_ = rollback

	usage := cmd.GenerateUsage(false)

	expected := `Deploy application to target environment

Usage:
  deploy <app-name> [environment] [timeout] [OPTIONS]

Arguments:
  app-name str        Application name
  environment str     (optional) Target environment (default staging)
      --timeout int   Deployment timeout Range: [30, ). (default 300)
  -f, --force         Force deployment even if checks fail
      --no-rollback   Disable automatic rollback on failure Excludes: force
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(usage))
}

func Test_Usage_HiddenAndLongHelpFlags(t *testing.T) {
	cmd := NewCmd("myapp")
	cmd.SetDescription("Testing hidden flag behavior")

	// Regular flag
	regular, err := NewString("config").
		SetUsage("Configuration file").
		Register(cmd)
	assert.NoError(t, err)

	// Hidden flag - should never appear
	hidden, err := NewString("secret").
		SetUsage("Secret flag").
		SetHidden(true).
		Register(cmd)
	assert.NoError(t, err)

	// Hidden in short help - should only appear in long help
	longOnly, err := NewString("debug-info").
		SetUsage("Debug information").
		SetHiddenInShortHelp(true).
		Register(cmd)
	assert.NoError(t, err)

	_ = regular
	_ = hidden
	_ = longOnly

	// Test short help (should exclude longOnly flag)
	shortUsage := cmd.GenerateUsage(false)
	expectedShort := `Testing hidden flag behavior

Usage:
  myapp <config> [OPTIONS]

Arguments:
      --config str   (required) Configuration file
`

	assert.Equal(t, strings.TrimSpace(expectedShort), strings.TrimSpace(shortUsage))

	// Test long help (should include longOnly flag)
	longUsage := cmd.GenerateUsage(true)
	expectedLong := `Testing hidden flag behavior

Usage:
  myapp <config> <debug-info> [OPTIONS]

Arguments:
      --config str       (required) Configuration file
      --debug-info str   (required) Debug information
`

	assert.Equal(t, strings.TrimSpace(expectedLong), strings.TrimSpace(longUsage))
}

func Test_Usage_GlobalFlagsInSubCommand(t *testing.T) {
	cmd := NewCmd("parent")
	cmd.SetDescription("Parent command with global flags")

	// Global flags
	verbose, err := NewBool("verbose").
		SetShort("v").
		SetUsage("Enable verbose output").
		Register(cmd, WithGlobal(true))
	assert.NoError(t, err)

	debug, err := NewString("log-level").
		SetUsage("Set log level").
		SetDefault("info").
		Register(cmd, WithGlobal(true))
	assert.NoError(t, err)

	// Create subcommand
	subCmd := NewCmd("sub")
	subCmd.SetDescription("Sub command inherits global flags")

	// Subcommand-specific flag
	input, err := NewString("input").
		SetUsage("Input file").
		Register(subCmd)
	assert.NoError(t, err)

	_, err = cmd.RegisterCmd(subCmd)
	assert.NoError(t, err)

	_ = verbose
	_ = debug
	_ = input

	// Subcommand should show both its own flags and global flags
	usage := subCmd.GenerateUsage(false)
	expected := `Sub command inherits global flags

Usage:
  sub <input> [OPTIONS]

Arguments:
      --input str   (required) Input file

Global options:
  -v, --verbose         Enable verbose output
      --log-level str   Set log level (default info)
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(usage))
}

func Test_Usage_FlagOrderingNonBoolFirst(t *testing.T) {
	cmd := NewCmd("ordering")
	cmd.SetDescription("Test flag ordering: non-bool first, then bool")

	// Register in mixed order to test sorting
	bool1, err := NewBool("verbose").
		SetShort("v").
		SetUsage("Enable verbose mode").
		Register(cmd)
	assert.NoError(t, err)

	str1, err := NewString("config").
		SetUsage("Config file path").
		Register(cmd)
	assert.NoError(t, err)

	bool2, err := NewBool("force").
		SetShort("f").
		SetUsage("Force operation").
		Register(cmd)
	assert.NoError(t, err)

	int1, err := NewInt("timeout").
		SetUsage("Timeout value").
		Register(cmd)
	assert.NoError(t, err)

	bool3, err := NewBool("quiet").
		SetShort("q").
		SetUsage("Quiet mode").
		Register(cmd)
	assert.NoError(t, err)

	_ = bool1
	_ = str1
	_ = bool2
	_ = int1
	_ = bool3

	usage := cmd.GenerateUsage(false)
	expected := `Test flag ordering: non-bool first, then bool

Usage:
  ordering <config> <timeout> [OPTIONS]

Arguments:
      --config str    (required) Config file path
      --timeout int   (required) Timeout value
  -v, --verbose       Enable verbose mode
  -f, --force         Force operation
  -q, --quiet         Quiet mode
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(usage))
}

func Test_Usage_VariadicSliceFormats(t *testing.T) {
	cmd := NewCmd("slices")
	cmd.SetDescription("Testing various slice configurations")

	// Positional-only variadic slice (must come first)
	files, err := NewStringSlice("files").
		SetUsage("Input files").
		SetPositionalOnly(true).
		SetVariadic(true).
		SetOptional(true).
		Register(cmd)
	assert.NoError(t, err)

	// Required variadic slice
	tags, err := NewStringSlice("tags").
		SetUsage("Resource tags").
		SetVariadic(true).
		Register(cmd)
	assert.NoError(t, err)

	// Optional variadic slice
	includes, err := NewStringSlice("include").
		SetUsage("Include patterns").
		SetVariadic(true).
		SetOptional(true).
		Register(cmd)
	assert.NoError(t, err)

	// Regular slice with separator
	ports, err := NewIntSlice("ports").
		SetUsage("Port numbers").
		SetSeparator(",").
		Register(cmd)
	assert.NoError(t, err)

	_ = tags
	_ = includes
	_ = ports
	_ = files

	usage := cmd.GenerateUsage(false)
	expected := `Testing various slice configurations

Usage:
  slices [files...] [OPTIONS]

Arguments:
  files [strs...]           Input files
      --tags strs...        Resource tags
      --include [strs...]   Include patterns
      --ports ints          (required) Port numbers Separator: ","
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(usage))
}

func Test_Usage_ComplexRelationshipsAndConstraints(t *testing.T) {
	cmd := NewCmd("complex")
	cmd.SetDescription("Complex example with all constraint types")

	// String with regex and enum (should show regex since it's more specific)
	mode, err := NewString("mode").
		SetUsage("Operation mode").
		SetRegexConstraint(regexp.MustCompile(`^(dev|prod|test)$`)).
		SetEnumConstraint([]string{"dev", "prod", "test"}).
		Register(cmd)
	assert.NoError(t, err)

	// Float with exclusive bounds
	rate, err := NewFloat64("rate").
		SetUsage("Processing rate").
		SetMin(0.0, false).
		SetMax(1.0, false).
		SetDefault(0.5).
		Register(cmd)
	assert.NoError(t, err)

	// Multiple requires/excludes
	verbose, err := NewBool("verbose").
		SetShort("v").
		SetUsage("Verbose output").
		SetExcludes([]string{"quiet", "silent"}).
		SetRequires([]string{"mode"}).
		Register(cmd)
	assert.NoError(t, err)

	quiet, err := NewBool("quiet").
		SetShort("q").
		SetUsage("Quiet mode").
		SetExcludes([]string{"verbose"}).
		Register(cmd)
	assert.NoError(t, err)

	silent, err := NewBool("silent").
		SetUsage("Silent mode").
		SetExcludes([]string{"verbose"}).
		Register(cmd)
	assert.NoError(t, err)

	// Slice with separator and requires
	configs, err := NewStringSlice("config-files").
		SetUsage("Configuration files").
		SetSeparator(":").
		SetRequires([]string{"mode"}).
		SetOptional(true).
		SetVariadic(true).
		Register(cmd)
	assert.NoError(t, err)

	_ = mode
	_ = rate
	_ = verbose
	_ = quiet
	_ = silent
	_ = configs

	usage := cmd.GenerateUsage(false)
	expected := `Complex example with all constraint types

Usage:
  complex <mode> [rate] [config-files...] [OPTIONS]

Arguments:
      --mode str                 (required) Operation mode Must match pattern: ^(dev|prod|test)$
      --rate float               Processing rate Range: (0, 1). (default 0.5)
      --config-files [strs...]   Configuration files Separator: ":". Requires: mode
  -v, --verbose                  Verbose output Requires: mode. Excludes: quiet, silent
  -q, --quiet                    Quiet mode Excludes: verbose
      --silent                   Silent mode Excludes: verbose
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(usage))
}

func Test_Usage_PositionalOnlyAfterVariadicError(t *testing.T) {
	cmd := NewCmd("test")

	// Register a variadic positional flag first
	_, err := NewStringSlice("files").
		SetUsage("Input files").
		SetVariadic(true).
		Register(cmd)
	assert.NoError(t, err)

	// Attempting to register a positional-only flag after should error
	_, err = NewString("output").
		SetUsage("Output file").
		SetPositionalOnly(true).
		Register(cmd)
	assert.Error(t, err)
	assert.Contains(
		t,
		err.Error(),
		"cannot register positional-only flag \"output\" after variadic positional flag \"files\"",
	)
}

func Test_Usage_SynopsisStopsAfterVariadic(t *testing.T) {
	cmd := NewCmd("test")
	cmd.SetDescription("Test synopsis stopping after variadic")

	// Register flags in order: regular, variadic, another regular
	_, err := NewString("first").
		SetUsage("First arg").
		Register(cmd)
	assert.NoError(t, err)

	_, err = NewStringSlice("variadic").
		SetUsage("Variadic arg").
		SetVariadic(true).
		Register(cmd)
	assert.NoError(t, err)

	_, err = NewString("after").
		SetUsage("After variadic").
		Register(cmd)
	assert.NoError(t, err)

	usage := cmd.GenerateUsage(false)
	expected := `Test synopsis stopping after variadic

Usage:
  test <first> [variadic...] [OPTIONS]

Arguments:
      --first str          (required) First arg
      --variadic strs...   Variadic arg
      --after str          (required) After variadic`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(usage))
}

func Test_Usage_RequiredFlagMarker(t *testing.T) {
	cmd := NewCmd("test")
	cmd.SetDescription("Test required flag marker")

	// Required flag (no default, not optional)
	_, err := NewString("required").
		SetUsage("Required flag").
		SetFlagOnly(true).
		Register(cmd)
	assert.NoError(t, err)

	// Optional flag
	_, err = NewString("optional").
		SetUsage("Optional flag").
		SetFlagOnly(true).
		SetOptional(true).
		Register(cmd)
	assert.NoError(t, err)

	// Flag with default (effectively optional)
	_, err = NewString("with-default").
		SetUsage("Flag with default").
		SetFlagOnly(true).
		SetDefault("defaultval").
		Register(cmd)
	assert.NoError(t, err)

	usage := cmd.GenerateUsage(false)
	expected := `Test required flag marker

Usage:
  test <required> [with-default] [OPTIONS]

Arguments:
      --required str       (required) Required flag
      --optional str       (optional) Optional flag
      --with-default str   Flag with default (default defaultval)`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(usage))
}

func Test_BoolFlag_DefaultFalse_NoDisplay(t *testing.T) {
	cmd := NewCmd("test")
	cmd.SetDescription("Test bool flag with default false behavior")

	// Explicit default false
	verbose, err := NewBool("verbose").
		SetShort("v").
		SetUsage("Enable verbose output").
		SetDefault(false).
		Register(cmd)
	assert.NoError(t, err)

	// Implicit default false (no SetDefault call)
	debug, err := NewBool("debug").
		SetUsage("Enable debug mode").
		Register(cmd)
	assert.NoError(t, err)

	_ = verbose
	_ = debug

	usage := cmd.GenerateUsage(false)

	// Should not contain "(default false)" anywhere
	assert.NotContains(t, usage, "(default false)")

	// But should still contain the descriptions
	assert.Contains(t, usage, "Enable verbose output")
	assert.Contains(t, usage, "Enable debug mode")
}

func Test_BoolFlag_DefaultTrue_ShowsDefault(t *testing.T) {
	cmd := NewCmd("test")
	cmd.SetDescription("Test bool flag with default true shows (default true)")

	// Explicit default true
	enabled, err := NewBool("enabled").
		SetShort("e").
		SetUsage("Feature is enabled by default").
		SetDefault(true).
		Register(cmd)
	assert.NoError(t, err)

	_ = enabled

	usage := cmd.GenerateUsage(false)

	// Should contain "(default true)"
	assert.Contains(t, usage, "(default true)")
	assert.Contains(t, usage, "Feature is enabled by default")

	// Should NOT contain "(default false)"
	assert.NotContains(t, usage, "(default false)")
}

func Test_BoolFlag_MixedDefaults(t *testing.T) {
	cmd := NewCmd("test")
	cmd.SetDescription("Test mixed boolean defaults")

	trueFlag, err := NewBool("true-flag").
		SetUsage("Defaults to true").
		SetDefault(true).
		Register(cmd)
	assert.NoError(t, err)

	falseFlag, err := NewBool("false-flag").
		SetUsage("Defaults to false").
		SetDefault(false).
		Register(cmd)
	assert.NoError(t, err)

	implicitFalse, err := NewBool("implicit").
		SetUsage("Implicitly false").
		Register(cmd)
	assert.NoError(t, err)

	_ = trueFlag
	_ = falseFlag
	_ = implicitFalse

	usage := cmd.GenerateUsage(false)

	// Should show (default true) but not (default false)
	assert.Contains(t, usage, "(default true)")
	assert.NotContains(t, usage, "(default false)")

	// All descriptions should be present
	assert.Contains(t, usage, "Defaults to true")
	assert.Contains(t, usage, "Defaults to false")
	assert.Contains(t, usage, "Implicitly false")
}
