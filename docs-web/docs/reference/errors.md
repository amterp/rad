---
title: Errors
---

Rad errors include a code (e.g. `RAD10001`) to help identify the type of error. This page documents all error codes.

## RAD1xxxx - Syntax Errors

| Code     | Name                | Description                         |
|----------|---------------------|-------------------------------------|
| RAD10001 | InvalidSyntax       | General syntax error                |
| RAD10002 | MissingColon        | Expected a colon                    |
| RAD10003 | MissingIdentifier   | Expected an identifier              |
| RAD10004 | MissingExpression   | Expected an expression              |
| RAD10005 | MissingCloseParen   | Expected closing parenthesis `)`    |
| RAD10006 | MissingCloseBracket | Expected closing bracket `]`        |
| RAD10007 | MissingCloseBrace   | Expected closing brace `}`          |
| RAD10008 | ReservedKeyword     | Attempted to use a reserved keyword |
| RAD10009 | UnexpectedToken     | Encountered an unexpected token     |

## RAD2xxxx - Runtime Errors

| Code     | Name                 | Description                                                  |
|----------|----------------------|--------------------------------------------------------------|
| RAD20000 | GenericRuntime       | General runtime error                                        |
| RAD20001 | ParseIntFailed       | `parse_int` failed to parse the input                        |
| RAD20002 | ParseFloatFailed     | `parse_float` failed to parse the input                      |
| RAD20003 | FileRead             | Failed to read the specified file                            |
| RAD20004 | FileNoPermission     | Did not have permission to access the specified file         |
| RAD20005 | FileNoExist          | The specified file or directory does not exist               |
| RAD20006 | FileWrite            | Failed to write to the specified file                        |
| RAD20007 | AmbiguousEpoch       | Ambiguous epoch length; use `unit` parameter to disambiguate |
| RAD20008 | InvalidTimeUnit      | Invalid time unit specified                                  |
| RAD20009 | InvalidTimeZone      | Invalid time zone specified                                  |
| RAD20010 | UserInput            | Error reading user input                                     |
| RAD20011 | ParseJson            | Failed to parse JSON                                         |
| RAD20012 | BugTypeCheck         | Internal type check error (likely a bug)                     |
| RAD20013 | FileWalk             | Error walking file tree                                      |
| RAD20014 | MutualExclArgs       | Mutually exclusive arguments were both provided              |
| RAD20015 | ZipStrict            | `zip` with `strict=true` received lists of different lengths |
| RAD20016 | Cast                 | Type cast failed                                             |
| RAD20017 | NumInvalidRange      | Number is outside valid range                                |
| RAD20018 | EmptyList            | Operation requires a non-empty list                          |
| RAD20019 | ArgsContradict       | Argument constraints contradict each other                   |
| RAD20020 | Fid                  | Field ID error                                               |
| RAD20021 | Decode               | Decoding error                                               |
| RAD20022 | NoStashId            | Script has no stash ID configured                            |
| RAD20023 | SleepStr             | Invalid sleep duration string                                |
| RAD20024 | InvalidRegex         | Invalid regular expression                                   |
| RAD20025 | ColorizeValNotInEnum | Colorize value not in enum                                   |
| RAD20026 | StdinRead            | Error reading from stdin                                     |
| RAD20027 | InvalidCheckDuration | Invalid check duration                                       |
| RAD20028 | KeyNotFound          | Map key does not exist                                       |
| RAD20029 | IndexOutOfBounds     | List or string index is out of bounds                        |

## RAD4xxxx - Validation Errors

| Code     | Name                             | Description                                                             |
|----------|----------------------------------|-------------------------------------------------------------------------|
| RAD40001 | ScientificNotationNotWholeNumber | Scientific notation resulted in non-whole number where integer expected |
| RAD40002 | HoistedFunctionShadowsArgument   | A hoisted function name shadows an argument                             |
| RAD40003 | UnknownFunction                  | Reference to unknown function                                           |
