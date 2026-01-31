package testing

import "testing"

func Test_Func_EncodeDecodeBase64(t *testing.T) {
	script := `
t = encode_base64("hello friend!")
print(t)
print(decode_base64(t))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `aGVsbG8gZnJpZW5kIQ==
hello friend!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_EncodeDecodeBase64NoPadding(t *testing.T) {
	script := `
t = encode_base64("hello friend!", padding=false)
print(t)
print(decode_base64(t, padding=false))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `aGVsbG8gZnJpZW5kIQ
hello friend!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_EncodeDecodeBase64UrlSafe(t *testing.T) {
	script := `
t = encode_base64("is this url friendly?!", url_safe=true)
print(t)
print(decode_base64(t, url_safe=true))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `aXMgdGhpcyB1cmwgZnJpZW5kbHk_IQ==
is this url friendly?!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_DecodeBase64ErrorsOnInvalidInput(t *testing.T) {
	script := `
decode_base64(";;;< those are not part of base64")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20021", "Error decoding base64: illegal base64 data at input byte 0")
}

func Test_Func_EncodeDecodeBase16(t *testing.T) {
	script := `
t = encode_base16("hello friend!")
print(t)
print(decode_base16(t))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `68656c6c6f20667269656e6421
hello friend!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
