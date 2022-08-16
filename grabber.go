//    HipparchiaGoDBHelper: search and vector helper app and functions for HipparchiaServer
//    Copyright: E Gunderson 2021
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

//
// the GRABBER is supposed to be pointedly basic
// [a] it looks to redis for a pile of SQL queries that were pre-rolled
// [b] it asks postgres to execute these queries
// [c] it stores the results on redis
// [d] it also updates the redis progress poll data relative to this search
//

package main

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"runtime"
	"sync"
)

// HipparchiaGolangSearcher : Execute a series of SQL queries stored in redis by dispatching a collection of goroutines
// the python module calls this; the module needs to be able to set the logging level, etc.
// and so you set up 'cfg' and then you can call HipparchiaSearcher()
func HipparchiaGolangSearcher(thekey string, hitcap int64, workercount int, ll int, rl RedisLogin, pl PostgresLogin) string {
	cfg.RedisKey = thekey
	cfg.MaxHits = hitcap
	cfg.WorkerCount = workercount
	cfg.LogLevel = ll
	cfg.RLogin = rl
	cfg.PGLogin = pl

	resultkey := HipparchiaSearcher()
	return resultkey
}

func HipparchiaSearcher() string {
	msg(fmt.Sprintf("Searcher Launched"), 1)

	runtime.GOMAXPROCS(cfg.WorkerCount + 1)

	recordinitialsizeofworkpile(cfg.RedisKey)

	var awaiting sync.WaitGroup

	for i := 0; i < cfg.WorkerCount; i++ {
		awaiting.Add(1)
		go grabber(i, cfg.RedisKey, &awaiting)
	}

	awaiting.Wait()

	resultkey := cfg.RedisKey + "_results"
	return resultkey
}

func grabber(clientnumber int, searchkey string, awaiting *sync.WaitGroup) {
	// this is where all of the work happens
	defer awaiting.Done()
	msg(fmt.Sprintf("Hello from grabber %d", clientnumber), 3)

	rc := grabredisconnection()
	defer func(rc redis.Conn) {
		err := rc.Close()
		checkerror(err)
	}(rc)

	dbpool := grabpgsqlconnection()
	defer dbpool.Close()

	resultkey := searchkey + "_results"

	for {
		// [i] get a pre-rolled query or break the loop
		thequery := rcpopstr(rc, searchkey)
		if thequery == "SET_IS_EMPTY" {
			break
		}

		// [ii] - [v] inside findtherows() because its code is common with HipparchiaBagger's needs
		thesefinds := findtherows(thequery, "grabber", searchkey, clientnumber, rc, dbpool)

		// [vi.2] load via pipeline
		err := rc.Send("MULTI")
		checkerror(err)

		for i := 0; i < len(thesefinds); i++ {
			jsonhit, ee := json.Marshal(thesefinds[i])
			checkerror(ee)

			e := rc.Send("SADD", resultkey, jsonhit)
			checkerror(e)
		}

		_, e := rc.Do("EXEC")
		checkerror(e)

		// [vi.3] busted the cap?
		done := checkcap(searchkey, clientnumber, rc)
		if done {
			// trigger the break in the outer loop
			rcdel(rc, searchkey)
		}
	}
}

func checkcap(searchkey string, client int, rc redis.Conn) bool {
	resultkey := searchkey + "_results"
	hitcount, e := redis.Int64(rc.Do("SCARD", resultkey))
	checkerror(e)

	k := searchkey + "_hitcount"
	_, ee := rc.Do("SET", k, hitcount)
	checkerror(ee)
	msg(fmt.Sprintf("grabber #%d reports that the hitcount is %d", client, hitcount), 3)
	if hitcount >= cfg.MaxHits {
		// trigger the break in the outer loop
		return true
	} else {
		return false
	}
}

func recordinitialsizeofworkpile(k string) {
	rc := grabredisconnection()

	defer func(rc redis.Conn) {
		err := rc.Close()
		checkerror(err)
	}(rc)

	remain, err := redis.Int64(rc.Do("SCARD", k))
	checkerror(err)
	kk := fmt.Sprintf("%s_poolofwork", k)
	_, e := rc.Do("SET", kk, remain)
	checkerror(e)

	msg(fmt.Sprintf("recordinitialsizeofworkpile(): initial size of workpile for '%s' is %d", k+"_poolofwork", remain), 2)
}

func fetchfinalnumberofresults(k string) int64 {
	rc := grabredisconnection()

	defer func(rc redis.Conn) {
		err := rc.Close()
		checkerror(err)
	}(rc)

	kk := fmt.Sprintf("%s_results", k)
	hits, err := redis.Int64(rc.Do("SCARD", kk))
	checkerror(err)
	return hits
}
