package testing

import "testing"

func TestAlgo_Index_CanExtractIndexLeafNode(t *testing.T) {
	script := `
url = "https://google.com"
Ids = json[].ids[1]
Names = json[].name
request url:
    fields Ids, Names
print([x * 10 for x in Ids])
print(Names)
`

	setupAndRunCode(t, script, "--mock-response", ".*:./responses/arrays.json", "--color=never")
	expected := `[ 20, 50, 100 ]
[ "Alice", "Bob", "Charlie" ]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestAlgo_Index_CanExtractViaMiddleNodeIndexing(t *testing.T) {
	script := `
url = "https://google.com"
Name = json[1].name
request url:
    fields Name
Name = Name[0]
print(Name, len(Name))
`

	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	expected := `Bob 3
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

// todo this should work, but doesn't. in part because of the 'num captures' logic doesn't seem to hit through, and also
//  the 'Name' field doesn't *know* it'll become an array, so it gets written as a single value.
//func TestAlgo_Index_RepeatedCapturesFromLevels(t *testing.T) {
//	rl := `
//url = "https://google.com"
//Name = json[1].name
//Friend = json[1].friends[].name
//request url:
//  fields Name, Friend
//print(Name)
//print(Friend)
//`
//
//	setupAndRunCode(t, rl, "--mock-response", ".*:./responses/obj_arr_with_arrays.json", "--color=never")
//	expected := `[Bob, Bob]
//[Alice, Charlie]
//`
//	assertOutput(t, stdOutBuffer, expected)
//	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
//	assertNoErrors(t)
//	//}
