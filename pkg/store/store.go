package store

import (
	"bytes"
	"encoding/gob"
	"errors"
)

var (
	ErrObjNotFound      = errors.New("store: obj not found")
	ErrInterfaceConvert = errors.New("store: Failed to convert user from interface{} to []bytes")
)

type Store[T any] interface {
	Get(key string) (*T, error)
	GetAll() ([]*T, error)
	Save(obj *T, key string) error
	Delete(key string) error
}

type Set interface {
	Add(obj string) error
	GetAll() ([]string, error)
	Has(obj string) (bool, error)
	Delete(obj string) error
}

func serialize[T any](obj *T) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(obj)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func deserialize[T any](data []byte, obj *T) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(obj)
}
