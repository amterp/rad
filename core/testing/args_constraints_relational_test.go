package testing

import (
	"testing"
)

func Test_Args_Constraints_Relational_Requires_OkayIfBothRequiredProvided(t *testing.T) {
	script := `
args:
    a str
    b str

    a requires b
print("ran")
`
	setupAndRunCode(t, script, "--a", "alice", "--b", "bob")
	assertOnlyOutput(t, stdOutBuffer, "ran\n")
	assertNoErrors(t)
}

func Test_Args_Constraints_Relational_Requires_ErrorsIfNotRequired(t *testing.T) {
	script := `
args:
    a str
    b str

    a requires b
print("ran")
`
	setupAndRunCode(t, script, "--a", "alex", "--color=never")
	expected := `Invalid args: 'a' requires 'b', but 'b' was not set

Usage:
  TestCase <a> <b> [OPTIONS]

Script args:
      --a str   Requires: b
      --b str

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Args_Constraints_Relational_Requires_ErrorsIfDefaultRequiresSomethingNotProvided(t *testing.T) {
	script := `
args:
    a str
    b str = "bob"
    c str

    b requires c
print("ran")
`
	setupAndRunCode(t, script, "--a", "alex", "--color=never")
	expected := `Invalid args: 'b' requires 'c', but 'c' was not set

Usage:
  TestCase <a> [b] <c> [OPTIONS]

Script args:
      --a str
      --b str   Requires: c (default bob)
      --c str

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Args_Constraints_Relational_Excludes_CanGiveFirst(t *testing.T) {
	script := `
args:
    file str
    url str

    file mutually excludes url

if is_defined("file"):
    print("Reading from file:", file)
else:
    print("Fetching from URL:", url)
`
	setupAndRunCode(t, script, "--file", "file.txt")
	assertOnlyOutput(t, stdOutBuffer, "Reading from file: file.txt\n")
	assertNoErrors(t)
}

func Test_Args_Constraints_Relational_Excludes_CanGiveSecond(t *testing.T) {
	script := `
args:
    file str
    url str

    file mutually excludes url

if is_defined("file"):
    print("Reading from file:", file)
else:
    print("Fetching from URL:", url)
`
	setupAndRunCode(t, script, "--url", "someurl")
	assertOnlyOutput(t, stdOutBuffer, "Fetching from URL: someurl\n")
	assertNoErrors(t)
}

func Test_Args_Constraints_Relational_Excludes_ErrorsIfBothProvided(t *testing.T) {
	script := `
args:
    file str
    url str

    file mutually excludes url

if is_defined("file"):
    print("Reading from file:", file)
else:
    print("Fetching from URL:", url)
`
	setupAndRunCode(t, script, "--file", "file.txt", "--url", "someurl", "--color=never")
	expected := `Invalid args: 'file' excludes 'url', but 'url' was set

Usage:
  TestCase <file> <url> [OPTIONS]

Script args:
      --file str   Excludes: url
      --url str    Excludes: file

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Args_Constraints_Relational_Mixed_CanGiveOne(t *testing.T) {
	script := `
args:
    token str
    username str
    password str

    username mutually requires password
    token mutually excludes username, password

if is_defined("token"):
    print("Authenticating with token:", token)
else:
    print("Authenticating user:", username)
`
	setupAndRunCode(t, script, "--token", "sometoken")
	assertOnlyOutput(t, stdOutBuffer, "Authenticating with token: sometoken\n")
	assertNoErrors(t)
}

func Test_Args_Constraints_Relational_Mixed_CanGiveBothOther(t *testing.T) {
	script := `
args:
    token str
    username str
    password str

    username mutually requires password
    token mutually excludes username, password

if is_defined("token"):
    print("Authenticating with token:", token)
else:
    print("Authenticating user:", username)
`
	setupAndRunCode(t, script, "--username", "alice", "--password", "pass")
	assertOnlyOutput(t, stdOutBuffer, "Authenticating user: alice\n")
	assertNoErrors(t)
}

func Test_Args_Constraints_Relational_Mixed_ErrorsIfAllGiven(t *testing.T) {
	script := `
args:
    token str
    username str
    password str

    username mutually requires password
    token mutually excludes username, password

if is_defined("token"):
    print("Authenticating with token:", token)
else:
    print("Authenticating user:", username)
`
	setupAndRunCode(t, script, "--token", "sometoken", "--username", "alice", "--password", "pass", "--color=never")
	expected := `Invalid args: 'token' excludes 'username', but 'username' was set

Usage:
  TestCase <token> <username> <password> [OPTIONS]

Script args:
      --token str      Excludes: username, password
      --username str   Requires: password. Excludes: token
      --password str   Requires: username. Excludes: token

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Args_Constraints_Relational_ErrorsIfConstraintOnUndefinedArg(t *testing.T) {
	script := `
args:
    token str
    token excludes username
`
	setupAndRunCode(t, script, "--color=never")
	expected := "Undefined flag 'username'\n"
	assertError(t, 1, expected)
}

func Test_Args_Constraints_Relational_Bool_Can_Require(t *testing.T) {
	script := `
args:
    authenticate bool
	token str

    authenticate mutually requires token

if authenticate:
    print("Token:", token)
`
	setupAndRunCode(t, script, "--authenticate", "--token", "sometoken")
	assertOnlyOutput(t, stdOutBuffer, "Token: sometoken\n")
	assertNoErrors(t)
}

func Test_Args_Constraints_Relational_Bool_ErrorsIfBoolFalse(t *testing.T) {
	script := `
args:
    authenticate bool
	token str

    authenticate mutually requires token

if authenticate:
    print("Token:", token)
`
	setupAndRunCode(t, script, "--token", "sometoken", "--color=never")
	expected := `Invalid args: 'token' requires 'authenticate', but 'authenticate' was not set

Usage:
  TestCase <token> [OPTIONS]

Script args:
      --token str      Requires: authenticate
      --authenticate   Requires: token

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Args_Constraints_Relational_Bool_CanDefineRequireeForNonMutualRequirement(t *testing.T) {
	script := `
args:
    authenticate bool
	token str

    authenticate requires token

if authenticate:
    print("Auth Token:", token)
print("Non-auth Token:", token)
`
	setupAndRunCode(t, script, "--token", "sometoken")
	assertOnlyOutput(t, stdOutBuffer, "Non-auth Token: sometoken\n")
	assertNoErrors(t)
}

func Test_Args_Constraints_Relational_Bool_OnlyRelevantIfTrue(t *testing.T) {
	script := `
args:
	mystring str
	mybool bool

	mybool excludes mystring

print(mystring)
`
	setupAndRunCode(t, script, "--mystring", "blah")
	assertOnlyOutput(t, stdOutBuffer, "blah\n")
	assertNoErrors(t)
}
