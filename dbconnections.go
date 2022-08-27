package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4/pgxpool"
)

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
// GENERAL AUTHENTICATION
//

func decodepsqllogin(psqllogininfo []byte) PostgresLogin {
	var ps PostgresLogin
	err := json.Unmarshal(psqllogininfo, &ps)
	if err != nil {
		fmt.Println(fmt.Sprintf("CANNOT PARSE YOUR POSTGRES LOGIN CREDENTIALS AS JSON [%s v.%s] ", MYNAME, VERSION))
		panic(err)
	}
	return ps
}

//
// POSTGRESQL
//

func grabpgsqlconnection() *pgxpool.Pool {
	var pl PostgresLogin

	if cfg.PGLogin.DBName == "" {
		makeconfig()
	}

	pl = cfg.PGLogin

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

func blankconfig() CurrentConfiguration {
	// need a non-commandline config
	var thecfg CurrentConfiguration
	thecfg.PGLogin.Port = PSDefaultPort
	// cfg.PGLogin.Pass = cfg.PSQP
	thecfg.PGLogin.Pass = ""
	thecfg.PGLogin.User = PSDefaultUser
	thecfg.PGLogin.DBName = PSDefaultDB
	thecfg.PGLogin.Host = PSDefaultHost
	return thecfg
}
