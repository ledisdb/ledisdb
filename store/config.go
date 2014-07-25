package store

type Config struct {
	Name string

	Path string

	//for leveldb, goleveldb
	Compression     bool
	BlockSize       int
	WriteBufferSize int
	CacheSize       int
	MaxOpenFiles    int

	//for lmdb
	MapSize int
}
