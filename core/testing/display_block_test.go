package testing

import "testing"

func Test_DisplayBlock_CanGiveOwnList(t *testing.T) {
	rsl := `
a = [
	{
		"name": "alice"
	},
	{
		"name": "bob"
	},
]
Name = json[].name
display a:
	fields Name
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Name  
alice  
bob    
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_DisplayBlock_CanGiveOwnMap(t *testing.T) {
	rsl := `
a = {
	"results": [
		{
			"name": "alice"
		},
		{
			"name": "bob"
		},
    ]
}
Name = json.results[].name
display a:
	fields Name
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Name  
alice  
bob    
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
