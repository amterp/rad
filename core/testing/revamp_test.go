package testing

import "testing"

func Test_Revamp_Strings(t *testing.T) {
	script := `
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
	setupAndRunCode(t, script, "--color=never")
	expected := `hi
hi
 there
hi
 there 2 
friend
hi \ \h yo
hi\n there {1 + 1} \nfriend
hi
 friend
2.0000000
"hi" 'there'` + " `bob`\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Revamp_ForLoop(t *testing.T) {
	script := `
a = [20, 30, 40, 50, 60]

for n in a:
    print(n)
    if n > 35:
        break
    print("after")

for n in a:
    print(n)
    if n > 35:
        continue
    print("after")

n = "alice"

for l in n with loop:
    print(loop.idx, l)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `20
after
30
after
40
20
after
30
after
40
50
60
0 a
1 l
2 i
3 c
4 e
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
