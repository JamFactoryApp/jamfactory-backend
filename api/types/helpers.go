package types

import "encoding/json"

type JSONString struct {
	Value string
	Valid bool
	Set   bool
}

func (i *JSONString) UnmarshalJSON(data []byte) error {

	i.Set = true

	if string(data) == "null" {
		i.Valid = false
		return nil
	}

	var temp string
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	i.Value = temp
	i.Valid = true
	return nil
}

type JSONInt struct {
	Value int
	Valid bool
	Set   bool
}

func (i *JSONInt) UnmarshalJSON(data []byte) error {

	i.Set = true

	if string(data) == "null" {
		i.Valid = false
		return nil
	}

	var temp int
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	i.Value = temp
	i.Valid = true
	return nil
}

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
