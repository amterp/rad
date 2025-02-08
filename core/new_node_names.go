package core

const (
	// Node Kinds
	K_COMMENT                 = "comment"
	K_SOURCE_FILE             = "source_file"
	K_SHEBANG                 = "shebang"
	K_FILE_HEADER             = "file_header"
	K_ARG_BLOCK               = "arg_block"
	K_ASSIGN                  = "assign"
	K_EXPR_STMT               = "expr_stmt"
	K_EXPR                    = "expr"
	K_PRIMARY_EXPR            = "primary_expr"
	K_LITERAL                 = "literal"
	K_VAR_PATH                = "var_path"
	K_INT                     = "int"
	K_BOOL                    = "bool"
	K_LIST                    = "list"
	K_CALL                    = "call"
	K_NOT_OP                  = "not_op"
	K_UNARY_OP                = "unary_op"
	K_BINARY_OP               = "binary_op"
	K_COMPARISON_OP           = "comparison_op"
	K_BOOL_OP                 = "bool_op"
	K_STRING                  = "string"
	K_STRING_CONTENT          = "string_content"
	K_BACKSLASH               = "\\"
	K_ESC_BACKSLASH           = "esc_backslash"
	K_ESC_SINGLE_QUOTE        = "esc_single_quote"
	K_ESC_DOUBLE_QUOTE        = "esc_double_quote"
	K_ESC_BACKTICK            = "esc_backtick"
	K_ESC_NEWLINE             = "esc_newline"
	K_ESC_TAB                 = "esc_tab"
	K_INTERPOLATION           = "interpolation"
	K_MAP                     = "map"
	K_IDENTIFIER              = "identifier"
	K_COMPOUND_ASSIGN         = "compound_assign"
	K_PLUS_EQUAL              = "+="
	K_MINUS_EQUAL             = "-="
	K_STAR_EQUAL              = "*="
	K_SLASH_EQUAL             = "/="
	K_PERCENT_EQUAL           = "%="
	K_PLUS                    = "+"
	K_MINUS                   = "-"
	K_NOT                     = "not"
	K_IF_STMT                 = "if_stmt"
	K_BREAK_STMT              = "break_stmt"
	K_CONTINUE_STMT           = "continue_stmt"
	K_FOR_LOOP                = "for_loop"
	K_DEFER_BLOCK             = "defer_block"
	K_ERRDEFER_BLOCK          = "errdefer_block"
	K_SLICE                   = "slice"
	K_SHELL_STMT              = "shell_stmt"
	K_CRITICAL_SHELL_CMD      = "critical_shell_cmd"
	K_FAIL                    = "fail"
	K_DEL_STMT                = "del_stmt"
	K_JSON_PATH               = "json_path"
	K_RAD_BLOCK               = "rad_block"
	K_RAD_FIELD_STMT          = "rad_field_stmt"
	K_RAD_SORT_STMT           = "rad_sort_stmt"
	K_RAD_SORT_SPECIFIER      = "rad_sort_specifier"
	K_ASC                     = "asc"
	K_DESC                    = "desc"
	K_RAD_FIELD_MODIFIER_STMT = "rad_field_modifier_stmt"
	K_RAD_FIELD_MOD_COLOR     = "rad_field_mod_color"
	K_RAD_FIELD_MOD_MAP       = "rad_field_mod_map"
	K_RAD_IF_STMT             = "rad_if_stmt"

	// Field names
	F_LEFT       = "left"
	F_LEFTS      = "lefts"
	F_RIGHT      = "right"
	F_ROOT       = "root"
	F_INDEXING   = "indexing"
	F_INDEX      = "index"
	F_LIST_ENTRY = "list_entry"
	F_FUNC       = "func"
	F_ARG        = "arg"
	F_OP         = "op"
	F_CONTENTS   = "contents"
	F_EXPR       = "expr"
	F_FORMAT     = "format"
	F_ALIGNMENT  = "alignment"
	F_PADDING    = "padding"
	F_PRECISION  = "precision"
	F_MAP_ENTRY  = "map_entry"
	F_KEY        = "key"
	F_VALUE      = "value"
	F_ALT        = "alt"
	F_CONDITION  = "condition"
	F_STMT       = "stmt"
	F_KEYWORD    = "keyword"
	F_START      = "start"
	F_END        = "end"
	F_SHELL_CMD  = "shell_cmd"
	F_QUIET_MOD  = "quiet_mod"
	F_COMMAND    = "command"
	F_RESPONSE   = "response"
	F_SEGMENT    = "segment"
	F_SOURCE     = "source"
	F_RAD_TYPE   = "rad_type"
	F_IDENTIFIER = "identifier"
	F_SPECIFIER  = "specifier"
	F_DIRECTION  = "direction"
	F_MOD_STMT   = "mod_stmt"
	F_COLOR      = "color"
	F_REGEX      = "regex"
	F_LAMBDA     = "lambda"
)
