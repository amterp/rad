package core

type RadError struct {
	msg RadString
}

func NewError(msg RadString) RadError {
	return RadError{msg: msg}
}

func (e RadError) Msg() RadString {
	return e.msg
}

func (e RadError) Equals(other RadError) bool {
	return e.Msg().Equals(other.Msg())
}

func (e RadError) Hash() string {
	return e.Msg().Plain()
}
