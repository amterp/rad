package core

import (
	com "rad/core/common"
	"strings"

	"github.com/amterp/rts/rsl"

	"github.com/amterp/color"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type RslString struct {
	Segments []rslStringSegment
	// todo offer isRegex funcs which will return true if all segments are regexes, for example?
	//  concat between non-regex and regex give non-regex? or regex?
}

type rslStringSegment struct {
	String     string
	Attributes []RslTextAttr
	Hyperlink  *string
	Rgb        *com.Rgb
}

// todo should these methods be returning *RslString?
func NewRslString(str string) RslString {
	if str == "" {
		return RslString{Segments: []rslStringSegment{}}
	}
	return RslString{Segments: []rslStringSegment{{String: str, Attributes: make([]RslTextAttr, 0)}}}
}

func newRslStringWithAttr(str string, segment rslStringSegment) RslString {
	// todo the fact that this copies without me explicitly telling it probably means we're being wasteful
	segment.String = str
	return RslString{Segments: []rslStringSegment{segment}}
}

// does not apply any attributes
func (s RslString) Plain() string {
	// todo can lazily compute and cache
	var result string
	for _, segment := range s.Segments {
		result += segment.String
	}
	return result
}

// applies all the attributes
func (s RslString) String() string {
	builder := strings.Builder{}
	for _, segment := range s.Segments {
		builder.WriteString(s.ApplyAttributes(segment.String, segment))
	}
	return builder.String()
}

func (s *RslString) ApplyAttributes(str string, segment rslStringSegment) string {
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

func (s *RslString) ToRuneList() *RslList {
	result := NewRslList()
	for i := int64(0); i < s.Len(); i++ {
		result.Append(newRslValueRslStr(s.IndexAt(i)))
	}
	return result
}

func (s *RslString) Concat(other RslString) RslString {
	return RslString{Segments: append(s.Segments, other.Segments...)}
}

func (s *RslString) ConcatStr(other string) RslString {
	return s.Concat(NewRslString(other))
}

func (s *RslString) Equals(other RslString) bool {
	return s.Plain() == other.Plain()
}

func (s RslString) Len() int64 {
	// todo also cachable
	return int64(com.StrLen(s.Plain()))
}

func (s *RslString) Index(i *Interpreter, idxNode *ts.Node) RslString {
	if idxNode.Kind() == rsl.K_SLICE {
		// todo should maintain attr info
		start, end := ResolveSliceStartEnd(i, idxNode, s.Len())
		return NewRslString(s.Plain()[start:end])
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
func (s *RslString) IndexAt(idx int64) RslString {
	cumLen := 0
	for _, segment := range s.Segments {
		nextSegmentLen := len(segment.String)
		if cumLen+nextSegmentLen > int(idx) {
			// rune array conversion required to handle multibyte characters e.g. emojis
			char := []rune(s.Plain())[idx] // todo inefficient, should just look up in segment
			return newRslStringWithAttr(string(char), segment)
		}
		cumLen += +nextSegmentLen
	}
	RP.RadErrorExit("Bug! IndexAt called with invalid index")
	panic(UNREACHABLE)
}

func (s *RslString) Compare(other RslString) int {
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

func (s *RslString) DeepCopy() RslString {
	cpy := *s
	cpy.Segments = make([]rslStringSegment, len(s.Segments))
	copy(cpy.Segments, s.Segments)
	return cpy
}

func (s *RslString) CopyWithAttr(attr RslTextAttr) RslString {
	cpy := s.DeepCopy()
	for i := range cpy.Segments {
		cpy.Segments[i].Attributes = append(cpy.Segments[i].Attributes, attr)
	}
	return cpy
}

func (s *RslString) Hyperlink(link RslString) RslString {
	cpy := s.DeepCopy()
	for i := range cpy.Segments {
		str := link.Plain()
		cpy.Segments[i].Hyperlink = &str
	}
	return cpy
}

func (s RslString) Upper() RslString {
	cpy := s.DeepCopy()
	for i, segment := range cpy.Segments {
		cpy.Segments[i].String = strings.ToUpper(segment.String)
	}
	return cpy
}

func (s RslString) Lower() RslString {
	cpy := s.DeepCopy()
	for i, segment := range cpy.Segments {
		cpy.Segments[i].String = strings.ToLower(segment.String)
	}
	return cpy
}

func (s *RslString) SetAttr(attr RslTextAttr) {
	for i := range s.Segments {
		s.Segments[i].Attributes = append(s.Segments[i].Attributes, attr)
	}
}

func (s *RslString) SetSegmentsHyperlink(link RslString) {
	for i := range s.Segments {
		str := link.Plain()
		s.Segments[i].Hyperlink = &str
	}
}

func (s *RslString) Trim(chars string) RslString {
	// todo should maintain attr info
	return NewRslString(strings.Trim(s.Plain(), chars))
}

func (s *RslString) TrimPrefix(prefix string) RslString {
	// todo should maintain attr info
	return NewRslString(strings.TrimLeft(s.Plain(), prefix))
}

func (s *RslString) TrimSuffix(suffix string) RslString {
	// todo should maintain attr info
	return NewRslString(strings.TrimRight(s.Plain(), suffix))
}

func (s *RslString) Reverse() RslString {
	// todo should maintain attr info
	return NewRslString(com.Reverse(s.Plain()))
}

func (s RslString) SetRgb(red int64, green int64, blue int64) {
	rgb := com.NewRgb64(red, green, blue)
	for i := range s.Segments {
		s.Segments[i].Rgb = &rgb
	}
}
