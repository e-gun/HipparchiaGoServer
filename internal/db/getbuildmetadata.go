package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"strings"
)

func LoadBuildMetadata() string {
	const (
		SELECTFROM = `SELECT category, builderver, gitcommit, date, notes FROM buildmetadata`
		TMPL       = "%s\t%s\t%s\t%s\t%s"
	)
	dbconn := getdbconnection()
	defer dbconn.Release()

	type tmpstruct struct {
		Category, Builderver, Gitcommit, Date, Notes string
	}

	var alldata []tmpstruct
	var TS tmpstruct

	foreach := []any{
		&TS.Category,
		&TS.Builderver,
		&TS.Gitcommit,
		&TS.Date,
		&TS.Notes}

	rwfnc := func() error {
		alldata = append(alldata, TS)
		return nil
	}

	foundrows, err := dbconn.Query(context.Background(), SELECTFROM)
	Msg.EC(err)
	_, e := pgx.ForEachRow(foundrows, foreach, rwfnc)
	if e != nil {
		fmt.Println(e)
	}

	var allstr []string
	for _, v := range alldata {
		allstr = append(allstr, fmt.Sprintf(TMPL, v.Date, v.Builderver, v.Gitcommit, v.Category, v.Notes))
	}

	return strings.Join(allstr, "\n")
}
