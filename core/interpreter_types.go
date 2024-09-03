package core

type RuntimeLiteral struct {
	Type  RslTypeEnum
	value interface{} // not a pointer, e.g. just 'string'
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
	existing := j.env.Get(j.Name, RslStringArray).value.([]string)
	existing = append(existing, match)
	j.env.Set(j.Name, existing)
}
