package testing

import "testing"

func Test_Assign_InsideCollection(t *testing.T) {
	rsl := `a = [1, 2]
a[0] = 3
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 3, 2 ]\n")
	assertNoErrors(t)
}

func Test_MultiAssign_InsideCollectionViaSwitch(t *testing.T) {
	rsl := `a = [1, [2, 3], 4]
b = "alice"
a[1][0], a[2] = switch b:
	case "alice": 20, 30
	case "bob": 40, 50
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, [ 20, 3 ], 30 ]\n")
	assertNoErrors(t)
}

func Test_MultiAssign_InsideCollectionViaFunc(t *testing.T) {
	rsl := `a = [1, [2, 3], 4]
a[1][0], a[2] = pick_from_resource("./resources/people.json", "alice")
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, [ \"Alice\", 3 ], 25 ]\n")
	assertNoErrors(t)
}

func Test_Assign_MultiSimple(t *testing.T) {
	rsl := `a, b = 1, 2
print(a, b)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1 2\n")
	assertNoErrors(t)
}

func Test_Assign_CannotAssignJsonFieldToIndexVarPath(t *testing.T) {
	rsl := `a = [0]
a[0] = json.id
`

	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  a[0] = json.id
  ^^^^ Json paths must be defined to plain identifiers
`
	assertError(t, 1, expected)
}

func Test_Assign_CannotAssignJsonFieldToIndexVarPathInMulti(t *testing.T) {
	rsl := `b = []
a, b[0], c = 1, json.id, 3
`

	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:4

  a, b[0], c = 1, json.id, 3
     ^^^^ Json paths must be defined to plain identifiers
`
	assertError(t, 1, expected)
}

func Test_Assign_CanAssignMultiReturningFunctionsIfSufficientLefts(t *testing.T) {
	rsl := `a, b, c, d = parse_int("1"), parse_int("2")
print(a, b, c, d)
`

	setupAndRunCode(t, rsl, "--color=never")
	expected := `1 { } 2 { }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Assign_CanAssignMultiIntoCollections(t *testing.T) {
	rsl := `a, b = [10], [20]
a[0], b[0] = 100, 200
print(a, b)
`

	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ 100 ] [ 200 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
