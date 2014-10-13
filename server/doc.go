// Package server supplies a way to use ledis as service.
// Server implements the redis protocol called RESP (REdis Serialization Protocol).
// For more information, please see http://redis.io/topics/protocol.
//
// You can use ledis with many available redis clients directly, for example, redis-cli.
// But I also supply some ledis client at client folder, and have been adding more for other languages.
//
// Usage
//
// Start a ledis server is very simple:
//
//  cfg := config.NewConfigDefault()
//  cfg.Addr = "127.0.0.1:6380"
//  cfg.DataDir = "/tmp/ledis"
//  app := server.NewApp(cfg)
//  app.Run()
//
// Replication
//
// You can start a slave ledis server for replication, open slave is simple too, you can set slaveof in config or run slaveof command in shell.
//
// For example, if you start a slave server, and the master server's address is 127.0.0.1:6380, you can start replication in shell:
//
//  ledis-cli -p 6381
//  ledis 127.0.0.1:6381 > slaveof 127.0.0.1 6380
//
// After you send slaveof command, the slave will start to sync master's write ahead log and replicate from it.
// You must notice that use_replication must be set true if you want to use it.
//
// HTTP Interface
//
// LedisDB provides http interfaces for most commands(except the replication commands)
//
//  curl http://127.0.0.1:11181/SET/hello/world
//  → {"SET":[true,"OK"]}
//
//  curl http://127.0.0.1:11181/0/GET/hello?type=json
//  → {"GET":"world"}
//
package server
