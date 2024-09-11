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
	case int:
		return NewRuntimeInt(val.(int))
	case []int:
		return NewRuntimeIntArray(val.([]int))
	case float64:
		return NewRuntimeFloat(val.(float64))
	case []float64:
		return NewRuntimeFloatArray(val.([]float64))
	case bool:
		return NewRuntimeBool(val.(bool))
	default:
		// todo via printer
		panic("unknown type")
	}
}

func NewRuntimeString(val string) RuntimeLiteral {
	return RuntimeLiteral{Type: RslString, value: val}
}

func NewRuntimeStringArray(val []string) RuntimeLiteral {
	return RuntimeLiteral{Type: RslStringArray, value: val}
}

func NewRuntimeInt(val int) RuntimeLiteral {
	return RuntimeLiteral{Type: RslInt, value: val}
}

func NewRuntimeIntArray(val []int) RuntimeLiteral {
	return RuntimeLiteral{Type: RslIntArray, value: val}
}

func NewRuntimeFloat(val float64) RuntimeLiteral {
	return RuntimeLiteral{Type: RslFloat, value: val}
}

func NewRuntimeFloatArray(val []float64) RuntimeLiteral {
	return RuntimeLiteral{Type: RslFloatArray, value: val}
}

func NewRuntimeBool(val bool) RuntimeLiteral {
	return RuntimeLiteral{Type: RslBool, value: val}
}

func (l RuntimeLiteral) GetString() string {
	return l.value.(string)
}

func (l RuntimeLiteral) GetStringArray() []string {
	return l.value.([]string)
}

func (l RuntimeLiteral) GetInt() int {
	return l.value.(int)
}

func (l RuntimeLiteral) GetIntArray() []int {
	return l.value.([]int)
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

type JsonFieldVar struct {
	Name Token
	Path JsonPath
	env  *Env
}

func (j *JsonFieldVar) AddMatch(match string) {
	existing := j.env.GetByToken(j.Name, RslStringArray).value.([]string)
	existing = append(existing, match)
	j.env.SetAndImplyType(j.Name, existing)
}
