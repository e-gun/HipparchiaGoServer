//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
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

// SearchAndInsertResults - take a SearchStruct; fan out its []PrerolledQuery; collect the results; insert a WorkLineBundle into the SearchStruct
func SearchAndInsertResults(ss *SearchStruct) {
	// see https://go.dev/blog/pipelines : see Parallel digestion & Fan-out, fan-in & Explicit cancellation
	// https://medium.com/amboss/applying-modern-go-concurrency-patterns-to-data-pipelines-b3b5327908d4
	// https://github.com/amboss-mededu/go-pipeline-article/blob/fe0cebe78ecc9c57cdb1ac83ae6af1cda44de475/main.go
	// https://itnext.io/golang-pipeline-in-practise-6007dafbb85f
	// https://medium.com/geekculture/golang-concurrency-patterns-fan-in-fan-out-1ee43c6830c4
	// https://pranav93.github.io/blog/golang-fan-inout-pattern/
	// https://github.com/luk4z7/go-concurrency-guide

	// theoretically possible to yield up the interim results while the search is in progress; but a pain/gain problem
	// specifically, two-part searches will always need a lot of fussing... websocket is perhaps the way to go

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // nobody sends "<-ctx.Done()" as seen below; but this gives you that

	emitqueries, err := SrchFeeder(ctx, ss)
	chke(err)

	workers := Config.WorkerCount

	findchannels := make([]<-chan *WorkLineBundle, workers)

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

	ResultCollation(ctx, ss, mx, ResultAggregator(ctx, findchannels...))
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
			SIUpdateRemain <- SIKVi{ss.WSID, remainder}
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
func SrchConsumer(ctx context.Context, prq <-chan PrerolledQuery) (<-chan *WorkLineBundle, error) {
	emitfinds := make(chan *WorkLineBundle)

	consume := func() {
		dbconn := GetDBConnection()
		defer dbconn.Release()
		defer close(emitfinds)
		for q := range prq {
			select {
			case <-ctx.Done():
				return
			default:
				wlb := AcquireWorkLineBundle(q, dbconn)
				emitfinds <- wlb
			}
		}
	}

	go consume()

	return emitfinds, nil
}

// ResultAggregator - gather all hits from the findchannels into one place and then feed them to ResultCollation
func ResultAggregator(ctx context.Context, findchannels ...<-chan *WorkLineBundle) <-chan *WorkLineBundle {
	var wg sync.WaitGroup
	emitaggregate := make(chan *WorkLineBundle)
	broadcast := func(wlbb <-chan *WorkLineBundle) {
		defer wg.Done()
		for b := range wlbb {
			select {
			case emitaggregate <- b:
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

// ResultCollation - insert the actual WorkLineBundle results into the SearchStruct after pulling them from the ResultAggregator channel
func ResultCollation(ctx context.Context, ss *SearchStruct, maxhits int, foundbundle <-chan *WorkLineBundle) {
	var collated WorkLineBundle

	addhits := func(foundbundle *WorkLineBundle) {
		// each foundbundle comes off of a single author table
		// so OneHit searches will just grab the top of that bundle
		if ss.OneHit && ss.PhaseNum == 1 && !foundbundle.IsEmpty() {
			collated.AppendOne(foundbundle.FirstLine())
		} else {
			collated.AppendLines(foundbundle.Lines)
		}
		SIUpdateHits <- SIKVi{ss.WSID, collated.Len()}
	}

	done := false
	for {
		if done {
			break
		}
		select {
		case <-ctx.Done():
			done = true
		case lb, ok := <-foundbundle:
			if ok {
				addhits(lb)
				if collated.Len() > maxhits {
					done = true
				}
			} else {
				done = true
			}
		}
	}

	collated.ResizeTo(maxhits) // a cap of N can collect >N hits before the "if collated.Len() > maxhits" halt check is made
	ss.Results = collated
}
