package store

type Config struct {
	Name string `json:"name"`

	Path string `json:"path"`

	//for leveldb, goleveldb
	Compression     bool `json:"compression"`
	BlockSize       int  `json:"block_size"`
	WriteBufferSize int  `json:"write_buffer_size"`
	CacheSize       int  `json:"cache_size"`
	MaxOpenFiles    int  `json:"max_open_files"`

	//for lmdb
	MapSize int `json:"map_size"`
}
