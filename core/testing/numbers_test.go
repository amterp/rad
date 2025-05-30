package testing

import "testing"

func Test_NumberPrecision_CanStoreBigInt(t *testing.T) {
	script := `
a = 38123123123123123
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `38123123123123123
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// not desired behavior, just documenting it
func Test_NumberPrecision_BigFloatLosesPrecision(t *testing.T) {
	script := `
a = 38123123123123123.0
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `38123123123123120
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_NumberPrecision_BigFloatInJsonMaintainsIntPrecision(t *testing.T) {
	script := `
a = r'{ "foo": 38123123123123123 }'
print(parse_json(a))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "foo": 38123123123123123 }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_NumberPrecision_BigFloatMaintainsPrecisionWhenPrettyPrinted(t *testing.T) {
	script := `
a = 46046166185414656
pprint(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `46046166185414656
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
