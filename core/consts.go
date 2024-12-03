package core

const (
	UNREACHABLE                     = "Bug! This should be unreachable"
	NOT_IMPLEMENTED                 = "not implemented"
	NO_NUM_RETURN_VALUES_CONSTRAINT = -1
)

const (
	WILDCARD = "*"
)

// function names
const (
	// todo add others
	PICK               = "pick"
	PICK_FROM_RESOURCE = "pick_from_resource"
	PICK_KV            = "pick_kv"
	PRINT              = "print"
	PPRINT             = "pprint"
	DEBUG              = "debug"
	EXIT               = "exit"
	KEYS               = "keys"
	VALUES             = "values"
	SLEEP              = "sleep"
	RAND               = "rand"
	RAND_INT           = "rand_int"
	SEED_RANDOM        = "seed_random"
	TRUNCATE           = "truncate"
	SPLIT              = "split"
	RANGE              = "range"
	UNIQUE             = "unique"
	SORT_FUNC          = "sort"
	CONFIRM            = "confirm"
	PARSE_JSON         = "parse_json"
	HTTP_GET           = "http_get"
	HTTP_POST          = "http_post"
	PARSE_INT          = "int"
)
