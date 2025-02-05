package core

const (
	// Node Kinds
	K_COMMENT      = "comment"
	K_SOURCE_FILE  = "source_file"
	K_SHEBANG      = "shebang"
	K_FILE_HEADER  = "file_header"
	K_ARG_BLOCK    = "arg_block"
	K_ASSIGN       = "assign"
	K_EXPR_STMT    = "expr_stmt"
	K_EXPR         = "expr"
	K_PRIMARY_EXPR = "primary_expr"
	K_LITERAL      = "literal"
	K_VAR_PATH     = "var_path"
	K_INT          = "int"
	K_LIST         = "list"
	K_CALL         = "call"

	// Field names
	F_LEFT       = "left"
	F_RIGHT      = "right"
	F_ROOT       = "root"
	F_INDEXING   = "indexing"
	F_LIST_ENTRY = "list_entry"
	F_FUNC       = "func"
	F_ARG        = "arg"
)
