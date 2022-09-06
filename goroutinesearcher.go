//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
)

var (
	upgrader = websocket.Upgrader{}
)

//
// THE MANAGER
//

// HGoSrch - the core of a search: coordingates the dispatch of queries and collection of results
func HGoSrch(ss SearchStruct) SearchStruct {
	// NOTE: this is all much more "go-like" than HipparchiaGolangSearcher() in grabber.go,
	// BUT python + redis + HipparchiaGolangSearcher() is marginally faster than what follows [channels produce overhead?]

	// see https://go.dev/blog/pipelines : see Parallel digestion & Fan-out, fan-in & Explicit cancellation
	// https://medium.com/amboss/applying-modern-go-concurrency-patterns-to-data-pipelines-b3b5327908d4
	// https://github.com/amboss-mededu/go-pipeline-article/blob/fe0cebe78ecc9c57cdb1ac83ae6af1cda44de475/main.go
	// https://itnext.io/golang-pipeline-in-practise-6007dafbb85f
	// https://medium.com/geekculture/golang-concurrency-patterns-fan-in-fan-out-1ee43c6830c4
	// https://pranav93.github.io/blog/golang-fan-inout-pattern/
	// https://github.com/luk4z7/go-concurrency-guide

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	proghost := progresssocket("pp_" + ss.ID)
	defer proghost.Close()

	reshost := progresssocket("rc_" + ss.ID)
	defer reshost.Close()

	emitqueries, err := SrchFeeder(ctx, proghost, &ss)
	chke(err)

	var findchannels []<-chan []DbWorkline

	workers := runtime.NumCPU()
	// to slow things down for testing...
	// workers = 2

	for i := 0; i < workers; i++ {
		fc, e := SrchConsumer(ctx, emitqueries)
		chke(e)
		findchannels = append(findchannels, fc)
	}

	max := ss.Limit
	if ss.HasPhrase {
		// windowing double-hits; c. 55% are valid; these get pared via findphrasesacrosslines()
		max = ss.Limit * 3
	}

	results := ResultCollation(ctx, reshost, max, ResultAggregator(ctx, findchannels...))
	if int64(len(results)) > max {
		results = results[0:max]
	}

	ss.Results = sortresults(results, ss)
	return ss
}

//
// THE WORKERS
//

// SrchFeeder - emit items to a channel from the []PrerolledQuery that will be consumed by the SrchConsumer
func SrchFeeder(ctx context.Context, host net.Listener, ss *SearchStruct) (<-chan PrerolledQuery, error) {
	emitqueries := make(chan PrerolledQuery, cfg.WorkerCount)
	remainder := -1
	ctxclosed := false

	// channel emitter: i.e., the actual work
	go func() {
		defer close(emitqueries)
		for i, q := range ss.Queries {
			// fmt.Println(q)
			select {
			case <-ctx.Done():
				ctxclosed = true
				break
			default:
				remainder = len(ss.Queries) - i - 1
				emitqueries <- q
			}
		}
	}()

	// unix socket remainder broadcaster: i.e., the fluff
	go func() {
		if host == nil {
			msg("progresssocket() has no access to a socket", 1)
			return
		}

		for {
			if ctxclosed || remainder == 0 {
				break
			}

			// [a] wait for someone to listen
			guest, err := host.Accept()
			if err != nil {
				continue
			}

			go func() {
				// send remainder value to it
				for {
					if ctxclosed || remainder == 0 {
						_, ioe := io.WriteString(guest, fmt.Sprintf("%d\n", remainder))
						chke(ioe)
						break
					} else if !ctxclosed && remainder > -1 {
						// msg(fmt.Sprintf("remain: %d", remainder), 1)
						_, ioe := io.WriteString(guest, fmt.Sprintf("%d\n", remainder))
						if ioe != nil {
							break // e.g., client disconnected
						}
					}
				}
			}()
		}
	}()

	return emitqueries, nil
}

// SrchConsumer - grab a PrerolledQuery; execute search; emit finds to a channel
func SrchConsumer(ctx context.Context, prq <-chan PrerolledQuery) (<-chan []DbWorkline, error) {
	emitfinds := make(chan []DbWorkline)
	go func() {
		dbpool := GetPSQLconnection()
		defer dbpool.Close()
		defer close(emitfinds)
		for q := range prq {
			select {
			case <-ctx.Done():
				return
			default:
				emitfinds <- worklinequery(q, dbpool)
			}
		}
	}()
	return emitfinds, nil
}

