package cache

// SourceFunc is a function that can be called if data has not yet been cached in redis
type SourceFunc func(string) (interface{}, error)
