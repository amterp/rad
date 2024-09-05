package core

var GLOBAL_KEYWORDS = map[string]TokenType{
	"for":      FOR,
	"in":       IN,
	"args":     ARGS,
	"switch":   SWITCH,
	"from":     FROM,
	"on":       ON,
	"rad":      RAD,
	"or":       OR,
	"and":      AND,
	"not":      NOT,
	"if":       IF,
	"else":     ELSE,
	"resource": RESOURCE,
}

var ARGS_BLOCK_KEYWORDS = map[string]TokenType{
	"string":   STRING,
	"int":      INT,
	"float":    FLOAT,
	"bool":     BOOL,
	"requires": REQUIRES,
	"one_of":   ONE_OF,
	"regex":    REGEX,
}

var RAD_BLOCK_KEYWORDS = map[string]TokenType{
	"fields":    FIELDS,
	"sort":      SORT,
	"asc":       ASC,
	"desc":      DESC,
	"color":     COLOR,
	"max_width": MAX_WIDTH,
	"uniq":      UNIQ,
	"quiet":     QUIET,
	"limit":     LIMIT,
	"table":     TABLE,
	"default":   DEFAULT,
	"markdown":  MARKDOWN,
}

var SWITCH_BLOCK_KEYWORDS = map[string]TokenType{
	"case":    CASE,
	"default": DEFAULT,
}

var ALL_KEYWORDS = mergeMaps(GLOBAL_KEYWORDS, ARGS_BLOCK_KEYWORDS, RAD_BLOCK_KEYWORDS, SWITCH_BLOCK_KEYWORDS)

func mergeMaps(maps ...map[string]TokenType) map[string]TokenType {
	result := make(map[string]TokenType)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
