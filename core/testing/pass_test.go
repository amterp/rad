package testing

import "testing"

func Test_Pass_Root(t *testing.T) {
	script := `
pass
print("Made it!")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Made it!\n")
	assertNoErrors(t)
}

func Test_Pass_IfStmt(t *testing.T) {
	script := `
if true:
	pass
else:
	pass

if false:
	pass
else:
	pass

print("Made it!")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Made it!\n")
	assertNoErrors(t)
}

func Test_Pass_ForLoop(t *testing.T) {
	script := `
for i in range(5):
	pass

print("Made it!")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Made it!\n")
	assertNoErrors(t)
}
