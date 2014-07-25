package store

type Config struct {
	Name string

	Path string

	//for leveldb, goleveldb, rocksdb
	Compression     bool
	BlockSize       int
	WriteBufferSize int
	CacheSize       int
	MaxOpenFiles    int
}
