package testing

import (
	"testing"
)

func Test_Args_Constraints_Relational_Requires_OkayIfBothRequiredProvided(t *testing.T) {
	rsl := `
args:
    a string
    b string

    a requires b
print("ran")
`
	setupAndRunCode(t, rsl, "--a", "alice", "--b", "bob")
	assertOnlyOutput(t, stdOutBuffer, "ran\n")
	assertNoErrors(t)
}

func Test_Args_Constraints_Relational_Requires_ErrorsIfNotRequired(t *testing.T) {
	rsl := `
args:
    a string
    b string

    a requires b
print("ran")
`
	setupAndRunCode(t, rsl, "--a", "alex")
	expected := `Invalid args: 'a' requires 'b', but 'b' was not set

Usage:
  <a> <b>

Script args:
      --a string   
      --b string   

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Args_Constraints_Relational_Requires_ErrorsIfDefaultRequiresSomethingNotProvided(t *testing.T) {
	rsl := `
args:
    a string
    b string = "bob"
    c string

    b requires c
print("ran")
`
	setupAndRunCode(t, rsl, "--a", "alex")
	expected := `Invalid args: 'b' requires 'c', but 'c' was not set

Usage:
  <a> [b] <c>

Script args:
      --a string   
      --b string    (default bob)
      --c string   

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Args_Constraints_Relational_Excludes_CanGiveFirst(t *testing.T) {
	rsl := `
args:
    file string
    url string

    file mutually excludes url

if is_defined("file"):
    print("Reading from file:", file)
else:
    print("Fetching from URL:", url)
`
	setupAndRunCode(t, rsl, "--file", "file.txt")
	assertOnlyOutput(t, stdOutBuffer, "Reading from file: file.txt\n")
	assertNoErrors(t)
}

func Test_Args_Constraints_Relational_Excludes_CanGiveSecond(t *testing.T) {
	rsl := `
args:
    file string
    url string

    file mutually excludes url

if is_defined("file"):
    print("Reading from file:", file)
else:
    print("Fetching from URL:", url)
`
	setupAndRunCode(t, rsl, "--url", "someurl")
	assertOnlyOutput(t, stdOutBuffer, "Fetching from URL: someurl\n")
	assertNoErrors(t)
}

func Test_Args_Constraints_Relational_Excludes_ErrorsIfBothProvided(t *testing.T) {
	rsl := `
args:
    file string
    url string

    file mutually excludes url

if is_defined("file"):
    print("Reading from file:", file)
else:
    print("Fetching from URL:", url)
`
	setupAndRunCode(t, rsl, "--file", "file.txt", "--url", "someurl")
	expected := `Invalid args: 'file' excludes 'url', but 'url' was set

Usage:
  <file> <url>

Script args:
      --file string   
      --url string    

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Args_Constraints_Relational_Mixed_CanGiveOne(t *testing.T) {
	rsl := `
args:
    token string
    username string
    password string

    username mutually requires password
    token mutually excludes username, password

if is_defined("token"):
    print("Authenticating with token:", token)
else:
    print("Authenticating user:", username)
`
	setupAndRunCode(t, rsl, "--token", "sometoken")
	assertOnlyOutput(t, stdOutBuffer, "Authenticating with token: sometoken\n")
	assertNoErrors(t)
}

func Test_Args_Constraints_Relational_Mixed_CanGiveBothOther(t *testing.T) {
	rsl := `
args:
    token string
    username string
    password string

    username mutually requires password
    token mutually excludes username, password

if is_defined("token"):
    print("Authenticating with token:", token)
else:
    print("Authenticating user:", username)
`
	setupAndRunCode(t, rsl, "--username", "alice", "--password", "pass")
	assertOnlyOutput(t, stdOutBuffer, "Authenticating user: alice\n")
	assertNoErrors(t)
}

func Test_Args_Constraints_Relational_Mixed_ErrorsIfAllGiven(t *testing.T) {
	rsl := `
args:
    token string
    username string
    password string

    username mutually requires password
    token mutually excludes username, password

if is_defined("token"):
    print("Authenticating with token:", token)
else:
    print("Authenticating user:", username)
`
	setupAndRunCode(t, rsl, "--token", "sometoken", "--username", "alice", "--password", "pass")
	expected := `Invalid args: 'token' excludes 'username', but 'username' was set

Usage:
  <token> <username> <password>

Script args:
      --token string      
      --username string   
      --password string   

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Args_Constraints_Relational_ErrorsIfConstraintOnUndefinedArg(t *testing.T) {
	rsl := `
args:
    token string
    token excludes username
`
	setupAndRunCode(t, rsl)
	expected := `Error at L4:20

      token excludes username
                     ^^^^^^^^ Undefined arg 'username'
`
	assertError(t, 1, expected)
}

func Test_Args_Constraints_Relational_Bool_Can_Require(t *testing.T) {
	rsl := `
args:
    authenticate bool
	token string

    authenticate mutually requires token

if authenticate:
    print("Token:", token)
`
	setupAndRunCode(t, rsl, "--authenticate", "--token", "sometoken")
	assertOnlyOutput(t, stdOutBuffer, "Token: sometoken\n")
	assertNoErrors(t)
}

func Test_Args_Constraints_Relational_Bool_ErrorsIfBoolFalse(t *testing.T) {
	rsl := `
args:
    authenticate bool
	token string

    authenticate mutually requires token

if authenticate:
    print("Token:", token)
`
	setupAndRunCode(t, rsl, "--token", "sometoken")
	expected := `Invalid args: 'token' requires 'authenticate', but 'authenticate' was not set

Usage:
  [authenticate] <token>

Script args:
      --authenticate   
      --token string   

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Args_Constraints_Relational_Bool_CanDefineRequireeForNonMutualRequirement(t *testing.T) {
	rsl := `
args:
    authenticate bool
	token string

    authenticate requires token

if authenticate:
    print("Auth Token:", token)
print("Non-auth Token:", token)
`
	setupAndRunCode(t, rsl, "--token", "sometoken")
	assertOnlyOutput(t, stdOutBuffer, "Non-auth Token: sometoken\n")
	assertNoErrors(t)
}
