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
program                    -> fileHeader? argBlock? statement* EOF
fileHeader                 -> '"""' ( NEWLINE fileHeaderContents NEWLINE )? '"""'
fileHeaderContents         -> fhOneLiner ( NEWLINE NEWLINE fhLongDescription )?
fhOneLiner                 -> .*
fhLongDescription          -> ( .* NEWLINE )+
statement                  -> assignment
                              | rad
                              | forStmt
                              | ifStmt
                              | switchStmt
                              | exprStmt
assignment                 -> jsonFieldAssignment
                              | switchAssignment
                              | switchResourceAssignment // todo, should split into separate 'resource' interpreter?
                              | primaryAssignment
argBlock                   -> "args" COLON NEWLINE ( INDENT argBlockStmt NEWLINE )*
argBlockStmt               -> argDeclaration
                              | argBlockConstraint
INDENT                     -> "  " | "   " | "    " | "\t"
argDeclaration             -> IDENTIFIER STRING? FLAG? ARG_TYPE argOptional? ARG_COMMENT
IDENTIFIER                 -> [A-Za-z_][A-Za-z0-9_]+ // probably overly restrictive
FLAG                       -> [A-Za-z0-9_]  // probably overly restrictive
ARG_TYPE                   -> ( ( "string" | "int" ) BRACKETS? ) | bool
BRACKETS                   -> "[]"
argOptional                -> argOptionalNoDefault | argOptionalDefault
argOptionalNoDefault       -> "?"
argOptionalDefault         -> "=" primary
ARG_COMMENT                -> "#" .*
argBlockConstraint         -> argStringRegexConstraint
                              | argNumberRangeConstraint
                              | argOneWayReq
                              | argMutualExcl
argStringRegexConstraint   -> IDENTIFIER ( "," IDENTIFIER )* "not"? "regex" REGEX
argNumberRangeConstraint   -> IDENTIFIER COMPARATORS NUMBER
argOneWayReq               -> IDENTIFIER "requires" IDENTIFIER
argMutualExcl              -> "one_of" IDENTIFIER ( "," IDENTIFIER )+
jsonFieldAssignment        -> IDENTIFIER "=" "json" BRACKETS? ( "." jsonFieldPathElement )*
jsonFieldPathElement       -> jsonFieldPathKey BRACKETS?
jsonFieldPathKey           -> ( escapedKeyChar | .* -- \ . [ )*
escapedKeyChar             -> '\' .*
ifStmt                     -> "if" expression COLON NEWLINE ( INDENT statement NEWLINE )* ( elseIf | else )?
elseIf                     -> "else" ifStmt
else                       -> "else" COLON NEWLINE ( INDENT statement NEWLINE )* // prob not correct, I think dangling stmts are a risk
switchAssignment           -> IDENTIFIER ( "," IDENTIFIER )* "=" "switch" discriminator? ( switchBlock | switchOnResource )
discriminator              -> IDENTIFIER
switchBlock                -> COLON NEWLINE ( INDENT switchCase NEWLINE )+
switchCase                 -> "case" primary ( "," primary )* COLON NEWLINE ( INDENT statement NEWLINE )*
switchOnResource           -> "on" RESOURCE IDENTIFIER
switchResourceAssignment   -> IDENTIFIER "=" "resource" "switch" COLON NEWLINE ( INDENT switchResourceCase NEWLINE )*
switchResourceCase         -> "case" primary ( "," primary )* COLON primary ( "," primary )*
RESOURCE                   -> STRING
primaryAssignment          -> IDENTIFIER "=" primaryExpr
rad                        -> "rad" IDENTIFIER? COLON NEWLINE ( INDENT radStmt NEWLINE )*
radStmt                    -> radIfStmt
                              | radFieldsStmt
                              | radSortStmt
                              | radModifierStmt
                              | radTableFormatStmt
                              | radFieldFormatBlock
radIfStmt                  -> "if" expression COLON NEWLINE ( INDENT radStmt NEWLINE )* ( radElseIf | radElse )?
radElseIf                  -> "else" radIfStmt
radElse                    -> "else" COLON NEWLINE ( INDENT radStmt NEWLINE )*
radFieldsStmt              -> "fields" IDENTIFIER ( "," IDENTIFIER )*
radSortStmt                -> "sort" IDENTIFIER SORT? ( "," IDENTIFIER SORT? )*
radModifierStmt            -> "uniq" | "quiet" | ( "limit" primaryExpr )
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
primaryExpr                -> primary | "(" expression ")"
primary                    -> literal | NULL | IDENTIFIER
literal                    -> STRING | NUMBER | BOOL // 'ANY' might need to be one? or just string in such cases?
switchStmt                 -> "switch" discriminator switchBlock
exprStmt                   -> expression ( "," expression )*

STRING                     -> '"' .* '"' // with escaping of quotes using \
NUMBER                     -> INT | FLOAT
INT                        -> [0-9]+
FLOAT                      -> [0-9]+.[0-9]+
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
- consider while loop? clear use cases not immediately clear but in theory allows for big step up in capability
