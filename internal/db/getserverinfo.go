package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func LoadServerInfo() string {
	const (
		SELECTFROM = `SELECT version()`
		NOTFOUND   = `(psql server version was not returned)`
	)

	var alldata []string // should be impossible for more than one row to come back
	var version string

	foreach := []any{
		&version}

	rwfnc := func() error {
		alldata = append(alldata, version)
		return nil
	}

	dbconn := getdbconnection()
	defer dbconn.Release()

	foundrows, err := dbconn.Query(context.Background(), SELECTFROM)
	Msg.EC(err)

	_, e := pgx.ForEachRow(foundrows, foreach, rwfnc)
	if e != nil {
		fmt.Println(e)
	}

	if len(alldata) > 0 {
		return alldata[0]
	}

	return NOTFOUND
}
