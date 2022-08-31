package main

import (
	"context"
	"log"
	"runtime"
	"sync"
)

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

	emitqueries, err := SrchFeeder(ctx, ss.Queries)
	chke(err)

	var findchannels []<-chan []DbWorkline

	for i := 0; i < runtime.NumCPU(); i++ {
		fc, e := SrchConsumer(ctx, emitqueries)
		chke(e)
		findchannels = append(findchannels, fc)
	}

	max := ss.Limit
	if ss.HasPhrase {
		// windowing double-hits; c. 55% are valid; these get pared via findphrasesacrosslines()
		max = ss.Limit * 3
	}

	results := ResultCollation(ctx, max, ResultAggregator(ctx, findchannels...))
	if int64(len(results)) > max {
		results = results[0:max]
	}

	ss.Results = sortresults(results, ss)

	return ss
}

// SrchFeeder - emit items to a channel from the []PrerolledQuery that will be consumed by the SrchConsumer
func SrchFeeder(ctx context.Context, qq []PrerolledQuery) (<-chan PrerolledQuery, error) {
	emitqueries := make(chan PrerolledQuery, cfg.WorkerCount)
	go func() {
		defer close(emitqueries)
		for _, q := range qq {
			// fmt.Println(q)
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

// SrchConsumer - grab a PrerolledQuery; execute search; emit finds to a channel
func SrchConsumer(ctx context.Context, prq <-chan PrerolledQuery) (<-chan []DbWorkline, error) {
	emitfinds := make(chan []DbWorkline)
	go func() {
		dbpool := grabpgsqlconnection()
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
func ResultCollation(ctx context.Context, max int64, values <-chan []DbWorkline) []DbWorkline {
	var allhits []DbWorkline
	for {
		select {
		case <-ctx.Done():
			log.Print(ctx.Err().Error())
			return allhits
		case val, ok := <-values:
			if ok {
				// the progress poll should be attached here
				// fmt.Println(fmt.Sprintf("current count: %d", len(allhits)))
				allhits = append(allhits, val...)
				if int64(len(allhits)) > max {
					// you popped over the cap...: this does in fact save time and exit in the middle
					// προκατελαβον cap of one: [Δ: 0.112s] HGoSrch()
					// προκατελαβον uncapped:   [Δ: 1.489s] HGoSrch()
					return allhits
				}
			} else {
				return allhits
			}
		}
	}
}

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

	crit := sessions[ss.User].SrchOutSettings.SortHitsBy

	switch {
	case crit == "Name":
		OrderedBy(nameIncreasing, titleIncreasing, increasingLines).Sort(results)
		return results
	case crit == "Date":
		OrderedBy(dateIncreasing, nameIncreasing, titleIncreasing, increasingLines).Sort(results)
		return results
	case crit == "ID":
		OrderedBy(increasingID).Sort(results)
		return results
	default:
		// author nameIncreasing
		OrderedBy(nameIncreasing, increasingLines).Sort(results)
		return results
	}
}
