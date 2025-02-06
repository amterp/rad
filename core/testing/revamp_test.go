package testing

import "testing"

func Test_Revamp_Strings(t *testing.T) {
	rsl := `
a = "hi"
print(a)

b = "hi\n there"
print(b)

c = "hi\n there {1 + 1} \nfriend"
print(c)

d = "hi \\ \h yo"
print(d)

e = r"hi\n there {1 + 1} \nfriend"
print(e)

f = """
 hi
  friend
 """
print(f)

g = "{1 + 1:6.7}"
print(g)

` + "h = `\"hi\" 'there' \\`bob\\``" + "\nprint(h)"
	setupAndRunCode(t, rsl)
	expected := `hi
hi
 there
hi
 there 2 
friend
hi \ h yo
hi\n there {1 + 1} \nfriend
hi
 friend
2
"hi" 'there'` + " `bob`\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
