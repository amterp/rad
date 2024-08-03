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
program                    -> fileHeader? statement* EOF
fileHeader                 -> '"""' fileHeaderContents '"""'
fileHeaderContents         -> fhOneLiner ( NEWLINE NEWLINE fhLongDescription )?
fhOneLiner                 -> .*
fhLongDescription          -> ( .* NEWLINE )+
statement                  -> assignment
                              | rad
                              | forStmt
assignment                 -> argBlock // todo maybe this should not be an assignment, but a once-off at the start
                              | jsonFieldAssignment
                              | ifStmt
                              | choiceAssignment
                              | choiceResourceAssignment
                              | primaryAssignment
argBlock                   -> "args" COLON NEWLINE ( INDENT argBlockStmt NEWLINE )*
argBlockStmt               -> argDeclaration
                              | argBlockConstraint
INDENT                     -> "  " | "   " | "    " | "\t"
argDeclaration             -> IDENTIFIER FLAG? ARG_TYPE argOptional? ARG_COMMENT
IDENTIFIER                 -> [A-Za-z_][A-Za-z0-9_]+ // probably overly restrictive
FLAG                       -> [A-Za-z0-9_]  // probably overly restrictive
ARG_TYPE                   -> ( ( "string" | "int" ) BRACKETS? ) | bool
BRACKETS                   -> "[]"
argOptional                -> argOptionalNoDefault | argOptionalDefault
argOptionalNoDefault       -> "?"
argOptionalDefault         -> "=" primary
ARG_COMMENT                -> "#" .*
argBlockConstraint         -> argStringRegexConstraint
                              | argIntRangeConstraint
                              | argOneWayReq
                              | argMutualExcl
argStringRegexConstraint   -> IDENTIFIER ( "," IDENTIFIER )* "not"? "regex" REGEX
argIntRangeConstraint      -> IDENTIFIER COMPARATORS INT
argOneWayReq               -> IDENTIFIER "requires" IDENTIFIER
argMutualExcl              -> "one_of" IDENTIFIER ( "," IDENTIFIER )+
jsonFieldAssignment        -> IDENTIFIER "=" "json" BRACKETS? ( "." jsonFieldPathElement )*
jsonFieldPathElement       -> jsonFieldPathKey BRACKETS?
jsonFieldPathKey           -> ( escapedKeyChar | .* -- \ . [ )*
escapedKeyChar             -> '\' .*
ifStmt                     -> "if" expression COLON NEWLINE ( INDENT statement NEWLINE )* ( elseIf | else )?
elseIf                     -> "else" ifStmt
else                       -> "else" COLON NEWLINE ( INDENT statement NEWLINE )* // prob not correct, I think dangling stmts are a risk
choiceAssignment           -> IDENTIFIER ( "," IDENTIFIER )* "=" "choice" discriminator? ( choiceBlock | choiceFromResource )
discriminator              -> IDENTIFIER
choiceBlock                -> COLON NEWLINE ( INDENT choiceOption NEWLINE )+
choiceOption               -> primary ( "," primary )* choiceOptionTags?
choiceOptionTags           -> "|" basic ( "," basic )*  
choiceFromResource         -> "from" RESOURCE "on" IDENTIFIER
choiceResourceAssignment   -> IDENTIFIER "=" "resource" "choice" choiceBlock
RESOURCE                   -> STRING
primaryAssignment          -> IDENTIFIER "=" primary
rad                        -> "rad" IDENTIFIER? COLON NEWLINE radBody
radBody                    -> INDENT IDENTIFIER ( "," IDENTIFIER )* NEWLINE ( INDENT radBodyStmt NEWLINE)*
radBodyStmt                -> radSortStmt
                              | radModifierStmt
                              | radTableFormatStmt
                              | radFieldFormatBlock
radSortStmt                -> "sort" IDENTIFIER SORT? ( "," IDENTIFIER SORT? )*
radModifierStmt            -> "uniq" | "quiet" | ( "limit" primary )
radTableFormatStmt         -> "table" ( "default" | "markdown" | "fancy" ) 
radFieldFormatBlock        -> IDENTIFIER COLON NEWLINE ( INDENT radFieldFormatStmt NEWLINE )+
radFieldFormatStmt         -> radFieldFormatMaxWidthStmt
                              | radFieldFormatColorStmt
radFieldFormatMaxWidthStmt -> "max_width" INT
radFieldFormatColorStmt    -> "color" COLOR REGEX?
SORT                       -> "asc" | "desc"
forStmt                    -> "for" IDENTIFIER "in" IDENTIFIER COLON NEWLINE ( INDENT statement NEWLINE )*

expression                 -> logic_or // functions should probably fit somewhere into this structure
logic_or                   -> logic_and ( "or" logic_and )*
logic_and                  -> equality ( "and" equality )*
equality                   -> comparison ( ( NOT_EQUAL | EQUAL ) comparison )*
comparison                 -> unary ( ( GT | GTE | LT | LTE ) unary )* // here is where I *could* allow arithmetic, but choose not to (yet?)
unary                      -> ( "!" | "-" ) unary
                              | primary
primary                    -> basic | NULL | "(" expression ")" | IDENTIFIER
basic                      -> STRING | INT | BOOL
STRING                     -> '"' .* '"' // with escaping of quotes using \
INT                        -> [0-9]+
BOOL                       -> "true" | "false"
REGEX                      -> a regex
COMPARATORS                -> GT | GTE | EQUAL | LT | LTE
GT                         -> ">"
GTE                        -> ">="
LT                         -> "<"
LTE                        -> "<="
EQUAL                      -> "=="
NOT_EQUAL                  -> "!="
NULL                       -> "null"
COLOR                      -> "red" | "green" | "blue" | "yellow" | "orange" // etc etc, whatever CLIs usually support
```

TODO:

- print statement, functions
- headerStmt
- max width for the whole table
- displaying not as a table, but as pure printed lines? and other things
