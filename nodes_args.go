package rts

import (
	"strconv"
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type ArgBlock struct {
	BaseNode
	Args []ArgDecl
	// ArgName -> Constraint
	EnumConstraints  map[string]*ArgEnumConstraint
	RegexConstraints map[string]*ArgRegexConstraint
}

type ArgDecl struct {
	BaseNode
	Name      ArgDeclName
	Type      ArgDeclType
	Rename    *ArgDeclRename
	Shorthand *ArgDeclShorthand
	Default   *ArgDeclDefault
	Comment   *ArgDeclComment
}

func (ad *ArgDecl) ExternalName() string {
	if ad.Rename != nil {
		return ad.Rename.ExternalName
	}
	return ad.Name.Name
}

func (ad *ArgDecl) CommentStr() *string {
	if ad.Comment != nil {
		return &ad.Comment.Comment
	}
	return nil
}

type ArgDeclName struct {
	BaseNode
	Name string
}

type ArgDeclType struct {
	BaseNode
	Type string // todo or enum?
}

type ArgDeclRename struct {
	BaseNode
	ExternalName string
}

type ArgDeclShorthand struct {
	BaseNode
	Shorthand string // single char
}

type ArgDeclDefault struct {
	BaseNode
	DefaultString     *string
	DefaultInt        *int64
	DefaultFloat      *float64
	DefaultBool       *bool
	DefaultStringList *[]string
	DefaultIntList    *[]int64
	DefaultFloatList  *[]float64
	DefaultBoolList   *[]bool
}

type ArgDeclComment struct {
	BaseNode
	Comment string
}

type ArgEnumConstraint struct {
	BaseNode
	ArgName ArgConstraintArgName
	Values  ArgEnumValues
}

type ArgConstraintArgName struct {
	BaseNode
	Name string
}

type ArgEnumValues struct {
	BaseNode
	Values []string
}

type ArgRegexConstraint struct {
	BaseNode
	ArgName ArgConstraintArgName
	Regex   ArgRegexValue
}

type ArgRegexValue struct {
	BaseNode
	Value string
}

func newArgBlock(src string, node *ts.Node) (*ArgBlock, bool) {
	declarations := findArgDeclarations(src, node)
	enumConstraints := findArgEnumConstraints(src, node)
	regexConstraints := findArgRegexConstraints(src, node)
	return &ArgBlock{
		BaseNode:         newBaseNode(src, node),
		Args:             declarations,
		EnumConstraints:  enumConstraints,
		RegexConstraints: regexConstraints,
	}, true
}

func findArgDeclarations(src string, node *ts.Node) []ArgDecl {
	decls := node.ChildrenByFieldName("declaration", node.Walk())

	var argDecls []ArgDecl
	for _, decl := range decls {
		nameNode := decl.ChildByFieldName("arg_name")
		renameNode := decl.ChildByFieldName("rename")
		shorthandNode := decl.ChildByFieldName("shorthand")
		typeNode := decl.ChildByFieldName("type")
		defaultNode := decl.ChildByFieldName("default")
		commentNode := decl.ChildByFieldName("comment")

		var argRename *ArgDeclRename
		if renameNode != nil {
			argRename = &ArgDeclRename{
				BaseNode:     newBaseNode(src, renameNode),
				ExternalName: src[renameNode.StartByte():renameNode.EndByte()],
			}
		}
		var argShorthand *ArgDeclShorthand
		if shorthandNode != nil {
			argShorthand = &ArgDeclShorthand{
				BaseNode:  newBaseNode(src, shorthandNode),
				Shorthand: src[shorthandNode.StartByte():shorthandNode.EndByte()],
			}
		}

		typeStr := src[typeNode.StartByte():typeNode.EndByte()]
		var argDefault *ArgDeclDefault
		if defaultNode != nil {
			defaultStr := src[defaultNode.StartByte():defaultNode.EndByte()]
			if typeStr == "string" {
				contents := extractString(src, defaultNode)
				argDefault = &ArgDeclDefault{
					BaseNode:      newBaseNode(src, defaultNode),
					DefaultString: &contents,
				}
			} else if typeStr == "int" {
				asInt, _ := strconv.ParseInt(defaultStr, 10, 64)
				argDefault = &ArgDeclDefault{
					BaseNode:   newBaseNode(src, defaultNode),
					DefaultInt: &asInt,
				}
			} else if typeStr == "float" {
				asFloat, _ := strconv.ParseFloat(defaultStr, 64)
				argDefault = &ArgDeclDefault{
					BaseNode:     newBaseNode(src, defaultNode),
					DefaultFloat: &asFloat,
				}
			} else if typeStr == "bool" {
				asBool, _ := strconv.ParseBool(defaultStr)
				argDefault = &ArgDeclDefault{
					BaseNode:    newBaseNode(src, defaultNode),
					DefaultBool: &asBool,
				}
			} else if typeStr == "string[]" {
				// todo
			} else if typeStr == "int[]" {
				// todo
			} else if typeStr == "float[]" {
				// todo
			} else if typeStr == "bool[]" {
				// todo
			}
		}

		var argComment *ArgDeclComment
		if commentNode != nil {
			argComment = &ArgDeclComment{
				BaseNode: newBaseNode(src, commentNode),
				Comment:  src[commentNode.StartByte():commentNode.EndByte()],
			}
		}

		argDecls = append(argDecls, ArgDecl{
			BaseNode: newBaseNode(src, node),
			Name: ArgDeclName{
				BaseNode: newBaseNode(src, nameNode),
				Name:     src[nameNode.StartByte():nameNode.EndByte()],
			},
			Type: ArgDeclType{
				BaseNode: newBaseNode(src, typeNode),
				Type:     typeStr,
			},
			Rename:    argRename,
			Shorthand: argShorthand,
			Default:   argDefault,
			Comment:   argComment,
		})
	}
	return argDecls
}

func findArgEnumConstraints(src string, node *ts.Node) map[string]*ArgEnumConstraint {
	constraints := make(map[string]*ArgEnumConstraint)
	enumConstraints := node.ChildrenByFieldName("enum_constraint", node.Walk())
	if len(enumConstraints) == 0 {
		return constraints
	}

	for _, constraint := range enumConstraints {
		nameNode := constraint.ChildByFieldName("arg_name")
		valuesNode := constraint.ChildByFieldName("values")

		name := src[nameNode.StartByte():nameNode.EndByte()]

		values := extractStringList(src, valuesNode)
		constraints[name] = &ArgEnumConstraint{
			BaseNode: newBaseNode(src, &constraint),
			ArgName: ArgConstraintArgName{
				BaseNode: newBaseNode(src, nameNode),
				Name:     name,
			},
			Values: ArgEnumValues{
				BaseNode: newBaseNode(src, valuesNode),
				Values:   values,
			},
		}
	}
	return constraints
}

func findArgRegexConstraints(src string, node *ts.Node) map[string]*ArgRegexConstraint {
	regexConstraints := node.ChildrenByFieldName("regex_constraint", node.Walk())
	constraints := make(map[string]*ArgRegexConstraint)
	if len(regexConstraints) == 0 {
		return constraints
	}

	for _, constraint := range regexConstraints {
		nameNode := constraint.ChildByFieldName("arg_name")
		regexStrNode := constraint.ChildByFieldName("regex")

		name := src[nameNode.StartByte():nameNode.EndByte()]

		regexStr := extractString(src, regexStrNode)
		constraints[name] = &ArgRegexConstraint{
			BaseNode: newBaseNode(src, &constraint),
			ArgName: ArgConstraintArgName{
				BaseNode: newBaseNode(src, nameNode),
				Name:     regexStr,
			},
			Regex: ArgRegexValue{
				BaseNode: newBaseNode(src, regexStrNode),
				Value:    regexStr,
			},
		}
	}

	return constraints
}

func extractStringList(src string, valuesNode *ts.Node) []string {
	valuesStringNodes := valuesNode.ChildrenByFieldName("string", valuesNode.Walk())
	var values []string
	for _, stringNode := range valuesStringNodes {
		values = append(values, extractString(src, &stringNode))
	}
	return values
}

// extractString of a string node. Does not perform interpolation.
func extractString(src string, stringNode *ts.Node) string {
	contents := stringNode.ChildByFieldName("contents")
	if contents == nil {
		return ""
	}

	// todo should handle string interpolation somehow
	var sb strings.Builder
	contentChildren := contents.Children(contents.Walk())
	for i, content := range contentChildren {
		childSrc := src[content.StartByte():content.EndByte()]
		childFieldName := contents.FieldNameForChild(uint32(i))
		switch childFieldName {
		case "content":
			sb.WriteString(childSrc)
		case "single_quote":
			sb.WriteString("'")
		case "double_quote":
			sb.WriteString(`"`)
		case "backtick":
			sb.WriteString("`")
		case "newline":
			sb.WriteString("\n")
		case "tab":
			sb.WriteString("\t")
		case "backslash":
			sb.WriteString("\\")
		default:
			// todo error
			panic("unexpected child field name")
		}
	}

	return sb.String()
}
