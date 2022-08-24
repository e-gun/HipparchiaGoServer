package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"runtime"
	"sync"
)

func HGoSrch(s SearchStruct) SearchStruct {
	// see https://go.dev/blog/pipelines : see Parallel digestion & Fan-out, fan-in & Explicit cancellation
	// https://medium.com/amboss/applying-modern-go-concurrency-patterns-to-data-pipelines-b3b5327908d4
	// https://github.com/amboss-mededu/go-pipeline-article/blob/fe0cebe78ecc9c57cdb1ac83ae6af1cda44de475/main.go
	// https://itnext.io/golang-pipeline-in-practise-6007dafbb85f
	// https://medium.com/geekculture/golang-concurrency-patterns-fan-in-fan-out-1ee43c6830c4
	// https://pranav93.github.io/blog/golang-fan-inout-pattern/

	msg(fmt.Sprintf("Searcher Launched"), 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emitqueries, err := SrchFeeder(ctx, s.Queries)
	checkerror(err)

	findchannels := []<-chan []DbWorkline{}

	for i := 0; i < runtime.NumCPU(); i++ {
		fc, e := SrchConsumer(ctx, emitqueries)
		checkerror(e)
		findchannels = append(findchannels, fc)
	}

	merge := ResultAggregator(ctx, findchannels...)
	s.Results = ResultCollation(ctx, merge)

	return s
}

// feeder

func SrchFeeder(ctx context.Context, qq []PrerolledQuery) (<-chan PrerolledQuery, error) {
	emitqueries := make(chan PrerolledQuery, cfg.WorkerCount)
	go func() {
		defer close(emitqueries)
		for _, q := range qq {
			select {
			case <-ctx.Done():
				return
			default:
				emitqueries <- q
			}
		}
	}()
	return emitqueries, nil
}

// consumer

func SrchConsumer(ctx context.Context, prq <-chan PrerolledQuery) (<-chan []DbWorkline, error) {
	emitfinds := make(chan []DbWorkline)
	go func() {
		dbpool := grabpgsqlconnection()
		defer close(emitfinds)
		for q := range prq {
			select {
			case <-ctx.Done():
				return
			default:
				emitfinds <- modworklinequery(q, dbpool)
			}
		}
	}()
	return emitfinds, nil
}

// aggregator

func ResultAggregator(ctx context.Context, hitchannels ...<-chan []DbWorkline) <-chan []DbWorkline {
	var wg sync.WaitGroup
	emitaggregate := make(chan []DbWorkline)
	output := func(ll <-chan []DbWorkline) {
		defer wg.Done()
		for l := range ll {
			select {
			case emitaggregate <- l:
			case <-ctx.Done():
				return
			}
		}
	}
	wg.Add(len(hitchannels))
	for _, h := range hitchannels {
		go output(h)
	}

	go func() {
		wg.Wait()
		close(emitaggregate)
	}()

	return emitaggregate
}

func ResultCollation(ctx context.Context, values <-chan []DbWorkline) []DbWorkline {
	// will this require locks?
	var allhits []DbWorkline
	for {
		select {
		case <-ctx.Done():
			log.Print(ctx.Err().Error())
			return allhits
		case val, ok := <-values:
			if ok {
				allhits = append(allhits, val...)
			} else {
				return allhits
			}
		}
	}
}

func modworklinequery(prq PrerolledQuery, dbpool *pgxpool.Pool) []DbWorkline {
	// we omit keeping polling data...

	// [iv] build a temp table if needed
	if prq.TempTable != "" {
		_, err := dbpool.Exec(context.Background(), prq.TempTable)
		checkerror(err)
	}

	// [v] execute the main query
	var foundrows pgx.Rows
	var err error

	if prq.PsqlData != "" {
		foundrows, err = dbpool.Query(context.Background(), prq.PsqlQuery, prq.PsqlData)
		checkerror(err)
	} else {
		foundrows, err = dbpool.Query(context.Background(), prq.PsqlQuery)
		checkerror(err)
	}

	// [vi] iterate through the finds
	// don't check-and-load find-by-find because some searches are effectively uncapped
	// faster to test only after you finish each query
	// can over-stuff redis because HipparchaServer should only display hitcap results no matter how many you push

	var thesefinds []DbWorkline

	defer foundrows.Close()
	for foundrows.Next() {
		// [vi.1] convert the finds into DbWorklines
		var thehit DbWorkline
		err := foundrows.Scan(&thehit.WkUID, &thehit.TbIndex, &thehit.Lvl5Value, &thehit.Lvl4Value, &thehit.Lvl3Value,
			&thehit.Lvl2Value, &thehit.Lvl1Value, &thehit.Lvl0Value, &thehit.MarkedUp, &thehit.Accented,
			&thehit.Stripped, &thehit.Hypenated, &thehit.Annotations)
		checkerror(err)
		thesefinds = append(thesefinds, thehit)
	}

	return thesefinds
}
