package testing

import "testing"

const (
	setupSortingScript = `
url = "https://google.com"

name = json[].name
age = json[].age
city = json[].city
`
)

func TestRadSort_NoSorting(t *testing.T) {
	script := setupSortingScript + `
rad url:
    fields name, age, city
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json", "--color=never")
	expected := `name     age  city        
Charlie  30   Paris        
Bob      40   London       
Alice    30   New York     
Bob      25   Los Angeles  
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestRadSort_GeneralAscNoToken(t *testing.T) {
	script := setupSortingScript + `
rad url:
    fields name, age, city
    sort
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json", "--color=never")
	expected := `name     age  city        
Alice    30   New York     
Bob      25   Los Angeles  
Bob      40   London       
Charlie  30   Paris        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestRadSort_GeneralAscWithToken(t *testing.T) {
	script := setupSortingScript + `
rad url:
    fields name, age, city
    sort asc
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json", "--color=never")
	expected := `name     age  city        
Alice    30   New York     
Bob      25   Los Angeles  
Bob      40   London       
Charlie  30   Paris        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestRadSort_GeneralDesc(t *testing.T) {
	script := setupSortingScript + `
rad url:
    fields name, age, city
    sort desc
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json", "--color=never")
	expected := `name     age  city        
Charlie  30   Paris        
Bob      40   London       
Bob      25   Los Angeles  
Alice    30   New York     
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestRadSort_ExplicitAsc(t *testing.T) {
	script := setupSortingScript + `
rad url:
    fields name, age, city
    sort name asc, age asc, city asc
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json", "--color=never")
	expected := `name     age  city        
Alice    30   New York     
Bob      25   Los Angeles  
Bob      40   London       
Charlie  30   Paris        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestRadSort_DescTiebreak(t *testing.T) {
	script := setupSortingScript + `
rad url:
    fields name, age, city
    sort name asc, age desc, city
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json", "--color=never")
	expected := `name     age  city        
Alice    30   New York     
Bob      40   London       
Bob      25   Los Angeles  
Charlie  30   Paris        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestRadSort_Mix(t *testing.T) {
	script := setupSortingScript + `
rad url:
    fields name, age, city
    sort age, city desc
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json", "--color=never")
	expected := `name     age  city        
Bob      25   Los Angeles  
Charlie  30   Paris        
Alice    30   New York     
Bob      40   London       
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestRadSort_LeavesInExtractionOrderIfNoTiebreaker(t *testing.T) {
	script := setupSortingScript + `
rad url:
    fields name, age, city
    sort age asc
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json", "--color=never")
	expected := `name     age  city        
Bob      25   Los Angeles  
Charlie  30   Paris        
Alice    30   New York     
Bob      40   London       
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestRadSort_CanUseInIfElseBlocks(t *testing.T) {
	script := setupSortingScript + `
if true:
	rad url:
		fields name, age, city
		sort age asc
else:
	rad url:
		fields name, age, city
		sort age desc
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json", "--color=never")
	expected := `name     age  city        
Bob      25   Los Angeles  
Charlie  30   Paris        
Alice    30   New York     
Bob      40   London       
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestRadSort_CanSortMixedTypes(t *testing.T) {
	script := `
col1 = [1, "a", 2, "b", true, false, { "alice": 1 }, 2, "a", [3, 1, 2], 1.5, -1.2]
col2 = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
display:
	fields col1, col2
	sort
`
	setupAndRunCode(t, script, "--color=never")
	expected := `col1            col2 
false           5     
true            4     
-1.2            11    
1               0     
1.5             10    
2               2     
2               7     
a               1     
a               8     
b               3     
[ 3, 1, 2 ]     9     
{ "alice": 1 }  6     
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRadSort_CanSortMixedTypesDesc(t *testing.T) {
	script := `
col1 = [1, "a", 2, "b", true, false, { "alice": 1 }, 2, "a", [3, 1, 2], 1.5, -1.2]
col2 = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
display:
	fields col1, col2
	sort desc
`
	setupAndRunCode(t, script, "--color=never")
	expected := `col1            col2 
{ "alice": 1 }  6     
[ 3, 1, 2 ]     9     
b               3     
a               8     
a               1     
2               7     
2               2     
1.5             10    
1               0     
-1.2            11    
true            4     
false           5     
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRadSort_SortingIsPriorToMapping(t *testing.T) {
	script := `
col = [0, 1, 2, 3, 4]
display:
	fields col
	sort asc
	col:
		map fn(num) -num
`
	setupAndRunCode(t, script, "--color=never")
	expected := `col 
0    
-1   
-2   
-3   
-4   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRadSort_ColumnsRemainSortAfter(t *testing.T) {
	script := `
col = [3, 4, 2, 1]
display:
	fields col
	sort asc
print(col)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `col 
1    
2    
3    
4    
[ 1, 2, 3, 4 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRadSort_DoesNotSortColumnsIfNotAskedTo(t *testing.T) {
	script := `
col = [3, 4, 2, 1]
display:
	fields col
print(col)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `col 
3    
4    
2    
1    
[ 3, 4, 2, 1 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
