package com

import (
	"github.com/sanity-io/litter"
)

func FlatStr(item any) string {
	return litter.Sdump(item)
}

func init() {
	litter.Config.Compact = true
	litter.Config.StripPackageNames = true
	litter.Config.DisablePointerReplacement = true
}
