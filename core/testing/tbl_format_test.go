package testing

import "testing"

func TestRad_FormatFloats(t *testing.T) {
	script := `
nums = [0.6342, 0.7, 1.63, 0.0923]
display:
	fields nums
	sort nums
	nums:
		map fn(num) "{num * 100:6.2}%"
`
	setupAndRunCode(t, script, "--color=never")
	expected := `nums    
  9.23%  
 63.42%  
 70.00%  
163.00%  
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRad_MultiplyInts(t *testing.T) {
	script := `
nums = [63, 20, 163, 9]
display:
	fields nums
	sort nums desc
	nums:
		map fn(num) num * 10
`
	setupAndRunCode(t, script, "--color=never")
	expected := `nums 
1630  
630   
200   
90    
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRad_CanTruncateWithMap(t *testing.T) {
	script := `
names = ["Alice", "Bob", "Charlie", "David"]
display:
	fields names
	sort
	names:
		map fn(name) name[:3]
`
	setupAndRunCode(t, script, "--color=never")
	expected := `names 
Ali    
Bob    
Cha    
Dav    
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRad_CanMapTwoFieldsAtOnce(t *testing.T) {
	script := `
FirstNames = ["Alice", "Bob", "Charlie", "David"]
LastNames = ["Smith", "Jones", "Brown", "White"]
display:
	fields FirstNames, LastNames
	sort
	FirstNames, LastNames:
		map fn(name) upper(name)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `FirstNames  LastNames 
ALICE       SMITH      
BOB         JONES      
CHARLIE     BROWN      
DAVID       WHITE      
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
