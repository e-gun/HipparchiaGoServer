package main

import (
	"context"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4/pgxpool"
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

type PostgresLogin struct {
	Host   string
	Port   int
	User   string
	Pass   string
	DBName string
}

var (
	RedisPool *redis.Pool
)

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
	signal.Notify(c, os.Interrupt)
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

//
// POSTGRESQL
//

func grabpgsqlconnection() *pgxpool.Pool {
	var pl PostgresLogin

	if cfg.PGLogin.DBName == "" {
		// this will probably need refactoring later: non-intuitive to "configure" here
		// grabpgsqlconnection() will be called before main() and so before configatstartup()
		// workmapper() in initialization.go does this
		// avoid: "flag redefined: psqp"
		//flag.StringVar(&cfg.PSQP, "psqp", "", "[testing] PSQL Password")
		configatstartup()
		cfg.PGLogin.Port = PSDefaultPort
		cfg.PGLogin.Pass = cfg.PSQP
		cfg.PGLogin.User = PSDefaultUser
		cfg.PGLogin.DBName = PSDefaultDB
		cfg.PGLogin.Host = PSDefaultHost
		pl = cfg.PGLogin
	} else {
		pl = cfg.PGLogin
	}

	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", pl.User, pl.Pass, pl.Host, pl.Port, pl.DBName)

	config, oops := pgxpool.ParseConfig(url)
	if oops != nil {
		msg(fmt.Sprintf("Could not execute pgxpool.ParseConfig(url) via %s", url), -1)
		panic(oops)
	}

	pooledconnection, err := pgxpool.ConnectConfig(context.Background(), config)

	if err != nil {
		msg(fmt.Sprintf("Could not connect to PostgreSQL via %s", url), -1)
		panic(err)
	}

	msg(fmt.Sprintf("Connected to %s on PostgreSQL", pl.DBName), 4)

	return pooledconnection
}
