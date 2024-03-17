//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package search

import (
	"context"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/launch"
	"github.com/e-gun/HipparchiaGoServer/internal/m"
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
	"github.com/e-gun/HipparchiaGoServer/internal/vaults"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"sync"
)

var (
	msg = m.NewMessageMaker(launch.BuildDefaultConfig(), m.LaunchStruct{})
)

// SearchAndInsertResults - take a SearchStruct; fan out its []PrerolledQuery; collect the results; insert a WorkLineBundle into the SearchStruct
func SearchAndInsertResults(ss *structs.SearchStruct) {
	// see https://go.dev/blog/pipelines : see Parallel digestion & Fan-out, fan-in & Explicit cancellation
	// https://medium.com/amboss/applying-modern-go-concurrency-patterns-to-data-pipelines-b3b5327908d4
	// https://github.com/amboss-mededu/go-pipeline-article/blob/fe0cebe78ecc9c57cdb1ac83ae6af1cda44de475/main.go
	// https://itnext.io/golang-pipeline-in-practise-6007dafbb85f
	// https://medium.com/geekculture/golang-concurrency-patterns-fan-in-fan-out-1ee43c6830c4
	// https://pranav93.github.io/blog/golang-fan-inout-pattern/
	// https://github.com/luk4z7/go-concurrency-guide

	// theoretically possible to yield up the interim results while the search is in progress; but a pain/gain problem
	// specifically, two-part searches will always need a lot of fussing... websocket is perhaps the way to go

	defer ss.CancelFnc()

	// [a] load the queries into a channel
	querychannel, err := SearchQueryFeeder(ss)
	msg.EC(err)

	// [b] fan out to run searches in parallel; searches fed by the query channel
	workers := launch.Config.WorkerCount
	searchchannels := make([]<-chan *structs.WorkLineBundle, workers)

	for i := 0; i < workers; i++ {
		foundlineschannel, e := PRQSearcher(ss.Context, querychannel)
		msg.EC(e)
		searchchannels[i] = foundlineschannel
	}

	mx := ss.CurrentLimit
	if ss.HasPhraseBoxA {
		// windowing generates double-hits; c. 55% are valid; these get pared via FindPhrasesAcrossLines()
		mx = ss.CurrentLimit * 3
	}

	// [c] fan in to gather the results into a single channel
	resultchan := ResultChannelAggregator(ss.Context, searchchannels...)

	// [d] pull the results off of the result channel and collate them
	FinalResultCollation(ss, mx, resultchan)
}

// SearchQueryFeeder - emit items to a channel from the []PrerolledQuery; they will be consumed by the PRQSearcher
func SearchQueryFeeder(ss *structs.SearchStruct) (<-chan structs.PrerolledQuery, error) {
	emitqueries := make(chan structs.PrerolledQuery, launch.Config.WorkerCount)
	remainder := -1

	emitone := func(i int) {
		remainder = len(ss.Queries) - i - 1
		if remainder%vv.POLLEVERYNTABLES == 0 {
			vaults.WSInfo.UpdateRemain <- vaults.WSSIKVi{ss.WSID, remainder}
		}
		emitqueries <- ss.Queries[i]
	}

	feed := func() {
		defer close(emitqueries)
		for i := 0; i < len(ss.Queries); i++ {
			select {
			case <-ss.Context.Done():
				break
			default:
				emitone(i)
			}
		}
	}

	go feed()

	return emitqueries, nil
}

// PRQSearcher - this is where the search happens... grab a PrerolledQuery; execute search; emit finds to a channel
func PRQSearcher(ctx context.Context, querychannel <-chan structs.PrerolledQuery) (<-chan *structs.WorkLineBundle, error) {
	foundlineschannel := make(chan *structs.WorkLineBundle)

	consume := func() {
		dbconn := db.GetDBConnection()
		defer dbconn.Release()
		defer close(foundlineschannel)
		for q := range querychannel {
			select {
			case <-ctx.Done():
				return
			default:
				// execute a search and send the finds over the channel
				b := AcquireWorkLineBundle(q, dbconn)
				foundlineschannel <- b
			}
		}
	}

	go consume()

	return foundlineschannel, nil
}

// ResultChannelAggregator - gather all hits from the searchchannels into one place and then feed them to FinalResultCollation
func ResultChannelAggregator(ctx context.Context, searchchannels ...<-chan *structs.WorkLineBundle) <-chan *structs.WorkLineBundle {
	var wg sync.WaitGroup
	resultchann := make(chan *structs.WorkLineBundle)

	broadcast := func(wlbb <-chan *structs.WorkLineBundle) {
		defer wg.Done()
		for b := range wlbb {
			select {
			case resultchann <- b:
			case <-ctx.Done():
				return
			}
		}
	}

	wg.Add(len(searchchannels))
	for _, fc := range searchchannels {
		go broadcast(fc)
	}

	go func() {
		wg.Wait()
		close(resultchann)
	}()

	return resultchann
}

// FinalResultCollation - insert the actual WorkLineBundle results into the SearchStruct after pulling them from the ResultChannelAggregator channel
func FinalResultCollation(ss *structs.SearchStruct, maxhits int, foundbundle <-chan *structs.WorkLineBundle) {
	var collated structs.WorkLineBundle

	addhits := func(foundbundle *structs.WorkLineBundle) {
		// each foundbundle comes off of a single author table
		// so OneHit searches will just grab the top of that bundle
		if ss.OneHit && ss.PhaseNum == 1 && !foundbundle.IsEmpty() {
			collated.AppendOne(foundbundle.FirstLine())
		} else {
			collated.AppendLines(foundbundle.Lines)
		}
		vaults.WSInfo.UpdateHits <- vaults.WSSIKVi{ss.WSID, collated.Len()}
	}

	done := false
	for {
		if done {
			break
		}
		select {
		case <-ss.Context.Done():
			done = true
		case lb, ok := <-foundbundle:
			if ok {
				addhits(lb)
				if collated.Len() > maxhits {
					collated.ResizeTo(maxhits)
					done = true
				}
			} else {
				done = true
			}
		}
	}

	ss.Results = collated
}
