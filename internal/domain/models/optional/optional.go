package optional

import (
	"encoding/json"
)

type Optional[T any] struct {
	Defined bool
	Value   *T
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Defined = true
	return json.Unmarshal(data, &o.Value)
}
