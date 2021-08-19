package db

import "MyRedis/interface/redis"

// DB is the interface for redis style storage engine
type DB interface {
	Exec(clent redis.Connection, args [][]byte) redis.Reply
	AfterClientClose(c redis.Connection)
	Close()
}
