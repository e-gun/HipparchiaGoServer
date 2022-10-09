//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresLogin struct {
	Host   string
	Port   int
	User   string
	Pass   string
	DBName string
}

// FillPSQLPoool - build the pgxpool that the whole program will Acquire() from
func FillPSQLPoool() *pgxpool.Pool {
	// costs about 1M RAM per connection
	// it is not clear that the casual user gains much from a pool; this mechanism mattered more for python

	// if min < WorkerCount the search will be slowed significantly
	// and remember that idle connections close, so you can have 20 workers fighting for one connection: very bad news

	// max should cap a networked server's resource allocation to the equivalent of N simultaneous users

	min := cfg.WorkerCount
	max := SIMULTANEOUSSEARCHES * cfg.WorkerCount

	pl := cfg.PGLogin
	u := "postgres://%s:%s@%s:%d/%s?pool_min_conns=%d&pool_max_conns=%d"
	url := fmt.Sprintf(u, pl.User, pl.Pass, pl.Host, pl.Port, pl.DBName, min, max)

	config, e := pgxpool.ParseConfig(url)
	if e != nil {
		msg(fmt.Sprintf("Could not execute pgxpool.ParseConfig(url) via %s", url), -1)
		panic(e)
	}

	thepool, e := pgxpool.ConnectConfig(context.Background(), config)
	if e != nil {
		msg(fmt.Sprintf("Could not connect to PostgreSQL via %s", url), -1)
		panic(e)
	}
	return thepool
}

// GetPSQLSimpleConnection - return a *pgx.Conn to the database
func GetPSQLSimpleConnection() *pgx.Conn {
	// this cannot be used ATM; worklinequery() needs a pgxpool.Conn
	// there is no simple way to toggle this while running
	// but when debugging it is possible hand-tweak that one function + sed -i "" "s/dbconn.Release/dbconn.Close" *go

	pl := cfg.PGLogin
	u := "postgres://%s:%s@%s:%d/%s"
	url := fmt.Sprintf(u, pl.User, pl.Pass, pl.Host, pl.Port, pl.DBName)
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		msg(fmt.Sprintf("PSQLSimpleConn() could not connect via %s", url), -1)
		panic(err)
	}
	return conn
}

// GetPSQLPooledConnection - Acquire() a connection from the main pgxpool
func GetPSQLPooledConnection() *pgxpool.Conn {
	dbc, err := psqlpool.Acquire(context.Background())
	if err != nil {
		msg(fmt.Sprintf("GetPSQLconnection() could not .Acquire() from psqlpool"), -1)
		panic(err)
	}
	return dbc
}

// GetPSQLconnection - Acquire() a connection from the main pgxpool
func GetPSQLconnection() *pgxpool.Conn {
	// alternate is:
	// return GetPSQLSimpleConnection()
	return GetPSQLPooledConnection()
}
