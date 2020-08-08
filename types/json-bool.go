package types

import "encoding/json"

type JSONBool struct {
	Value bool
	Valid bool
	Set   bool
}

func (i *JSONBool) UnmarshalJSON(data []byte) error {

	i.Set = true

	if string(data) == "null" {
		i.Valid = false
		return nil
	}

	var temp bool
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	i.Value = temp
	i.Valid = true
	return nil
}