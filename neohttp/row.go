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
	buf.WriteString("{")
	for name, index := range r.columns {
		buf.WriteString("\"" + name + "\":")
		b, _ := json.Marshal(r.Row[index])
		buf.Write(b)
		if index != len(r.columns)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("}")
	return buf.Bytes(), nil
}
