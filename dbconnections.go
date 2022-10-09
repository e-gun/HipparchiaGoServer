//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresLogin struct {
	Host   string
	Port   int
	User   string
	Pass   string
	DBName string
}

//
// POSTGRESQL
//

// FillPSQLPoool - build the pgxpool that the whole program will Acquire() from
func FillPSQLPoool() *pgxpool.Pool {
	// costs about 1M RAM per connection
	pl := cfg.PGLogin

	// url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?pool_min_conns=%d", pl.User, pl.Pass, pl.Host, pl.Port, pl.DBName, cfg.WorkerCount)
	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", pl.User, pl.Pass, pl.Host, pl.Port, pl.DBName)

	config, oops := pgxpool.ParseConfig(url)
	if oops != nil {
		msg(fmt.Sprintf("Could not execute pgxpool.ParseConfig(url) via %s", url), -1)
		panic(oops)
	}

	thepool, err := pgxpool.ConnectConfig(context.Background(), config)

	if err != nil {
		msg(fmt.Sprintf("Could not connect to PostgreSQL via %s", url), -1)
		panic(err)
	}

	return thepool
}

// GetPSQLconnection - Acquire() a connection from the main pgxpool
func GetPSQLconnection() *pgxpool.Conn {
	c, _ := dbpool.Acquire(context.Background())
	return c
}
