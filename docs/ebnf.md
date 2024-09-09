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
program                     -> shebang? fileHeader? argBlock? statement* EOF
shebang                     -> "#!" .* NEWLINE
fileHeader                  -> '---' ( NEWLINE fileHeaderContents NEWLINE )? '---'
fileHeaderContents          -> fhOneLiner ( NEWLINE NEWLINE fhLongDescription )?
fhOneLiner                  -> .*
fhLongDescription           -> ( .* NEWLINE )+
statement                   -> assignment
                               | radBlock
                               | queryBlock
                               | tblBlock
                               | forStmt
                               | ifStmt
                               | switchStmt
                               | exprStmt
assignment                  -> jsonFieldAssignment
                               | switchAssignment
                               | switchResourceAssignment // todo, should split into separate 'resource' interpreter?
                               | compoundAssignment
                               | arrayAssignment
                               | expressionAssignment
argBlock                    -> "args" COLON NEWLINE ( INDENT argBlockStmt NEWLINE )*
argBlockStmt                -> argDeclaration
                               | argBlockConstraint
INDENT                      -> "  " | "   " | "    " | "\t"
argDeclaration              -> IDENTIFIER STRING? FLAG? anyType argOptional? ARG_COMMENT
IDENTIFIER                  -> [A-Za-z_][A-Za-z0-9_]+ // probably overly restrictive
FLAG                        -> [A-Za-z0-9_]  // probably overly restrictive
anyType                     -> primitiveType BRACKETS?
arrayType                   -> primitiveType BRACKETS
primitiveType               -> "string" | "int" | "float" | "bool"
BRACKETS                    -> "[]"
argOptional                 -> argOptionalNoDefault | argOptionalDefault
argOptionalNoDefault        -> "?"
argOptionalDefault          -> "=" literalOrArray
ARG_COMMENT                 -> "#" .*
argBlockConstraint          -> argStringRegexConstraint
                               | argNumberRangeConstraint
                               | argOneWayReq
                               | argsSpecifiedConstraint
