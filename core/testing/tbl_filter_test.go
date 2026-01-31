package testing

import "testing"

func Test_Rad_Filter_Basic(t *testing.T) {
	script := `
ages = [15, 25, 30, 12]
names = ["Alice", "Bob", "Charlie", "David"]
display:
    fields ages, names
    ages:
        filter fn(a) a >= 18
`
	setupAndRunCode(t, script, "--color=never")
	expected := `ages  names   
25    Bob      
30    Charlie  
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Rad_Filter_MultipleFields_AND(t *testing.T) {
	script := `
ages = [15, 25, 30, 12]
names = ["Alice", "Bob", "Charlie", "Di"]
display:
    fields ages, names
    ages:
        filter fn(a) a >= 18
    names:
        filter fn(n) len(n) > 3
`
	setupAndRunCode(t, script, "--color=never")
	expected := `ages  names   
30    Charlie  
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Rad_Filter_WithMap(t *testing.T) {
	script := `
ages = [15, 25, 30, 12]
display:
    fields ages
    ages:
        filter fn(a) a >= 18
        map fn(a) a * 12
`
	setupAndRunCode(t, script, "--color=never")
	expected := `ages 
300   
360   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Rad_Filter_EmptyResult(t *testing.T) {
	script := `
ages = [15, 12, 10]
display:
    fields ages
    ages:
        filter fn(a) a >= 18
`
	setupAndRunCode(t, script, "--color=never")
	expected := `ages 
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Rad_Filter_UsingIdentifier(t *testing.T) {
	script := `
is_adult = fn(age) age >= 18
ages = [15, 25, 30, 12]
display:
    fields ages
    ages:
        filter is_adult
`
	setupAndRunCode(t, script, "--color=never")
	expected := `ages 
25    
30    
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Rad_Filter_WithSorting(t *testing.T) {
	script := `
ages = [15, 25, 10, 30, 12, 22]
names = ["Alice", "Bob", "Eve", "Charlie", "David", "Frank"]
display:
    fields ages, names
    sort ages desc
    ages:
        filter fn(a) a >= 18
`
	setupAndRunCode(t, script, "--color=never")
	expected := `ages  names   
30    Charlie  
25    Bob      
22    Frank    
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Rad_Filter_FilterBeforeMapVerification(t *testing.T) {
	script := `
values = [10, 20, 30, 40]
display:
    fields values
    values:
        filter fn(v) v < 35
        map fn(v) v * 2
`
	setupAndRunCode(t, script, "--color=never")
	expected := `values 
20      
40      
60      
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_RequestBlock_FilterDoesMutate(t *testing.T) {
	script := `
Ages = json[].age
print("Before request:", Ages)
request "http://example.com":
	fields Ages
	Ages:
		filter fn(a) a >= 18
print("After request:", Ages)
`
	setupAndRunCode(t, script, "--color=never", "--mock-response", "example.com:./resources/mock_ages.json")
	expected := `Before request: [ ]
After request: [ 30, 20 ]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \"example.com\"): http://example.com\n")
	assertHttpInvocationUrls(t, "http://example.com")
	assertNoErrors(t)
}

func Test_Rad_Filter_CannotFilterUnevenLengths(t *testing.T) {
	script := `
values1 = [1, 2, 3, 4]
values2 = [11, 12, 13]
display:
    fields values1, values2
    values1, values2:
        filter fn(v) v % 2 == 0
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20042", "values2", "3 rows but expected 4")
}
