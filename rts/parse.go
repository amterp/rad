package rts

import (
	"strconv"
	"strings"
)

func ParseInt(src string) (int64, error) {
	toParse := strings.ReplaceAll(src, "_", "")
	return strconv.ParseInt(toParse, 10, 64)
}

func ParseFloat(src string) (float64, error) {
	toParse := strings.ReplaceAll(src, "_", "")
	return strconv.ParseFloat(toParse, 64)
}
