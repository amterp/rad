package ra

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_Basic(t *testing.T) {
	fs := NewFlagSet()

	boolFlag := fs.AddBool("foo").
		SetShort("f").
		SetUsage("foo usage here").
		SetDefault(true).
		Value
	strFlag := fs.AddString("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetDefault("alice").
		Value

	err := fs.Parse(os.Args)
	assert.Nil(t, err)

	assert.Equal(t, true, boolFlag)
	assert.Equal(t, "alice", strFlag)
}
