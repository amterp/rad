package testing

import "testing"

func Test_Tbl_FillsMissingValuesWithEmptyStrings(t *testing.T) {
	rsl := `
names = ["Alice", "Bob", "Charlie"]
ages = [25, 30]
twice = [50, 60, 70]
display:
	fields names, ages, twice
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `names    ages  twice 
Alice    25    50     
Bob      30    60     
Charlie        70     
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Tbl_FillsMissingValuesWithEmptyStringsShortestFirst(t *testing.T) {
	rsl := `
ages = [25, 30]
names = ["Alice", "Bob", "Charlie"]
twice = [50, 60, 70]
display:
	fields ages, names, twice
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `ages  names    twice 
25    Alice    50     
30    Bob      60     
      Charlie  70     
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
