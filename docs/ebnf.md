# EBNF

The [Extended Backus-Naur Form (EBNF)](https://en.wikipedia.org/wiki/Extended_Backus%E2%80%93Naur_form) for the Rad Scripting Language (RSL).

Not *exactly* EBNF, I prefer the syntax used in [Crafting Interpreters](https://craftinginterpreters.com/), and I don't have formal training here, so will write it perhaps loosely!

```
program                   -> fileHeader? statement* EOF
fileHeader                -> '"""' fileHeaderContents '"""'
fileHeaderContents        -> fhOneLiner ( NEWLINE NEWLINE fhLongDescription )?
fhOneLiner                -> .*
fhLongDescription         -> ( .* NEWLINE )+
statement                 -> assignment
assignment                -> argBlock
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
argIntRangeConstraint     -> IDENTIFIER COMPARISON INT
argOneWayReq              -> IDENTIFIER "requires" IDENTIFIER
argMutualExcl             -> "one_of" IDENTIFIER ( "," IDENTIFIER )+
primary                   -> STRING | INT | BOOL
STRING                    -> '"' .* '"' // with escaping of quotes using \
INT                       -> [0-9]+
BOOL                      -> "true" | "false"
REGEX                     -> a regex
COMPARISON                -> ">" | ">=" | "==" | "<" | "<="
```

TODO:

- print statement, functions
- headerStmt
