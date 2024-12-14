package core

func (p *Parser) GetIdentifiers(paths []VarPath) []Token {
	var identifiers []Token
	for _, path := range paths {
		identifier := path.Identifier
		if identifier == nil {
			p.error("Expected identifier for variable assignment")
		}
		identifiers = append(identifiers, path.Identifier)
	}
	return identifiers
}
