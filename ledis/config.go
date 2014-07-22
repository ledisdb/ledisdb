package ledis

type Config struct {
	DataDir string `json:"data_dir"`

	DB struct {
		Compression     bool `json:"compression"`
		BlockSize       int  `json:"block_size"`
		WriteBufferSize int  `json:"write_buffer_size"`
		CacheSize       int  `json:"cache_size"`
		MaxOpenFiles    int  `json:"max_open_files"`
	} `json:"db"`

	BinLog struct {
		Use         bool `json:"use"`
		MaxFileSize int  `json:"max_file_size"`
		MaxFileNum  int  `json:"max_file_num"`
	} `json:"binlog"`
}
