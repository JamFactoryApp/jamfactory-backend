package cache

import (
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
)

// Cache caches requests in redis
type Cache interface {
	// Query cached data or calls the source function to query uncached data
	Query(key pkgredis.Key, index string, source SourceFunc) (interface{}, error)
}

// SourceFunc is a function that can be called if data has not yet been cached in redis
type SourceFunc func(string) (interface{}, error)
