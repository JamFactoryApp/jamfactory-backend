package store

type Store[T any] interface {
	Get(key string) (*T, error)
	GetAll() ([]*T, error)
	Save(obj *T, key string) error
	Delete(key string) error
}
