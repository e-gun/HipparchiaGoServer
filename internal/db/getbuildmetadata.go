package db

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/olekukonko/tablewriter" // tablewriter is a 'free' import because it was already a dependency via wego
)

func LoadBuildMetadata() string {
	const (
		SELECTFROM = `SELECT category, builderver, gitcommit, date, notes FROM buildmetadata ORDER BY category ASC`
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

	var allstr [][]string
	for _, v := range alldata {
		bv := fmt.Sprintf("%s (git: %s)", v.Builderver, v.Gitcommit)
		allstr = append(allstr, []string{v.Category, bv, v.Date, v.Notes})
	}

	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf)
	table.Header("Category", "HGB version", "Build Date", "Notes")
	table.Bulk(allstr)
	table.Render()

	return buf.String()
}
