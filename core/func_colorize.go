package core

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/samber/lo"
)

var FuncColorize = BuiltInFunc{
	Name: FUNC_COLORIZE,
	Execute: func(f FuncInvocationArgs) RadValue {
		valueArg := f.args[0]
		possibleValuesArg := f.args[1]
		possibleValues := possibleValuesArg.value.RequireList(f.i, possibleValuesArg.node).AsStringList(false)

		if len(possibleValues) == 0 {
			f.i.errorf(possibleValuesArg.node, "Possible values list cannot be empty, but was.")
		}

		possibleValues = lo.Uniq(possibleValues)
		sort.Strings(possibleValues)

		value := valueArg.value.RequireStr(f.i, valueArg.node).Plain()
		if !lo.Contains(possibleValues, value) {
			f.i.errorf(
				valueArg.node,
				"Value '%s' not found in the provided list of possible values: %s",
				value,
				possibleValues,
			)
		}

		r, g, b, err := GetEnumColor(value, possibleValues)
		if err != nil {
			f.i.errorf(f.callNode, "Failed to get color for value '%s': %s", value, err.Error())
		}

		s := NewRadString(value)
		s.SetRgb(r, g, b)
		return newRadValues(f.i, f.callNode, s)
	},
}

const goldenRatioConjugate = 0.61803398875

// slPair stores a Saturation and Lightness pair.
type slPair struct {
	s float64 // Saturation (0.0 to 1.0)
	l float64 // Lightness (0.0 to 1.0)
}

// predefinedSLPairs to cycle through for varied appeal.
// These are chosen to be generally appealing: not too dark/light, reasonably saturated.
var predefinedSLPairs = []slPair{
	{s: 0.80, l: 0.50}, // Standard, good visibility and vibrancy
	{s: 0.75, l: 0.65}, // Lighter, slightly less saturated
	{s: 0.85, l: 0.40}, // Darker, more saturated (good contrast potential)
	{s: 0.60, l: 0.55}, // Mid-light, more pastel-like
	{s: 0.90, l: 0.70}, // Bright and very saturated
}

// GetEnumColor generates a visually distinct and appealing RGB color for a given value from a list.
// It returns R, G, B values (0-255) and an error if the value is not found.
func GetEnumColor(value string, possibleSortedValues []string) (r, g, b int, err error) {
	if len(possibleSortedValues) == 0 {
		return 0, 0, 0, errors.New("values list cannot be empty")
	}

	idx := -1
	for i, v := range possibleSortedValues {
		if v == value {
			idx = i
			break
		}
	}

	if idx == -1 {
		return 0, 0, 0, fmt.Errorf("value '%s' not found in the provided list of values", value)
	}

	// 1. Calculate Hue (H)
	// Adding a small initial offset to vary the starting color slightly.
	hueInitialOffset := 0.01
	hue := math.Mod(hueInitialOffset+float64(idx)*goldenRatioConjugate, 1.0)
	if hue < 0 { // Ensure hue is in [0, 1)
		hue += 1.0
	}

	// 2. Select Lightness (L) and Saturation (S) from predefined pairs
	pair := predefinedSLPairs[idx%len(predefinedSLPairs)]
	lightness := pair.l
	saturation := pair.s

	// 3. Convert HSL to RGB
	rFloat, gFloat, bFloat := hslToRgb(hue, saturation, lightness)

	// 4. Convert RGB from 0-1 range to 0-255
	// Adding 0.5 before casting to int effectively rounds the number.
	r = int(rFloat*255.0 + 0.5)
	g = int(gFloat*255.0 + 0.5)
	b = int(bFloat*255.0 + 0.5)

	// Clamp values to ensure they are within [0, 255] after rounding, though unlikely needed here.
	r = int(math.Max(0, math.Min(255, float64(r))))
	g = int(math.Max(0, math.Min(255, float64(g))))
	b = int(math.Max(0, math.Min(255, float64(b))))

	return r, g, b, nil
}

// hslToRGB converts HSL color values to RGB.
// H, S, L are all in the range [0.0, 1.0].
// Returns R, G, B values also in the range [0.0, 1.0].
func hslToRgb(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		// Achromatic (gray)
		return l, l, l
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r = hueToRgbComponent(p, q, h+1.0/3.0)
	g = hueToRgbComponent(p, q, h)
	b = hueToRgbComponent(p, q, h-1.0/3.0)

	return
}

func hueToRgbComponent(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}
