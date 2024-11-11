package testing

import "testing"

const (
	setupSortingRsl = `
url = "https://google.com"

name = json[].name
age = json[].age
city = json[].city
`
)

func TestRadSort_NoSorting(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/people.json", "--NO-COLOR")
	expected := `name     age  city        
Charlie  30   Paris        
Bob      40   London       
Alice    30   New York     
Bob      25   Los Angeles  
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestRadSort_GeneralAscNoToken(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/people.json", "--NO-COLOR")
	expected := `name     age  city        
Alice    30   New York     
Bob      25   Los Angeles  
Bob      40   London       
Charlie  30   Paris        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestRadSort_GeneralAscWithToken(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort asc
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/people.json", "--NO-COLOR")
	expected := `name     age  city        
Alice    30   New York     
Bob      25   Los Angeles  
Bob      40   London       
Charlie  30   Paris        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestRadSort_GeneralDesc(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort desc
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/people.json", "--NO-COLOR")
	expected := `name     age  city        
Charlie  30   Paris        
Bob      40   London       
Bob      25   Los Angeles  
Alice    30   New York     
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestRadSort_ExplicitAsc(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort name asc, age asc, city asc
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/people.json", "--NO-COLOR")
	expected := `name     age  city        
Alice    30   New York     
Bob      25   Los Angeles  
Bob      40   London       
Charlie  30   Paris        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestRadSort_DescTiebreak(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort name asc, age desc, city
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/people.json", "--NO-COLOR")
	expected := `name     age  city        
Alice    30   New York     
Bob      40   London       
Bob      25   Los Angeles  
Charlie  30   Paris        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestRadSort_Mix(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort age, city desc
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/people.json", "--NO-COLOR")
	expected := `name     age  city        
Bob      25   Los Angeles  
Charlie  30   Paris        
Alice    30   New York     
Bob      40   London       
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestRadSort_LeavesInExtractionOrderIfNoTiebreaker(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort age asc
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/people.json", "--NO-COLOR")
	expected := `name     age  city        
Bob      25   Los Angeles  
Charlie  30   Paris        
Alice    30   New York     
Bob      40   London       
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestRadSort_CanUseInIfElseBlocks(t *testing.T) {
	rsl := setupSortingRsl + `
if true:
	rad url:
		fields name, age, city
		sort age asc
else:
	rad url:
		fields name, age, city
		sort age desc
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/people.json", "--NO-COLOR")
	expected := `name     age  city        
Bob      25   Los Angeles  
Charlie  30   Paris        
Alice    30   New York     
Bob      40   London       
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestRadSort_CanSortMixedTypes(t *testing.T) {
	rsl := `
col1 = [1, "a", 2, "b", true, false, { "alice": 1 }, 2, "a", [3, 1, 2], 1.5, -1.2]
col2 = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
display:
	fields col1, col2
	sort
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `col1          col2 
false         5     
true          4     
-1.2          11    
1             0     
1.5           10    
2             2     
2             7     
a             1     
a             8     
b             3     
[3, 1, 2]     9     
{ alice: 1 }  6     
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestRadSort_CanSortMixedTypesDesc(t *testing.T) {
	rsl := `
col1 = [1, "a", 2, "b", true, false, { "alice": 1 }, 2, "a", [3, 1, 2], 1.5, -1.2]
col2 = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
display:
	fields col1, col2
	sort desc
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `col1          col2 
{ alice: 1 }  6     
[3, 1, 2]     9     
b             3     
a             8     
a             1     
2             7     
2             2     
1.5           10    
1             0     
-1.2          11    
true          4     
false         5     
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestRadSort_SortingIsPriorToMapping(t *testing.T) {
	rsl := `
col = [0, 1, 2, 3, 4]
display:
	fields col
	sort asc
	col:
		map num -> -num
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `col 
0    
-1   
-2   
-3   
-4   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestRadSort_ColumnsRemainSortAfter(t *testing.T) {
	rsl := `
col = [3, 4, 2, 1]
display:
	fields col
	sort asc
print(col)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `col 
1    
2    
3    
4    
[1, 2, 3, 4]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestRadSort_DoesNotSortColumnsIfNotAskedTo(t *testing.T) {
	rsl := `
col = [3, 4, 2, 1]
display:
	fields col
print(col)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `col 
3    
4    
2    
1    
[3, 4, 2, 1]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
