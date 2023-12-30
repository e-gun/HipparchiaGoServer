//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"sync"
)

//
// THE MANAGER
//

// HGoSrch - the core of a search: coordinates the dispatch of queries and collection of results
func HGoSrch(ss SearchStruct) SearchStruct {
	// NOTE: this is all much more "go-like" than HipparchiaGolangSearcher() in grabber.go,

	// see https://go.dev/blog/pipelines : see Parallel digestion & Fan-out, fan-in & Explicit cancellation
	// https://medium.com/amboss/applying-modern-go-concurrency-patterns-to-data-pipelines-b3b5327908d4
	// https://github.com/amboss-mededu/go-pipeline-article/blob/fe0cebe78ecc9c57cdb1ac83ae6af1cda44de475/main.go
	// https://itnext.io/golang-pipeline-in-practise-6007dafbb85f
	// https://medium.com/geekculture/golang-concurrency-patterns-fan-in-fan-out-1ee43c6830c4
	// https://pranav93.github.io/blog/golang-fan-inout-pattern/
	// https://github.com/luk4z7/go-concurrency-guide

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emitqueries, err := SrchFeeder(ctx, &ss)
	chke(err)

	workers := Config.WorkerCount

	findchannels := make([]<-chan []DbWorkline, workers)

	for i := 0; i < workers; i++ {
		fc, e := SrchConsumer(ctx, emitqueries)
		chke(e)
		findchannels[i] = fc
	}

	mx := ss.CurrentLimit
	if ss.HasPhraseBoxA {
		// windowing generates double-hits; c. 55% are valid; these get pared via findphrasesacrosslines()
		mx = ss.CurrentLimit * 3
	}

	results := ResultCollation(ctx, &ss, mx, ResultAggregator(ctx, findchannels...))
	if len(results) > mx {
		results = results[0:mx]
	}

	ss.Results = results
	return ss
}

//
// THE WORKERS
//

// SrchFeeder - emit items to a channel from the []PrerolledQuery that will be consumed by the SrchConsumer
func SrchFeeder(ctx context.Context, ss *SearchStruct) (<-chan PrerolledQuery, error) {
	emitqueries := make(chan PrerolledQuery, Config.WorkerCount)
	remainder := -1

	emitone := func(i int) {
		remainder = len(ss.Queries) - i - 1
		if remainder%POLLEVERYNTABLES == 0 {
			SIUpdateRemain <- SIKVi{ss.ID, remainder}
		}
		emitqueries <- ss.Queries[i]
	}

	feed := func() {
		defer close(emitqueries)
		for i := 0; i < len(ss.Queries); i++ {
			select {
			case <-ctx.Done():
				break
			default:
				emitone(i)
			}
		}
	}

	go feed()

	return emitqueries, nil
}

// SrchConsumer - grab a PrerolledQuery; execute search; emit finds to a channel
func SrchConsumer(ctx context.Context, prq <-chan PrerolledQuery) (<-chan []DbWorkline, error) {
	emitfinds := make(chan []DbWorkline)

	consume := func() {
		conn := GrabDBConnection()
		defer conn.Release()

		defer close(emitfinds)
		for q := range prq {
			select {
			case <-ctx.Done():
				return
			default:
				emitfinds <- WorklineQuery(q, conn)
			}
		}
	}

	go consume()

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
func ResultCollation(ctx context.Context, ss *SearchStruct, maxhits int, foundlines <-chan []DbWorkline) []DbWorkline {
	var allhits []DbWorkline

	addhits := func(worklines []DbWorkline) {
		// each bundle of []DbWorkline comes off of a single author table
		// so OneHit searches will just grab the top of that bundle
		if ss.OneHit && ss.PhaseNum == 1 && len(worklines) > 0 {
			allhits = append(allhits, worklines[0])
		} else {
			allhits = append(allhits, worklines...)
		}
		SIUpdateHits <- SIKVi{ss.ID, len(allhits)}
	}

	done := false
	for {
		if done {
			break
		}
		select {
		case <-ctx.Done():
			// log.Print(ctx.Err().Error())
			done = true
		case ll, ok := <-foundlines:
			if ok {
				addhits(ll)
				if len(allhits) > maxhits {
					done = true
				}
			} else {
				done = true
			}
		}
	}
	return allhits
}
