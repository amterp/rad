package testing

import "testing"

// todo sorting should occur before mapping, should also be done on original types

func TestRad_FormatFloats(t *testing.T) {
	rsl := `
nums = [0.6342, 0.7, 1.63, 0.0923]
display:
	fields nums
	sort nums
	nums:
		map num -> "{num * 100:6.2}%"
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `nums    
  9.23%  
 63.42%  
 70.00%  
163.00%  
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestRad_MultiplyInts(t *testing.T) {
	rsl := `
nums = [63, 20, 163, 9]
display:
	fields nums
	sort nums desc
	nums:
		map num -> num * 10
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `nums 
90    
630   
200   
1630  
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestRad_CanTruncateWithMap(t *testing.T) {
	rsl := `
names = ["Alice", "Bob", "Charlie", "David"]
display:
	fields names
	sort
	names:
		map name -> name[:3]
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `names 
Ali    
Bob    
Cha    
Dav    
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestRad_CanMapTwoFieldsAtOnce(t *testing.T) {
	rsl := `
FirstNames = ["Alice", "Bob", "Charlie", "David"]
LastNames = ["Smith", "Jones", "Brown", "White"]
display:
	fields FirstNames, LastNames
	sort
	FirstNames, LastNames:
		map name -> upper(name)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `FirstNames  LastNames 
ALICE       SMITH      
BOB         JONES      
CHARLIE     BROWN      
DAVID       WHITE      
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
