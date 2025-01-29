package analysis

import "rls/log"

type DocState struct {
	uri  string
	text string
}

func NewDocState(uri, text string) DocState {
	return DocState{uri: uri, text: text}
}

type State struct {
	// URI -> Text
	docs map[string]DocState
}

func NewState() *State {
	return &State{docs: make(map[string]DocState)}
}

func (s *State) AddDoc(uri, text string) {
	log.L.Infof("Adding doc %s", uri)
	s.docs[uri] = NewDocState(uri, text)
}
