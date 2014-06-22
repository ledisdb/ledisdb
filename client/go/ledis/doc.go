// Package ledis is a client for the ledisdb.
//
// Config
//
// Config struct contains configuration for ledisdb:
//
//     Addr            ledisdb server address, like 127.0.0.1:6380
//     MaxIdleConns    max idle connections for ledisdb
//
// Client
//
// The client is the primary interface for ledisdb. You must first create a client with proper config for working.
//
//     cfg := new(Config)
//     cfg.Addr = "127.0.0.1:6380"
//     cfg.MaxIdleConns = 4
//
//     c := NewClient(cfg)
//
// The most important function for client is Do function to send commands to remote server.
//
//     reply, err := c.Do("ping")
//
//     reply, err := c.Do("set", "key", "value")
//
//     reply, err := c.Do("get", "key")
//
// Connection
//
// You can use an independent connection to send commands.
//
//     //get a connection
//     conn := c.Get()
//
//     //connection send command
//     conn.Do("ping")
//
// Reply Helper
//
// You can use reply helper to convert a reply to a specific type.
//
//     exists, err := ledis.Bool(c.Do("exists", "key"))
//
//     score, err := ledis.Int64(c.Do("zscore", "key", "member"))
package ledis
