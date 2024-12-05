package core

type RslError string

const (
	// RAD1xxxx Syntax Errors
	// RAD2xxxx Runtime Errors
	PARSE_INT_FAILED   RslError = "RAD20001"
	PARSE_FLOAT_FAILED RslError = "RAD20002"
	// RAD3xxxx Type Errors?
	// RAD4xxxx Validation Errors?
)
