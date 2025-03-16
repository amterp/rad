package testing

import "testing"

func Test_If_PreStmt_Works(t *testing.T) {
	rsl := `
a = -10
if a += 3; a > 4:
    print("1")
else if a += 5; a > 6:
    print("2")
else:
    print("3")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
}

func Test_If_PreStmt_BasicAssignment(t *testing.T) {
	rsl := `
if val = 5; val > 0:
    print("positive")
else:
    print("non-positive")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "positive\n")
	assertNoErrors(t)
}

func Test_If_PreStmt_ChainConditions(t *testing.T) {
	rsl := `
x = 2
if x *= 3; x > 10:
    print("big")
else if x += 5; x > 5:
    print("medium")
else:
    print("small")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "medium\n")
	assertNoErrors(t)
}

func Test_If_PreStmt_Nested(t *testing.T) {
	rsl := `
y = 4
if y -= 2; y > 0:
    if y *= 3; y > 5:
        print("yes")
    else:
        print("no")
else:
    print("negative")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "yes\n")
	assertNoErrors(t)
}

func Test_If_PreStmt_ExecutesAll(t *testing.T) {
	rsl := `
if print(10); false:
	a = 2
else if print(20); false:
	a = 2
else:
    print("done")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "10\n20\ndone\n")
	assertNoErrors(t)
}
