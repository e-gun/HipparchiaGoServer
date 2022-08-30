package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type RedisLogin struct {
	Addr     string
	Password string
	DB       int
}

func decoderedislogin(redislogininfo []byte) RedisLogin {
	var rl RedisLogin
	err := json.Unmarshal(redislogininfo, &rl)
	if err != nil {
		fmt.Println(fmt.Sprintf("CANNOT PARSE YOUR REDIS LOGIN CREDENTIALS AS JSON [%s v.%s] ", myname, version))
		panic(err)
	}
	return rl
}

//
// REDIS
//

func grabredisconnection() redis.Conn {
	// to test that we really are pooling:
	// if you uncomment the two Printfs you will see one and only one "called" vs multiple "grabbed"
	if RedisPool == nil {
		poolinit()
		// fmt.Printf("poolinit() called\n")
	}
	connection := RedisPool.Get()
	// fmt.Printf("connection grabbed\n")
	return connection
}

func poolinit() {
	RedisPool = newPool(cfg.RLogin.Addr)
	cleanupHook()
}

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		// Dial or DialContext must be set. When both are set, DialContext takes precedence over Dial.
		Dial: func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

func cleanupHook() {
	c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		e := RedisPool.Close()
		checkerror(e)
		os.Exit(0)
	}()
}

func rcdel(c redis.Conn, k string) {
	_, err := c.Do("DEL", k)
	checkerror(err)
}

func rcsetint(c redis.Conn, k string, v int64) {
	_, err := c.Do("SET", k, v)
	checkerror(err)
}

func rcsetstr(c redis.Conn, k string, v string) {
	_, err := c.Do("SET", k, v)
	checkerror(err)
}

func rcpopstr(c redis.Conn, k string) string {
	s, err := redis.String(c.Do("SPOP", k))
	if err != nil {
		s = "SET_IS_EMPTY"
	}
	return s
}
