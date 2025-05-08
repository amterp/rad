package testing

import "testing"

func Test_Pass_Root(t *testing.T) {
	rsl := `
pass
print("Made it!")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Made it!\n")
	assertNoErrors(t)
}

func Test_Pass_IfStmt(t *testing.T) {
	rsl := `
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
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Made it!\n")
	assertNoErrors(t)
}

func Test_Pass_ForLoop(t *testing.T) {
	rsl := `
for i in range(5):
	pass

print("Made it!")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Made it!\n")
	assertNoErrors(t)
}
