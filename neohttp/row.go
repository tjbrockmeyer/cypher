package neohttp

import (
	"bytes"
	"encoding/json"
)

type row struct {
	columns map[string]int
	Row     []interface{} `json:"row"`
	Meta    []interface{} `json:"meta"`
}

func (r *row) GetAt(i int) interface{} {
	return r.Row[i]
}

func (r *row) Get(n string) interface{} {
	return r.Row[r.columns[n]]
}

func (r *row) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	i := 0
	for name, index := range r.columns {
		buf.WriteString("\"" + name + "\":")
		b, _ := json.Marshal(r.Row[index])
		buf.Write(b)
		if i < len(r.columns)-1 {
			buf.WriteString(",")
			i++
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
