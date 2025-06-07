package core

import (
	com "rad/core/common"
	"strings"

	"github.com/amterp/rad/rts/rl"

	"github.com/amterp/color"

	ts "github.com/tree-sitter/go-tree-sitter"
)

var EMPTY_STR = NewRadString("")

type RadString struct {
	Segments []radStringSegment
	// todo offer isRegex funcs which will return true if all segments are regexes, for example?
	//  concat between non-regex and regex give non-regex? or regex?
}

type radStringSegment struct {
	String     string
	Attributes []RadTextAttr
	Hyperlink  *string
	Rgb        *com.Rgb
}

// todo should these methods be returning *RadString?
func NewRadString(str string) RadString {
	if str == "" {
		return RadString{Segments: []radStringSegment{}}
	}
	return RadString{Segments: []radStringSegment{{String: str, Attributes: make([]RadTextAttr, 0)}}}
}

func newRadStringWithAttr(str string, segment radStringSegment) RadString {
	// todo the fact that this copies without me explicitly telling it probably means we're being wasteful
	segment.String = str
	return RadString{Segments: []radStringSegment{segment}}
}

// Copies only the attributes of the first segment. Maybe could change somehow?
func (s RadString) CopyAttrTo(otherStr string) RadString {
	if len(s.Segments) == 0 {
		return NewRadString(otherStr)
	}

	cpy := s.DeepCopy()
	cpy.Segments[0].String = otherStr
	cpy.Segments = cpy.Segments[:1] // keep only the first segment
	return cpy
}

// does not apply any attributes
func (s RadString) Plain() string {
	// todo can lazily compute and cache
	var result string
	for _, segment := range s.Segments {
		result += segment.String
	}
	return result
}

// applies all the attributes
func (s RadString) String() string {
	builder := strings.Builder{}
	for _, segment := range s.Segments {
		builder.WriteString(s.applyAttributes(segment.String, segment))
	}
	return builder.String()
}

func (s *RadString) ToRuneList() *RadList {
	result := NewRadList()
	for i := int64(0); i < s.Len(); i++ {
		result.Append(newRadValueRadStr(s.IndexAt(i)))
	}
	return result
}

func (s RadString) Concat(other RadString) RadString {
	return RadString{Segments: append(s.Segments, other.Segments...)}
}

func (s RadString) ConcatStr(other string) RadString {
	return s.Concat(NewRadString(other))
}

func (s RadString) Equals(other RadString) bool {
	return s.Plain() == other.Plain()
}

func (s RadString) Len() int64 {
	// todo also cachable
	return int64(com.StrLen(s.Plain()))
}

func (s *RadString) Index(i *Interpreter, idxNode *ts.Node) RadString {
	if idxNode.Kind() == rl.K_SLICE {
		// todo should maintain attr info
		start, end := ResolveSliceStartEnd(i, idxNode, s.Len())
		return NewRadString(s.Plain()[start:end])
	}

	idxVal := i.evaluate(idxNode, 1)[0]
	rawIdx := idxVal.RequireInt(i, idxNode)
	idx := CalculateCorrectedIndex(rawIdx, s.Len(), false)
	if idx < 0 || idx >= s.Len() {
		ErrIndexOutOfBounds(i, idxNode, rawIdx, s.Len())
	}

	return s.IndexAt(idx)
}

// assumes idx is valid for this string
func (s *RadString) IndexAt(idx int64) RadString {
	cumLen := 0
	for _, segment := range s.Segments {
		nextSegmentLen := len(segment.String)
		if cumLen+nextSegmentLen > int(idx) {
			// rune array conversion required to handle multibyte characters e.g. emojis
			char := []rune(s.Plain())[idx] // todo inefficient, should just look up in segment
			return newRadStringWithAttr(string(char), segment)
		}
		cumLen += +nextSegmentLen
	}
	RP.RadErrorExit("Bug! IndexAt called with invalid index")
	panic(UNREACHABLE)
}

func (s *RadString) Compare(other RadString) int {
	sVal := s.Plain()
	oVal := other.Plain()
	if sVal < oVal {
		return -1
	}
	if sVal > oVal {
		return 1
	}
	return 0
}

func (s *RadString) DeepCopy() RadString {
	cpy := *s
	cpy.Segments = make([]radStringSegment, len(s.Segments))
	copy(cpy.Segments, s.Segments)
	return cpy
}

func (s *RadString) CopyWithAttr(attr RadTextAttr) RadString {
	cpy := s.DeepCopy()
	for i := range cpy.Segments {
		cpy.Segments[i].Attributes = append(cpy.Segments[i].Attributes, attr)
	}
	return cpy
}

func (s *RadString) Hyperlink(link RadString) RadString {
	cpy := s.DeepCopy()
	for i := range cpy.Segments {
		str := link.Plain()
		cpy.Segments[i].Hyperlink = &str
	}
	return cpy
}

func (s RadString) Upper() RadString {
	cpy := s.DeepCopy()
	for i, segment := range cpy.Segments {
		cpy.Segments[i].String = strings.ToUpper(segment.String)
	}
	return cpy
}

func (s RadString) Lower() RadString {
	cpy := s.DeepCopy()
	for i, segment := range cpy.Segments {
		cpy.Segments[i].String = strings.ToLower(segment.String)
	}
	return cpy
}

func (s *RadString) SetAttr(attr RadTextAttr) {
	for i := range s.Segments {
		s.Segments[i].Attributes = append(s.Segments[i].Attributes, attr)
	}
}

func (s *RadString) SetSegmentsHyperlink(link RadString) {
	for i := range s.Segments {
		str := link.Plain()
		s.Segments[i].Hyperlink = &str
	}
}

func (s *RadString) Trim(chars string) RadString {
	// todo should maintain attr info
	return NewRadString(strings.Trim(s.Plain(), chars))
}

func (s *RadString) TrimPrefix(prefix string) RadString {
	// todo should maintain attr info
	return NewRadString(strings.TrimLeft(s.Plain(), prefix))
}

func (s *RadString) TrimSuffix(suffix string) RadString {
	// todo should maintain attr info
	return NewRadString(strings.TrimRight(s.Plain(), suffix))
}

func (s *RadString) Reverse() RadString {
	// todo should maintain attr info
	return NewRadString(com.Reverse(s.Plain()))
}

func (s RadString) SetRgb(red int, green int, blue int) {
	rgb := com.NewRgb(red, green, blue)
	for i := range s.Segments {
		s.Segments[i].Rgb = &rgb
	}
}

func (s RadString) SetRgb64(red int64, green int64, blue int64) {
	s.SetRgb(int(red), int(green), int(blue))
}

func (s RadString) Repeat(multiplier int64) RadString {
	if multiplier <= 0 {
		return NewRadString("")
	}

	cpy := s.DeepCopy()
	for i := int64(1); i < multiplier; i++ {
		cpy.Segments = append(cpy.Segments, s.Segments...)
	}
	return cpy
}

func (s *RadString) applyAttributes(str string, segment radStringSegment) string {
	if len(segment.Attributes) == 0 && segment.Hyperlink == nil && segment.Rgb == nil {
		return str
	}

	clr := color.New()

	for _, attr := range segment.Attributes {
		attr.AddAttrTo(clr)
	}

	if segment.Rgb != nil {
		// todo note: ordering here means RGB is applied after the other attributes.
		//  What if yellow() invoked after RGB on a string?
		clr = clr.AddRGB(segment.Rgb.R, segment.Rgb.G, segment.Rgb.B)
	}

	if segment.Hyperlink != nil {
		clr = clr.Hyperlink(*segment.Hyperlink)
	}

	return clr.Sprint(str)
}
