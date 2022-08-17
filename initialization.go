package main

import (
	"context"
	"fmt"
)

// all functions in here should be run in order to prepare the core data

// authormapper - build a map of all authors keyed to the authorUID: map[string]DbAuthor
func authormapper() map[string]DbAuthor {
	dbpool := grabpgsqlconnection()
	qt := "SELECT %s FROM authors ORDER by universalid ASC"
	q := fmt.Sprintf(qt, AUTHORTEMPLATE)

	foundrows, err := dbpool.Query(context.Background(), q)
	checkerror(err)

	var thefinds []DbAuthor

	defer foundrows.Close()
	for foundrows.Next() {
		// fmt.Println(foundrows.Values())
		// this will die if <nil> comes back inside any of the columns: "cannot scan null into *string"
		// the builder should address this: fixing it here is less ideal
		var thehit DbAuthor
		err := foundrows.Scan(&thehit.UID, &thehit.Language, &thehit.IDXname, &thehit.Name, &thehit.Shortname,
			&thehit.Cleaname, &thehit.Genres, &thehit.RecDate, &thehit.ConvDate, &thehit.Location)
		checkerror(err)
		thefinds = append(thefinds, thehit)
	}

	authormap := make(map[string]DbAuthor)
	for _, val := range thefinds {
		authormap[val.UID] = val
	}

	return authormap

}

// workmapper - build a map of all works keyed to the authorUID: map[string]DbWork
func workmapper() map[string]DbWork {
	dbpool := grabpgsqlconnection()
	qt := "SELECT %s FROM works"
	q := fmt.Sprintf(qt, WORKTEMPLATE)

	foundrows, err := dbpool.Query(context.Background(), q)
	checkerror(err)

	var thefinds []DbWork

	defer foundrows.Close()
	for foundrows.Next() {
		// fmt.Println(foundrows.Values())
		// this will die if <nil> comes back inside any of the columns
		var thehit DbWork
		err := foundrows.Scan(&thehit.UID, &thehit.Title, &thehit.Language, &thehit.Pub, &thehit.LL0,
			&thehit.LL1, &thehit.LL2, &thehit.LL3, &thehit.LL4, &thehit.LL5, &thehit.Genre,
			&thehit.Xmit, &thehit.Type, &thehit.Prov, &thehit.RecDate, &thehit.ConvDate, &thehit.WdCount,
			&thehit.FirstLine, &thehit.LastLine, &thehit.Authentic)
		checkerror(err)
		thefinds = append(thefinds, thehit)
	}

	for _, val := range thefinds {
		val.WorkNum = val.FindWorknumber()
	}

	workmap := make(map[string]DbWork)
	for _, val := range thefinds {
		workmap[val.UID] = val
	}

	return workmap

}

// loadworksintoauthors - load all works in the workmap into the authormap WorkList
func loadworksintoauthors(aa map[string]DbAuthor, ww map[string]DbWork) map[string]DbAuthor {
	for _, w := range ww {
		aa[w.FindAuthor()].AddWork(w.UID)
	}
	return aa
}

// dateworksviaauthors - if we do now know the date of a work, give it the date of the author
func dateworksviaauthors(aa map[string]DbAuthor, ww map[string]DbWork) map[string]DbWork {
	for _, w := range ww {
		if w.ConvDate == 2500 && aa[w.FindAuthor()].ConvDate != 2500 {
			w.ConvDate = aa[w.FindAuthor()].ConvDate
		}
	}
	return ww
}
