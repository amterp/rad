package core

func (i *MainInterpreter) mustIdentifier(tkn Token, path VarPath) Token {
	identifier := path.Identifier
	if identifier == nil {
		i.error(tkn, "Expected identifier")
	}
	return path.Identifier
}
