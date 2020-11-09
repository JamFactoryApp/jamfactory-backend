package redis

import "strings"

type Key struct {
	Keys []string
}

func NewKey(root string) Key {
	return Key{Keys: []string{root}}
}

func (k Key) Append(key string) Key {
	return Key{Keys: append(k.Keys, key)}
}

func (k Key) AppendKey(key Key) Key {
	return Key{Keys: append(k.Keys, key.Keys...)}
}

func (k Key) String() string {
	return strings.Join(k.Keys, ":")
}
