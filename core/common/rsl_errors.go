package com

type RslError string

// todo can we avoid 5 digits?

const (
	// RAD1xxxx Syntax Errors
	// RAD2xxxx Runtime Errors
	ErrParseIntFailed   RslError = "RAD20001"
	ErrParseFloatFailed          = "RAD20002"
	ErrFileRead                  = "RAD20003"
	ErrFileNoPermission          = "RAD20004"
	ErrFileNoExist               = "RAD20005"
	// RAD3xxxx Type Errors?
	// RAD4xxxx Validation Errors?
)
