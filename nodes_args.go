package rts

import (
	"fmt"
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
	RangeConstraints map[string]*ArgRangeConstraint
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

func (ad *ArgDecl) ShorthandStr() *string {
	if ad.Shorthand != nil {
		return &ad.Shorthand.Shorthand
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

type ArgRangeConstraint struct {
	BaseNode
	ArgName ArgConstraintArgName
	Range   ArgRangeValue
}

type ArgRangeValue struct {
	Opener ArgRangeOpenerCloser
	Closer ArgRangeOpenerCloser
	Min    *ArgRangeMinMax
	Max    *ArgRangeMinMax
}

type ArgRangeOpenerCloser struct {
	BaseNode
	Token string
}

type ArgRangeMinMax struct {
	BaseNode
	Value float64
}

func newArgBlock(src string, node *ts.Node) (*ArgBlock, bool) {
	return &ArgBlock{
		BaseNode:         newBaseNode(src, node),
		Args:             findArgDeclarations(src, node),
		EnumConstraints:  findArgEnumConstraints(src, node),
		RegexConstraints: findArgRegexConstraints(src, node),
		RangeConstraints: findArgRangeConstraints(src, node),
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
				ExternalName: extractString(src, renameNode),
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
				asInt := extractArgInt(src, defaultNode)
				argDefault = &ArgDeclDefault{
					BaseNode:   newBaseNode(src, defaultNode),
					DefaultInt: &asInt,
				}
			} else if typeStr == "float" {
				asFloat := extractArgFloat(src, defaultNode)
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
				stringList := extractStringList(src, defaultNode)
				argDefault = &ArgDeclDefault{
					BaseNode:          newBaseNode(src, defaultNode),
					DefaultStringList: &stringList,
				}
			} else if typeStr == "int[]" {
				intList := extractIntList(src, defaultNode)
				argDefault = &ArgDeclDefault{
					BaseNode:       newBaseNode(src, defaultNode),
					DefaultIntList: &intList,
				}
			} else if typeStr == "float[]" {
				floatList := extractFloatList(src, defaultNode)
				argDefault = &ArgDeclDefault{
					BaseNode:         newBaseNode(src, defaultNode),
					DefaultFloatList: &floatList,
				}
			} else if typeStr == "bool[]" {
				boolList := extractBoolList(src, defaultNode)
				argDefault = &ArgDeclDefault{
					BaseNode:        newBaseNode(src, defaultNode),
					DefaultBoolList: &boolList,
				}
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
			BaseNode: newBaseNode(src, &decl),
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

func findArgRangeConstraints(src string, node *ts.Node) map[string]*ArgRangeConstraint {
	rangeConstraints := node.ChildrenByFieldName("range_constraint", node.Walk())
	constraints := make(map[string]*ArgRangeConstraint)
	if len(rangeConstraints) == 0 {
		return constraints
	}

	for _, constraint := range rangeConstraints {
		nameNode := constraint.ChildByFieldName("arg_name")

		name := src[nameNode.StartByte():nameNode.EndByte()]

		openerNode := constraint.ChildByFieldName("opener")
		closerNode := constraint.ChildByFieldName("closer")
		minNode := constraint.ChildByFieldName("min")
		maxNode := constraint.ChildByFieldName("max")

		opener := src[openerNode.StartByte():openerNode.EndByte()]
		closer := src[closerNode.StartByte():closerNode.EndByte()]

		var rMin *ArgRangeMinMax
		if minNode != nil {
			minValue, _ := strconv.ParseFloat(src[minNode.StartByte():minNode.EndByte()], 64)
			rMin = &ArgRangeMinMax{
				BaseNode: newBaseNode(src, minNode),
				Value:    minValue,
			}
		}

		var rMax *ArgRangeMinMax
		if maxNode != nil {
			maxValue, _ := strconv.ParseFloat(src[maxNode.StartByte():maxNode.EndByte()], 64)
			rMax = &ArgRangeMinMax{
				BaseNode: newBaseNode(src, maxNode),
				Value:    maxValue,
			}
		}

		constraints[name] = &ArgRangeConstraint{
			BaseNode: newBaseNode(src, &constraint),
			ArgName: ArgConstraintArgName{
				BaseNode: newBaseNode(src, nameNode),
				Name:     name,
			},
			Range: ArgRangeValue{
				Opener: ArgRangeOpenerCloser{
					BaseNode: newBaseNode(src, openerNode),
					Token:    opener,
				},
				Closer: ArgRangeOpenerCloser{
					BaseNode: newBaseNode(src, closerNode),
					Token:    closer,
				},
				Min: rMin,
				Max: rMax,
			},
		}
	}

	return constraints
}

func extractArgInt(src string, defaultNode *ts.Node) int64 {
	multiplier := int64(1)
	ops := defaultNode.ChildrenByFieldName("op", defaultNode.Walk())
	for _, op := range ops {
		opSrc := src[op.StartByte():op.EndByte()]
		if opSrc == "-" {
			multiplier *= -1
		}
	}
	valueNode := defaultNode.ChildByFieldName("value")
	valueStr := src[valueNode.StartByte():valueNode.EndByte()]
	value, _ := strconv.ParseInt(valueStr, 10, 64)
	return value * multiplier
}

func extractArgFloat(src string, defaultNode *ts.Node) float64 {
	multiplier := 1.0
	ops := defaultNode.ChildrenByFieldName("op", defaultNode.Walk())
	for _, op := range ops {
		opSrc := src[op.StartByte():op.EndByte()]
		if opSrc == "-" {
			multiplier *= -1
		}
	}
	valueNode := defaultNode.ChildByFieldName("value")
	valueStr := src[valueNode.StartByte():valueNode.EndByte()]
	value, _ := strconv.ParseFloat(valueStr, 64)
	return value * multiplier
}

func extractStringList(src string, valuesNode *ts.Node) []string {
	valuesStringNodes := valuesNode.ChildrenByFieldName("list_entry", valuesNode.Walk())
	var values []string
	for _, stringNode := range valuesStringNodes {
		values = append(values, extractString(src, &stringNode))
	}
	return values
}

func extractIntList(src string, intListNode *ts.Node) []int64 {
	intArgNodes := intListNode.ChildrenByFieldName("list_entry", intListNode.Walk())
	var values []int64
	for _, intArgNode := range intArgNodes {
		values = append(values, extractArgInt(src, &intArgNode))
	}
	return values
}

func extractFloatList(src string, floatListNode *ts.Node) []float64 {
	floatArgNodes := floatListNode.ChildrenByFieldName("list_entry", floatListNode.Walk())
	var values []float64
	for _, floatArgNode := range floatArgNodes {
		values = append(values, extractArgFloat(src, &floatArgNode))
	}
	return values
}

func extractBoolList(src string, boolListNode *ts.Node) []bool {
	boolArgNodes := boolListNode.ChildrenByFieldName("list_entry", boolListNode.Walk())
	var values []bool
	for _, boolArgNode := range boolArgNodes {
		asBool, _ := strconv.ParseBool(src[boolArgNode.StartByte():boolArgNode.EndByte()])
		values = append(values, asBool)
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
			panic(fmt.Sprintf("unexpected child field name [%d, %d]: %q",
				content.StartPosition().Row+1, content.StartPosition().Column+1, childFieldName))
		}
	}

	return sb.String()
}
