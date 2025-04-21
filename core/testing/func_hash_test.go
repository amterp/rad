package testing

import "testing"

func Test_Func_Hash(t *testing.T) {
	rsl := `
hash("hello friend!").print()
hash("hello friend!", algo="sha1").print()
hash("hello friend!", algo="sha256").print()
hash("hello friend!", algo="sha512").print()
hash("hello friend!", algo="md5").print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `925ad7fb5133dfdbed43b9c5a472b09fb53d54c2
925ad7fb5133dfdbed43b9c5a472b09fb53d54c2
0f3319990f3d61147e6fe024c13eb54a986cf707a2bd76fe7930611e52303ff2
c36a9dc2ed5a18ffb7c529a6df6cfd214b34ff618441f1254dd34830ad6c4d271b42fdf6e1bcc54f1858067d18841e4bd815ea6653fa8f49ff171e9969ca84d2
9e3e03e6abdf9ef54942eb4e58319b4a
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_HashErrorsForUnknownAlgo(t *testing.T) {
	rsl := `
hash("hello friend!", algo="does not exist")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:28

  hash("hello friend!", algo="does not exist")
                             ^^^^^^^^^^^^^^^^
                             Unsupported hash algorithm "does not exist"; supported: sha1, sha256, sha512, md5
`
	assertError(t, 1, expected)
}