// ResultAggregator - gather all hits from the findchannels into one place and then feed them to ResultCollation
func ResultAggregator(ctx context.Context, findchannels ...<-chan []DbWorkline) <-chan []DbWorkline {
	var wg sync.WaitGroup
	emitaggregate := make(chan []DbWorkline)
	broadcast := func(ll <-chan []DbWorkline) {
		defer wg.Done()
		for l := range ll {
			select {
			case emitaggregate <- l:
			case <-ctx.Done():
				return
			}
		}
	}
	wg.Add(len(findchannels))
	for _, fc := range findchannels {
		go broadcast(fc)
	}

	go func() {
		wg.Wait()
		close(emitaggregate)
	}()

	return emitaggregate
}

// ResultCollation - return the actual []DbWorkline results after pulling them from the ResultAggregator channel
func ResultCollation(ctx context.Context, host net.Listener, max int64, values <-chan []DbWorkline) []DbWorkline {
	var allhits []DbWorkline
	done := false
	for {
		// the actual substance of the thing
		if done {
			break
		}
		select {
		case <-ctx.Done():
			log.Print(ctx.Err().Error())
			done = true
		case val, ok := <-values:
			if ok {
				allhits = append(allhits, val...)
				if int64(len(allhits)) > max {
					// you popped over the cap...: this does in fact save time and exit in the middle
					// προκατελαβον cap of one: [Δ: 0.112s] HGoSrch()
					// προκατελαβον uncapped:   [Δ: 1.489s] HGoSrch()
					done = true
				}
			} else {
				done = true
			}
		}

		// unix socket hits broadcaster: i.e., the fluff
		go func() {
			// [a] open a tcp port to broadcast on

			for {
				if done == true {
					break
				}
				// [b] wait for someone to listen
				guest, err := host.Accept()
				if err != nil {
					continue
				}
				go func() {
					// send remainder value to it
					for {
						_, err := io.WriteString(guest, fmt.Sprintf("%d\n", len(allhits)))
						if err != nil {
							break
						}
						if done == true {
							break
						}
					}
				}()
			}
		}()

	}
	return allhits
}

//
// HELPER FUNCTIONS
//

func sortresults(results []DbWorkline, ss SearchStruct) []DbWorkline {
	// Closures that order the DbWorkline structure:
	// see generichelpers.go and https://pkg.go.dev/sort#example__sortMultiKeys
	nameIncreasing := func(one, two *DbWorkline) bool {
		a1 := AllAuthors[one.FindAuthor()].Shortname
		a2 := AllAuthors[two.FindAuthor()].Shortname
		return a1 < a2
	}

	titleIncreasing := func(one, two *DbWorkline) bool {
		return AllWorks[one.WkUID].Title < AllWorks[two.WkUID].Title
	}

	dateIncreasing := func(one, two *DbWorkline) bool {
		return AllWorks[one.FindWork()].ConvDate < AllWorks[two.FindWork()].ConvDate
	}

	//dateDecreasing := func(one, two *DbWorkline) bool {
	//	return AllWorks[one.FindWork()].ConvDate > AllWorks[two.FindWork()].ConvDate
	//}

	increasingLines := func(one, two *DbWorkline) bool {
		return one.TbIndex < two.TbIndex
	}

	//decreasingLines := func(one, two *DbWorkline) bool {
	//	return one.TbIndex > two.TbIndex // Note: > orders downwards.
	//}

	increasingID := func(one, two *DbWorkline) bool {
		return one.BuildHyperlink() < two.BuildHyperlink()
	}

	crit := sessions[ss.User].SortHitsBy

	switch {
	// unhandled are "location" & "provenance"
	case crit == "shortname":
		OrderedBy(nameIncreasing, titleIncreasing, increasingLines).Sort(results)
		return results
	case crit == "converted_date":
		OrderedBy(dateIncreasing, nameIncreasing, titleIncreasing, increasingLines).Sort(results)
		return results
	case crit == "universalid":
		OrderedBy(increasingID).Sort(results)
		return results
	default:
		// author nameIncreasing
		OrderedBy(nameIncreasing, increasingLines).Sort(results)
		return results
	}
}

// progresssocket - from where should the progress info be served?
func progresssocket(name string) net.Listener {
	// return a listener and the value of the port selected
	// socket vs tcp: tcp connection ends up in TIME_WAIT state and will block the port

	host, err := net.Listen("unix", fmt.Sprintf("%s/hgs_%s", UNIXSOCKETPATH, name))

	if err != nil {
		msg(fmt.Sprintf("progresssocket() could not open '%s/hgs_%s'", UNIXSOCKETPATH, name), 1)
	} else {
		// msg(fmt.Sprintf("progresssocket() opened '%s/hgs_%s'", UNIXSOCKETPATH, name), 1)
		return host
	}
	return nil
}
