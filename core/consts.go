package core

const (
	UNREACHABLE                     = "Bug! This should be unreachable"
	NOT_IMPLEMENTED                 = "not implemented"
	NO_NUM_RETURN_VALUES_CONSTRAINT = -1
	USAGE_ALIGNMENT_CHAR            = "\x00"
	PADDING_CHAR                    = "\x00"
)

const (
	WILDCARD = "*"
)

// function names
const (
	TRUNCATE    = "truncate"
	SPLIT       = "split"
	RANGE       = "range"
	UNIQUE      = "unique"
	CONFIRM     = "confirm"
	PARSE_JSON  = "parse_json"
	HTTP_GET    = "http_get"
	HTTP_POST   = "http_post"
	HTTP_PUT    = "http_put"
	PARSE_INT   = "parse_int"
	PARSE_FLOAT = "parse_float"
	ABS         = "abs"
)
