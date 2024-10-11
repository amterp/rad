package core

type JsonFieldVar struct {
	Name    Token
	Path    JsonPath
	IsArray bool
	env     *Env
}

func (j *JsonFieldVar) AddMatch(match interface{}) {
	jsonFieldVar := j.env.GetJsonField(j.Name)
	if jsonFieldVar.IsArray {
		existing := j.env.GetByToken(j.Name, RslArrayT).([]interface{})
		existing = append(existing, match)
		j.env.SetAndImplyType(j.Name, existing)
	} else {
		j.env.SetAndImplyType(j.Name, match)
	}
}
