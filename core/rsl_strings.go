package core

import "strings"

type RslString struct {
	Segments []rslStringSegment
	// todo offer isRegex funcs which will return true if all segments are regexes, for example?
	//  concat between non-regex and regex give non-regex? or regex?
}

type rslStringSegment struct {
	String string
	Color  *RslColor
	// can add bold, etc here too
}

// todo should these methods be returning *RslString?
func NewRslString(str string) RslString {
	return RslString{Segments: []rslStringSegment{{String: str}}}
}

func newRslStringWithAttr(str string, segment rslStringSegment) RslString {
	// todo the fact that this copies without me explicitly telling it probably means it's very wasteful
	segment.String = str
	return RslString{Segments: []rslStringSegment{segment}}
}

func (s *RslString) Plain() string {
	// todo can lazily compute and cache
	var result string
	for _, segment := range s.Segments {
		result += segment.String
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

func (s *RslString) Len() int64 {
	// todo also cachable
	return int64(len(s.Plain()))
}

// assumes idx is valid for this string
func (s *RslString) IndexAt(idx int64) RslString {
	cumLen := 0
	for _, segment := range s.Segments {
		nextSegmentLen := len(segment.String)
		if cumLen+nextSegmentLen > int(idx) {
			char := s.Plain()[idx] // todo inefficient, should just look up in segment
			return newRslStringWithAttr(string(char), segment)
		}
		cumLen += +nextSegmentLen
	}
	RP.RadErrorExit("Bug! IndexAt called with invalid index")
	panic(UNREACHABLE)
}

func (s *RslString) Slice(start int64, end int64) RslString {
	// todo should maintain attr info
	return NewRslString(s.Plain()[start:end])
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

func (s *RslString) Upper() RslString {
	cpy := *s
	for i, segment := range cpy.Segments {
		cpy.Segments[i].String = strings.ToUpper(segment.String)
	}
	return cpy
}

func (s *RslString) Lower() RslString {
	cpy := *s
	for i, segment := range cpy.Segments {
		cpy.Segments[i].String = strings.ToLower(segment.String)
	}
	return cpy
}
