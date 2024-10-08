package core

type RuntimeLiteral struct {
	Type  RslTypeEnum
	value interface{} // not a pointer, e.g. just 'string'
}

func NewRuntimeLiteral(val interface{}) RuntimeLiteral {
	switch val.(type) {
	case string:
		return NewRuntimeString(val.(string))
	case []string:
		return NewRuntimeStringArray(val.([]string))
	case int64:
		return NewRuntimeInt(val.(int64))
	case []int64:
		return NewRuntimeIntArray(val.([]int64))
	case float64:
		return NewRuntimeFloat(val.(float64))
	case []float64:
		return NewRuntimeFloatArray(val.([]float64))
	case bool:
		return NewRuntimeBool(val.(bool))
	case []bool:
		return NewRuntimeBoolArray(val.([]bool))
	case []interface{}:
		return NewRuntimeMixedArray(val.([]interface{}))
	default:
		// todo via printer
		panic("unknown type")
	}
}

func NewRuntimeString(val string) RuntimeLiteral {
	return RuntimeLiteral{Type: RslStringT, value: val}
}

func NewRuntimeStringArray(val []string) RuntimeLiteral {
	return RuntimeLiteral{Type: RslStringArrayT, value: val}
}

func NewRuntimeInt(val int64) RuntimeLiteral {
	return RuntimeLiteral{Type: RslIntT, value: val}
}

func NewRuntimeIntArray(val []int64) RuntimeLiteral {
	return RuntimeLiteral{Type: RslIntArrayT, value: val}
}

func NewRuntimeFloat(val float64) RuntimeLiteral {
	return RuntimeLiteral{Type: RslFloatT, value: val}
}

func NewRuntimeFloatArray(val []float64) RuntimeLiteral {
	return RuntimeLiteral{Type: RslFloatArrayT, value: val}
}

func NewRuntimeBool(val bool) RuntimeLiteral {
	return RuntimeLiteral{Type: RslBoolT, value: val}
}

func NewRuntimeBoolArray(val []bool) RuntimeLiteral {
	return RuntimeLiteral{Type: RslBoolArrayT, value: val}
}

func NewRuntimeMixedArray(val []interface{}) RuntimeLiteral {
	return RuntimeLiteral{Type: RslArrayT, value: val}
}

func (l RuntimeLiteral) GetString() string {
	return l.value.(string)
}

func (l RuntimeLiteral) GetStringArray() []string {
	return l.value.([]string)
}

func (l RuntimeLiteral) GetInt() int64 {
	return l.value.(int64)
}

func (l RuntimeLiteral) GetIntArray() []int64 {
	return l.value.([]int64)
}

func (l RuntimeLiteral) GetFloat() float64 {
	return l.value.(float64)
}

func (l RuntimeLiteral) GetFloatArray() []float64 {
	return l.value.([]float64)
}

func (l RuntimeLiteral) GetBool() bool {
	return l.value.(bool)
}

func (l RuntimeLiteral) GetBoolArray() []bool {
	return l.value.([]bool)
}

func (l RuntimeLiteral) GetMixedArray() []interface{} {
	return l.value.([]interface{})
}

type JsonFieldVar struct {
	Name    Token
	Path    JsonPath
	IsArray bool
	env     *Env
}

func (j *JsonFieldVar) AddMatch(match interface{}) {
	jsonFieldVar := j.env.GetJsonField(j.Name)
	if jsonFieldVar.IsArray {
		existing := j.env.GetByToken(j.Name, RslArrayT).value.([]interface{})
		existing = append(existing, match)
		j.env.SetAndImplyType(j.Name, existing)
	} else {
		j.env.SetAndImplyType(j.Name, match)
	}
}
