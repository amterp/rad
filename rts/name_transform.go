package rts

import "strings"

// ToExternalName converts internal argument names to external CLI flag names.
// This is the single source of truth for name transformations in Rad.
func ToExternalName(internalName string) string {
	return strings.Replace(internalName, "_", "-", -1)
}
