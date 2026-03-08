package core

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Matches day components like "1d", "0.5d", ".5d", "3.14d"
var dayPattern = regexp.MustCompile(`(\d*\.?\d+)d`)

// ParseDurationString parses a human-readable duration string into nanoseconds.
// Extends Go's time.ParseDuration with support for "d" (days, where 1d = 24h).
// Spaces are stripped, and a leading "-" negates the whole duration.
func ParseDurationString(s string) (int64, error) {
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	}

	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	var dayNanos float64
	dayMatches := 0
	remainder := dayPattern.ReplaceAllStringFunc(s, func(match string) string {
		m := dayPattern.FindStringSubmatch(match)
		days, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return match // leave it for time.ParseDuration to error on
		}
		dayNanos += days * 24 * float64(time.Hour)
		dayMatches++
		return ""
	})

	if dayMatches > 1 {
		return 0, fmt.Errorf("multiple day components in duration string")
	}

	dayInt := int64(dayNanos)

	if dayNanos != 0 {
		if math.Abs(dayNanos) > float64(math.MaxInt64) {
			return 0, fmt.Errorf("duration overflow: days component too large")
		}
	}

	var remainderNanos int64
	if remainder != "" {
		dur, err := time.ParseDuration(remainder)
		if err != nil {
			return 0, fmt.Errorf("invalid duration: %w", err)
		}
		remainderNanos = dur.Nanoseconds()
	}

	// Overflow check: both positive and sum wrapped negative, or both
	// negative and sum wrapped positive.
	if dayInt > 0 && remainderNanos > 0 && remainderNanos > math.MaxInt64-dayInt {
		return 0, fmt.Errorf("duration overflow: total exceeds ~292 years")
	}
	if dayInt < 0 && remainderNanos < 0 && remainderNanos < math.MinInt64-dayInt {
		return 0, fmt.Errorf("duration overflow: total exceeds ~292 years")
	}

	totalNanos := dayInt + remainderNanos

	if negative {
		totalNanos = -totalNanos
	}

	return totalNanos, nil
}

// NewDurationMap builds a RadMap from a nanosecond value, providing
// conversions to all common time units.
func NewDurationMap(nanos int64) *RadMap {
	m := NewRadMap()
	m.SetPrimitiveInt64("nanos", nanos)
	m.SetPrimitiveFloat("micros", float64(nanos)/float64(time.Microsecond))
	m.SetPrimitiveFloat("millis", float64(nanos)/float64(time.Millisecond))
	m.SetPrimitiveFloat("seconds", float64(nanos)/float64(time.Second))
	m.SetPrimitiveFloat("minutes", float64(nanos)/float64(time.Minute))
	m.SetPrimitiveFloat("hours", float64(nanos)/float64(time.Hour))
	m.SetPrimitiveFloat("days", float64(nanos)/(24*float64(time.Hour)))
	return m
}
