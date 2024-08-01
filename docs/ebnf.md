# EBNF

The [Extended Backus-Naur Form (EBNF)](https://en.wikipedia.org/wiki/Extended_Backus%E2%80%93Naur_form) for the Rad Scripting Language (RSL).

Not *exactly* EBNF, I prefer the syntax used in [Crafting Interpreters](https://craftinginterpreters.com/), and I don't have formal training here, so will write it perhaps loosely!

Syntax Legend:

*Somewhat typical EBNF but closer to regex syntax*

- `a | b` : a or b
- `a?` : one 'a' may or may not be present. optional.
- `( a b )?` : 'a followed by b' is optional
- `a*` : 0 or more 'a' 
- `a+` : 1 or more 'a'
- `.* -- & $` : any char except '&' or '$'

```
program                   -> fileHeader? statement* EOF
fileHeader                -> '"""' fileHeaderContents '"""'
fileHeaderContents        -> fhOneLiner ( NEWLINE NEWLINE fhLongDescription )?
fhOneLiner                -> .*
fhLongDescription         -> ( .* NEWLINE )+
statement                 -> assignment
assignment                -> argBlock // todo maybe this should not be an assignment, but a once-off at the start
                             | jsonFieldAssignment
                             | ifStmt
argBlock                  -> "args" COLON NEWLINE ( INDENT argBlockStmt NEWLINE )*
argBlockStmt              -> argDeclaration
                             | argBlockConstraint
INDENT                    -> "  " | "   " | "    " | "\t"
argDeclaration            -> IDENTIFIER FLAG? ARG_TYPE argOptional? ARG_COMMENT
IDENTIFIER                -> [A-Za-z_][A-Za-z0-9_]+ // probably overly restrictive
FLAG                      -> [A-Za-z0-9_]  // probably overly restrictive
ARG_TYPE                  -> ( ( "string" | "int" ) BRACKETS? ) | bool
BRACKETS                  -> "[]"
argOptional               -> argOptionalNoDefault | argOptionalDefault
argOptionalNoDefault      -> "?"
argOptionalDefault        -> "=" primary
ARG_COMMENT               -> "#" .*
argBlockConstraint        -> argStringRegexConstraint
                             | argIntRangeConstraint
                             | argOneWayReq
                             | argMutualExcl
argStringRegexConstraint  -> IDENTIFIER ( "," IDENTIFIER )* "regex" REGEX
argIntRangeConstraint     -> IDENTIFIER COMPARATORS INT
argOneWayReq              -> IDENTIFIER "requires" IDENTIFIER
argMutualExcl             -> "one_of" IDENTIFIER ( "," IDENTIFIER )+
jsonFieldAssignment       -> IDENTIFIER "=" "json" BRACKETS? ( "." jsonFieldPathElement )*
jsonFieldPathElement      -> jsonFieldPathKey BRACKETS?
jsonFieldPathKey          -> ( escapedKeyChar | .* -- \ . [ )*
escapedKeyChar            -> '\' .*
ifStmt                    -> "if" expression COLON NEWLINE ( INDENT statement NEWLINE )* ( elseIf | else )?
elseIf                    -> "else" ifStmt
else                      -> "else" COLON NEWLINE ( INDENT statement NEWLINE )* // prob not correct, I think dangling stmts are a risk
expression                -> logic_or
logic_or                  -> logic_and ( "or" logic_and )*
logic_and                 -> equality ( "and" equality )*
equality                  -> comparison ( ( NOT_EQUAL | EQUAL ) comparison )*
comparison                -> unary ( ( GT | GTE | LT | LTE ) unary )* // here is where I *could* allow arithmetic, but choose not to (yet?)
unary                     -> ( "!" | "-" ) unary
                             | primary
primary                   -> STRING | INT | BOOL | NULL | "(" expression ")" | IDENTIFIER
STRING                    -> '"' .* '"' // with escaping of quotes using \
INT                       -> [0-9]+
BOOL                      -> "true" | "false"
REGEX                     -> a regex
COMPARATORS               -> GT | GTE | EQUAL | LT | LTE
GT                        -> ">"
GTE                       -> ">="
LT                        -> "<"
LTE                       -> "<="
EQUAL                     -> "=="
NOT_EQUAL                 -> "!="
NULL                      -> "null"
```

TODO:

- print statement, functions
- headerStmt
