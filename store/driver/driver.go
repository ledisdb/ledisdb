package driver

type IDB interface {
	Close() error

	Get(key []byte) ([]byte, error)

	Put(key []byte, value []byte) error
	Delete(key []byte) error

	NewIterator() IIterator

	NewWriteBatch() IWriteBatch
}

type IIterator interface {
	Close() error

	First()
	Last()
	Seek(key []byte)

	Next()
	Prev()

	Valid() bool

	Key() []byte
	Value() []byte
}

type IWriteBatch interface {
	Close() error

	Put(key []byte, value []byte)
	Delete(key []byte)
	Commit() error
	Rollback() error
}