argStringRegexConstraint    -> IDENTIFIER ( "," IDENTIFIER )* "not"? "regex" REGEX
argNumberRangeConstraint    -> IDENTIFIER COMPARATORS NUMBER
argOneWayReq                -> IDENTIFIER "requires" IDENTIFIER
argsSpecifiedConstraint     -> ( "at_least" | "exactly" | "at_most" ) INT IDENTIFIER ( "," IDENTIFIER )+
jsonFieldAssignment         -> IDENTIFIER "=" "json" BRACKETS? ( "." jsonFieldPathElement )*
jsonFieldPathElement        -> jsonFieldPathKey BRACKETS?
jsonFieldPathKey            -> ( escapedKeyChar | .* -- \ . [ )*
escapedKeyChar              -> '\' .*
ifStmt                      -> "if" expression COLON NEWLINE ( INDENT statement NEWLINE )* ( elseIf | else )?
elseIf                      -> "else" ifStmt
else                        -> "else" COLON NEWLINE ( INDENT statement NEWLINE )* // prob not correct, I think dangling stmts are a risk
switchAssignment            -> IDENTIFIER ( "," IDENTIFIER )* "=" "switch" discriminator? ( switchBlock | switchOnResource )
discriminator               -> IDENTIFIER
switchBlock                 -> COLON NEWLINE ( INDENT switchCase NEWLINE )+
switchCase                  -> "case" literal ( "," literal )* COLON NEWLINE ( INDENT statement NEWLINE )*
switchOnResource            -> "on" RESOURCE IDENTIFIER
switchResourceAssignment    -> IDENTIFIER "=" "resource" "switch" COLON NEWLINE ( INDENT switchResourceCase NEWLINE )*
switchResourceCase          -> "case" literal ( "," literal )* COLON expression ( "," expression )*
RESOURCE                    -> STRING
compoundAssignment          -> addCompoundAssignemnt
                               | minusCompoundAssignemnt
                               | multiplyCompoundAssignemnt
                               | divideCompoundAssignemnt
addCompoundAssignment       -> IDENTIFIER "+=" IDENTIFIER
minusCompoundAssignment     -> IDENTIFIER "-=" IDENTIFIER
multiplyCompoundAssignment  -> IDENTIFIER "*=" IDENTIFIER
divideCompoundAssignment    -> IDENTIFIER "/=" IDENTIFIER
arrayAssignment             -> IDENTIFIER arrayType "=" arrayExpr
arrayExpr                   -> "[" ( expression ( "," expression )* )? "]"
expressionAssignment        -> IDENTIFIER primitiveType? "=" expression
radBlock                    -> "rad" IDENTIFIER COLON NEWLINE ( INDENT radStmt NEWLINE )*
radStmt                     -> radIfStmt
                               | queryFieldsStmt
                               | queryHeaderStmt
                               | tblSortStmt
                               | radModifierStmt
                               | tblStyleStmt
                               | tblFieldFormatBlock
radIfStmt                   -> "if" expression COLON NEWLINE ( INDENT radStmt NEWLINE )* ( radElseIf | radElse )?
radElseIf                   -> "else" radIfStmt
radElse                     -> "else" COLON NEWLINE ( INDENT radStmt NEWLINE )*
queryFieldsStmt             -> "fields" IDENTIFIER ( "," IDENTIFIER )*
queryHeaderStmt             -> "header" expression
queryModifierStmt           -> "quiet"
tblModifierStmt             -> "uniq" | ( "limit expression )
tblSortStmt                 -> "sort" IDENTIFIER SORT? ( "," IDENTIFIER SORT? )*
radModifierStmt             -> queryModifierStmt | tblModifierStmt
tblStyleStmt                -> "style" ( "default" | "markdown" | "fancy" ) 
tblFieldFormatBlock         -> IDENTIFIER ( "," IDENTIFIER )* COLON NEWLINE ( INDENT tblFieldFormatStmt NEWLINE )+
tblFieldFormatStmt          -> tblFormatMaxWidthStmt
                               | tblFormatColorStmt
tblFormatMaxWidthStmt       -> "max_width" INT
tblFormatColorStmt          -> "color" COLOR REGEX?
SORT                        -> "asc" | "desc"
queryBlock                  -> "query" IDENTIFIER COLON NEWLINE ( INDENT queryStmt NEWLINE )*
queryStmt                   -> queryFieldsStmt
                               | queryHeaderStmt
                               | queryModifierStmt
                               | queryIfStmt
queryIfStmt                 -> "if" expression COLON NEWLINE ( INDENT queryStmt NEWLINE )* ( queryElseIf | queryElse )?
queryElseIf                 -> "else" queryIfStmt
queryElse                   -> "else" COLON NEWLINE ( INDENT queryStmt NEWLINE )*
tblBlock                    -> "table" COLON NEWLINE ( INDENT tblStmt NEWLINE )*
tblStmt                     -> tblFieldsStmt
                               | tblModifierStmt
                               | tblSortStmt
                               | tblStyleStmt
                               | tblFieldFormatBlock
                               | tblIfStmt
tblIfStmt                   -> "if" expression COLON NEWLINE ( INDENT tblStmt NEWLINE )* ( tblElseIf | tblElse )?
tblElseIf                   -> "else" tblIfStmt
tblElse                     -> "else" COLON NEWLINE ( INDENT tblStmt NEWLINE )*
forStmt                     -> "for" IDENTIFIER ( forStmtIndex | forStmtNoIndex )
forStmtIndex                -> "," IDENTIFIER forStmtNoIndex
forStmtNoIndex              -> "in" IDENTIFIER COLON NEWLINE ( INDENT statement NEWLINE )*

expression                  -> logic_or
logic_or                    -> logic_and ( "or" logic_and )*
logic_and                   -> equality ( "and" equality )*
equality                    -> comparison ( ( NOT_EQUAL | EQUAL ) comparison )*
comparison                  -> term ( ( GT | GTE | LT | LTE ) term )*
term                        -> factor ( ( "-" | "+" ) factor )*
factor                      -> unary ( ( "/" | "*" ) unary )*
unary                       -> ( "!" | "-" ) unary
                               | primary
primary                     -> "(" expression ")" | literalOrArray | arrayExpr | arrayAccess | functionCall | IDENTIFIER
literalOrArray              -> literal | arrayLiteral
literal                     -> STRING | NUMBER | BOOL
arrayLiteral                -> "[" ( literal ( "," literal )* )? "]"
arrayAccess                 -> IDENTIFIER "[" expression "]"
functionCall                -> IDENTIFIER "(" ( ( expression ( "," expression )* )? ( IDENTIFIER "=" expression ( "," IDENTIFIER "=" expression )* )? )? ")"
switchStmt                  -> "switch" discriminator switchBlock
exprStmt                    -> expression ( "," expression )*

STRING                      -> '"' .* '"' // with escaping of quotes using \
NUMBER                      -> INT | FLOAT
INT                         -> [0-9]+
FLOAT                       -> [0-9]+.[0-9]+
BOOL                        -> "true" | "false"
REGEX                       -> a regex
COMPARATORS                 -> GT | GTE | EQUAL | LT | LTE
GT                          -> ">"
GTE                         -> ">="
LT                          -> "<"
LTE                         -> "<="
EQUAL                       -> "=="
NOT_EQUAL                   -> "!="
COLOR                       -> "red" | "green" | "blue" | "yellow" | "orange" // etc etc, whatever CLIs usually support
```

TODO:

- print statement, functions
- headerStmt
- max width for the whole tbl
- displaying not as a tbl, but as pure printed lines? and other things
- consider while loop? clear use cases not immediately clear but in theory allows for big step up in capability
