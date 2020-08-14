package utils

import "strings"

type RedisKey struct {
	Keys []string
}


func (rediskey RedisKey) Append (key string) RedisKey {
	return RedisKey{Keys: append(rediskey.Keys, key)}
}

func (rediskey RedisKey) AppendKey (key RedisKey) RedisKey {
	return RedisKey{Keys: append(rediskey.Keys, key.Keys...)}
}

func (rediskey RedisKey) String () string {
	return strings.Join(rediskey.Keys, ":")
}

