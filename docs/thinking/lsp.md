# Language Server / Syntax Highlighting

- TextMate
- Pygment
- LSPs

---

- Pygments require a lexer implementation
- Sounds like LSPs do too, but can delegate to a Pygment lexer implementation?
- If I do LSP/Pygment combo, TextMate is probably redundant?
- Sounds like LSPs *can* generate highlighting without Pygment?
- Spec: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/

---

## Deduplication

- I think we risk duplicating lexer/parser impls between the core interpreter and these 'helper' tools. How avoid?
- Can potentially use lexer/parser generators like ANTLR
  - Define RSL grammar, generate different language e.g. Go implementations of lexer/parser.
  - Seems like it'd be a massive overhaul - probably something I should have done from the start if I had the foresight.
- Alternatively, can modify lexer/parser to be more shareable, and re-use between the interpreter and LSP implementations.
  - In this picture, avoiding Pygment and TextMate entirely might be a good way to avoid duplication and major overhauls.
  - Rely on LSP for syntax highlighting.
  - Unclear if we can avoid Pygment tho, if MkDocs requires it for highlighting.
- Can you have a Pygment implementation which simply invokes the Go 'lexer as a library' and forwards whatever it needs to?

## Tools

- ANTLR
- Tree-sitter
- Alternatives to Pygments for Material for MkDocs:
  - Highlight.js? Prism.js?

---

- https://www.youtube.com/watch?v=EkK8Jxjj95s
- Consider making at least the lexer into a library
  - Text doc in, tokens out, including text ranges start/end
- Consider making a 'hello world' parser which simply highlights instances of =, for example
- Consider [Parse lib](https://github.com/a-h/parse)?
