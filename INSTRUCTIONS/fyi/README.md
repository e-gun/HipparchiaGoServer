
# HipparchiaGoServer FYI

## Frontpage size

`10 requests 1.46 MB / 1.47 MB transferred`

the fonts are most of the data; they have been subsetted (but `iosevka` is 756k of the above)

comparanda:

* `nytimes.com` is c. 7MB.
* `amazon.com` is c. 7MB.
* `instagram.com` is c. 8.75MB.
* `bing.com` is c. 9.5MB.

## CLI

![options](../gitimg/hgscli.png)

## self-test

self-test without vectors is now `HipparchiaGoServer -st -dv`

``` 
e-gun/HipparchiaGoServer/ % ./HipparchiaGoServer -st -gl 3
[HGS] HipparchiaGoServer (v2.0.0-pre) [git: db012bd0] [default.pgo] [gl=3; el=0]
	Built:	2025-06-17@10:22:24		Golang:	go1.24.4
	System:	darwin-arm64			WKvCPU:	20/20
[HGS-DBI] Number of stored vector models: 12
[HGS] [A1: 0.078s][Δ: 0.078s] 7462 works built: map[string]DbWork
[HGS] [A2: 0.087s][Δ: 0.009s] 2186 authors built: map[string]DbAuthor
[HGS] [A3: 0.090s][Δ: 0.002s] corpus maps built
[HGS] [B1: 0.154s][Δ: 0.154s] unnested lemma map built (158966 items)
[HGS] [B2: 0.203s][Δ: 0.049s] nested lemma map built
[HGS] initialization took 0.292s
[HGS] to stop the server press Control-C or close this window
[HGS-WEB] (tls unavailable)
[HGS-SELFTEST] Running Selftest 1 of 1
[HGS-SELFTEST] entering selftestsuite mode (4 segments)
[HGS-SELFTEST] [I] 6 search tests
⇨ http server started on 127.0.0.1:8001
[HGS-SELFTEST] [A1: 1.250s][Δ: 1.250s] single word in corpus: 'vervex'
[HGS-SELFTEST] [A2: 2.667s][Δ: 1.417s] phrase in corpus: 'plato omnem'
[HGS-SELFTEST] [A3: 4.313s][Δ: 1.645s] phrase near phrase: 'καὶ δὴ καὶ' near 'εἴ που καὶ'
[HGS-SELFTEST] [B1: 5.680s][Δ: 1.367s] lemma in corpus: 'φθορώδηϲ'
[HGS-SELFTEST] [B2: 7.377s][Δ: 1.697s] lemma near phrase: 'γαῖα' near 'ἐϲχάτη χθονόϲ'
[HGS-SELFTEST] [B3: 27.858s][Δ: 20.481s] lemma near lemma in corpus: 'πόλιϲ' near 'ὁπλίζω'
[HGS-SELFTEST] [II] 3 text, index, and vocab maker tests
[HGS-SELFTEST] [C1: 28.266s][Δ: 0.409s] build a text for 40000 arbitrary lines
[HGS-SELFTEST] [C2: 28.634s][Δ: 0.368s] build an index to 40000 arbitrary lines
[HGS-SELFTEST] [C3: 32.720s][Δ: 4.087s] build vocabulary list for 40000 arbitrary lines
[HGS-SELFTEST] [III] 4 browsing and lexical tests
[HGS-WEB] could not find a work for tlg0021w001
[HGS-WEB] could not find a work for tlg0025w001
[HGS-SELFTEST] [D1: 32.913s][Δ: 0.193s] browse 50 passages
[HGS-SELFTEST] [D2: 35.602s][Δ: 2.689s] look up 48 specific words
[HGS-SELFTEST] [D3: 36.662s][Δ: 1.059s] look up 18 word substrings
[HGS-SELFTEST] [D4: 37.864s][Δ: 1.203s] reverse lookup for 6 word substrings
[HGS-DBI] VectorDBReset() dropped semantic_vectors_nn
[HGS-SELFTEST] [IV] nearest neighbor vectorization tests
[HGS-DBI] vectordbinitnn(): success
[HGS-SELFTEST] [E1: 58.820s][Δ: 20.955s] semantic vector model test: w2v - 1 author(s) with 4 text preparation modes per author
[HGS-SELFTEST] [E2: 80.535s][Δ: 21.715s] semantic vector model test: lexvec - 1 author(s) with 4 text preparation modes per author
[HGS-SELFTEST] [E3: 118.893s][Δ: 38.358s] semantic vector model test: glove - 1 author(s) with 4 text preparation modes per author
[HGS-SELFTEST] [V] lda vectorization tests
[HGS-SELFTEST] [F: 140.244s][Δ: 21.351s] lda vector model test - 1 author(s) with 4 text preparation modes per author
[HGS-SELFTEST] exiting selftestsuite mode

```

### some vectorless and vector selftest times

* apple m1 chip and 20 workers on macOS: 38s / 140s
* intel i9-9900K and 8 workers on FreeBSD: 52s / 217s
* ryzen 9600x with 12 workers on Windows 11: 63s / 125s
* ryzen 9600x with 12 workers on Ubuntu: 30s / 99s

there is a hefty Windows tax...

### self-test: how many cores is enough?

Some individual tests split well over multiple workers. Some do not. And even then, more cores helps until it does not.

Note that 20 cores can be slower than 12 in some (fast-to-finish) tests. And more cores barely matters after 6
for vector tests. The lemmata tests yield the major differences. Diminishing returns start to hit hard around 7 cores. Real gains stop around 14 cores in every case. It looks like I/O speed can leave the cores starved for data.

The default setting is `use all cores`, but many people can safely dial this back. 

![workers vs time](../gitimg/workers_vs_time_a.png)

![workers vs time](../gitimg/workers_vs_time_e.png)

## profiling

`HipparchiaGoServer -st` called to generate
* `default.pgo`
* `MEMProfile.pdf`
* `CPUProfile.pdf`

## memory use during self-test 

### with manual gc set to `true`

``` 
% ./HipparchiaGoServer -gl 4 -el 0 -st
[HGS] HipparchiaGoServer (v2.1.0) [no pgo] [gl=4; el=0]
	Golang:	go1.26.0
	System:	darwin-arm64			WKvCPU:	20/20
[HGS-DBI] Number of stored vector models: 12
[HGS-MPS] MapNewWorkCorpus() added 837 works from 'lat'
[HGS-MPS] MapNewWorkCorpus() added 6625 works from 'tlg'
[HGS] [A1: 0.131s][Δ: 0.131s] 7462 works built: map[string]DbWork
[HGS-MPS] MapNewAuthorCorpus() added 1823 authors from 'tlg'
[HGS-MPS] MapNewAuthorCorpus() added 363 authors from 'lat'
[HGS] [A2: 0.138s][Δ: 0.007s] 2186 authors built: map[string]DbAuthor
[HGS] [A3: 0.140s][Δ: 0.003s] corpus maps built
[HGS] [A4: 0.147s][Δ: 0.010s] load word weights
[HGS] [B1: 0.178s][Δ: 0.178s] unnested lemma map built (158966 items)
[HGS] [B2: 0.225s][Δ: 0.047s] nested lemma map built
[HGS] initialization took 0.312s
[HGS] to stop the server press Control-C or close this window
[HGS] server started at http://127.0.0.1:8001 (tls unavailable)
[HGS-SELFTEST] Running Selftest 1 of 1
[HGS-SELFTEST] entering selftestsuite mode (4 segments)
[HGS-SELFTEST] [I] 6 search tests
[HGS-VLT] RtSearch() runtime.GC() 89M --> 82M
[HGS-SELFTEST] [A1: 0.764s][Δ: 0.764s] single word in corpus: 'vervex'
[HGS-VLT] RtSearch() runtime.GC() 96M --> 85M
[HGS-SELFTEST] [A2: 2.322s][Δ: 1.557s] phrase in corpus: 'plato omnem'
[HGS-SEA] [Δ: 1.317s]  WithinXLinesSearch(): 2308 initial hits
[HGS-SEA] [Δ: 0.003s]  SSBuildQueries() rerun
[HGS-SEA] [Δ: 0.026s]  WithinXLinesSearch(): 3 subsequent hits
[HGS-VLT] RtSearch() runtime.GC() 136M --> 90M
[HGS-SELFTEST] [A3: 3.700s][Δ: 1.378s] phrase near phrase: 'καὶ δὴ καὶ' near 'εἴ που καὶ'
[HGS-VLT] RtSearch() runtime.GC() 101M --> 93M
[HGS-SELFTEST] [B1: 4.800s][Δ: 1.101s] lemma in corpus: 'φθορώδηϲ'
[HGS-SEA] swapphraseandlemma() was called: lemmatized 'γαῖα' swapped with 'ἐϲχάτη χθονόϲ'
[HGS-SEA] [Δ: 1.604s]  WithinXLinesSearch(): 11 initial hits
[HGS-SEA] [Δ: 0.001s]  SSBuildQueries() rerun
[HGS-SEA] [Δ: 0.005s]  WithinXLinesSearch(): 4 subsequent hits
[HGS-VLT] RtSearch() runtime.GC() 111M --> 95M
[HGS-SELFTEST] [B2: 6.442s][Δ: 1.642s] lemma near phrase: 'γαῖα' near 'ἐϲχάτη χθονόϲ'
[HGS-SEA] pickfastestlemma() is NOT swapping πόλιϲ for ὁπλίζω: possible hits 205378 vs 3205; known forms 50 vs 192
[HGS-SEA] [Δ: 6.826s]  WithinXLinesSearch(): 99194 initial hits
[HGS-SEA] [Δ: 0.174s]  SSBuildQueries() rerun
[HGS-SEA] [Δ: 14.493s]  WithinXLinesSearch(): 101 subsequent hits
[HGS-VLT] RtSearch() runtime.GC() 189M --> 105M
[HGS-SELFTEST] [B3: 28.172s][Δ: 21.730s] lemma near lemma in corpus: 'πόλιϲ' near 'ὁπλίζω'
[HGS-SELFTEST] [II] 3 text, index, and vocab maker tests
[HGS-VLT] RtTextMaker() runtime.GC() 177M --> 138M
[HGS-SELFTEST] [C1: 28.536s][Δ: 0.364s] build a text for 40000 arbitrary lines
[HGS-VLT] RtIndexMaker() runtime.GC() 287M --> 127M
[HGS-SELFTEST] [C2: 28.829s][Δ: 0.293s] build an index to 40000 arbitrary lines
[HGS-DBI] ArrayToGetRequiredMorphObjects() will search among 177369 words
[HGS-VLT] RtVocabMaker() runtime.GC() 587M --> 143M
[HGS-SELFTEST] [C3: 33.309s][Δ: 4.480s] build vocabulary list for 40000 arbitrary lines
[HGS-SELFTEST] [III] 4 browsing and lexical tests
[HGS-VLT] RtBrowseLine() runtime.GC() 153M --> 127M
...
[HGS-WEB] could not find a work for tlg0025w001
[HGS-VLT] RtBrowseLine() runtime.GC() 127M --> 127M
...
[HGS-SELFTEST] [D1: 34.101s][Δ: 0.792s] browse 50 passages
[HGS-VLT] RtLexFindByForm() runtime.GC() 127M --> 127M
...
[HGS-SELFTEST] [D2: 37.522s][Δ: 3.421s] look up 48 specific words
[HGS-VLT] RtLexLookup() runtime.GC() 134M --> 128M
...
[HGS-SELFTEST] [D3: 38.883s][Δ: 1.361s] look up 18 word substrings
[HGS-VLT] RtLexReverse() runtime.GC() 185M --> 141M
...
[HGS-SELFTEST] [D4: 40.305s][Δ: 1.422s] reverse lookup for 6 word substrings
```

### with manual gc set to `false` (default)

```
./HipparchiaGoServer -gl 4 -el 0 -st
[HGS] HipparchiaGoServer (v2.1.0-pre) [git: d1ac12bf] [default.pgo] [gl=4; el=0]
	Built:	2026-02-16@20:26:55		Golang:	go1.26.0-X:jsonv2
	System:	darwin-arm64			WKvCPU:	20/20
[HGS-DBI] Number of stored vector models: 12
[HGS-MPS] MapNewWorkCorpus() added 837 works from 'lat'
[HGS-MPS] MapNewWorkCorpus() added 6625 works from 'tlg'
[HGS] [A1: 0.077s][Δ: 0.077s] 7462 works built: map[string]DbWork
[HGS-MPS] MapNewAuthorCorpus() added 363 authors from 'lat'
[HGS-MPS] MapNewAuthorCorpus() added 1823 authors from 'tlg'
[HGS] [A2: 0.085s][Δ: 0.007s] 2186 authors built: map[string]DbAuthor
[HGS] [A3: 0.088s][Δ: 0.003s] corpus maps built
[HGS] [A4: 0.090s][Δ: 0.005s] load word weights
[HGS] [B1: 0.158s][Δ: 0.158s] unnested lemma map built (158966 items)
[HGS] [B2: 0.202s][Δ: 0.043s] nested lemma map built
[HGS] initialization took 0.257s
[HGS] to stop the server press Control-C or close this window
[HGS] server started at http://127.0.0.1:8001 (tls unavailable)
[HGS-SELFTEST] Running Selftest 1 of 1
[HGS-SELFTEST] entering selftestsuite mode (4 segments)
[HGS-SELFTEST] [I] 6 search tests
[HGS-VLT] RtSearch() current heap: 89M
[HGS-SELFTEST] [A1: 0.876s][Δ: 0.876s] single word in corpus: 'vervex'
[HGS-VLT] RtSearch() current heap: 104M
[HGS-SELFTEST] [A2: 1.972s][Δ: 1.096s] phrase in corpus: 'plato omnem'
[HGS-SEA] [Δ: 1.352s]  WithinXLinesSearch(): 2308 initial hits
[HGS-SEA] [Δ: 0.004s]  SSBuildQueries() rerun
[HGS-SEA] [Δ: 0.023s]  WithinXLinesSearch(): 3 subsequent hits
[HGS-VLT] RtSearch() current heap: 102M
[HGS-SELFTEST] [A3: 3.365s][Δ: 1.393s] phrase near phrase: 'καὶ δὴ καὶ' near 'εἴ που καὶ'
[HGS-VLT] RtSearch() current heap: 113M
[HGS-SELFTEST] [B1: 4.348s][Δ: 0.983s] lemma in corpus: 'φθορώδηϲ'
[HGS-SEA] swapphraseandlemma() was called: lemmatized 'γαῖα' swapped with 'ἐϲχάτη χθονόϲ'
[HGS-SEA] [Δ: 1.369s]  WithinXLinesSearch(): 11 initial hits
[HGS-SEA] [Δ: 0.001s]  SSBuildQueries() rerun
[HGS-SEA] [Δ: 0.004s]  WithinXLinesSearch(): 4 subsequent hits
[HGS-VLT] RtSearch() current heap: 131M
[HGS-SELFTEST] [B2: 5.737s][Δ: 1.389s] lemma near phrase: 'γαῖα' near 'ἐϲχάτη χθονόϲ'
[HGS-SEA] pickfastestlemma() is NOT swapping πόλιϲ for ὁπλίζω: possible hits 205378 vs 3205; known forms 50 vs 192
[HGS-SEA] [Δ: 5.706s]  WithinXLinesSearch(): 99194 initial hits
[HGS-SEA] [Δ: 0.164s]  SSBuildQueries() rerun
[HGS-SEA] [Δ: 13.901s]  WithinXLinesSearch(): 101 subsequent hits
[HGS-VLT] RtSearch() current heap: 194M
[HGS-SELFTEST] [B3: 25.730s][Δ: 19.993s] lemma near lemma in corpus: 'πόλιϲ' near 'ὁπλίζω'
[HGS-SELFTEST] [II] 3 text, index, and vocab maker tests
[HGS-VLT] RtTextMaker() current heap: 203M
[HGS-SELFTEST] [C1: 26.166s][Δ: 0.436s] build a text for 40000 arbitrary lines
[HGS-VLT] RtIndexMaker() current heap: 170M
[HGS-SELFTEST] [C2: 26.433s][Δ: 0.267s] build an index to 40000 arbitrary lines
[HGS-DBI] ArrayToGetRequiredMorphObjects() will search among 191070 words
[HGS-VLT] RtVocabMaker() current heap: 609M
[HGS-SELFTEST] [C3: 30.596s][Δ: 4.164s] build vocabulary list for 40000 arbitrary lines
[HGS-SELFTEST] [III] 4 browsing and lexical tests
[HGS-VLT] RtBrowseLine() current heap: 624M
[HGS-VLT] RtBrowseLine() current heap: 625M
[HGS-VLT] RtBrowseLine() current heap: 627M
[HGS-VLT] RtBrowseLine() current heap: 627M
[HGS-VLT] RtBrowseLine() current heap: 628M
[HGS-VLT] RtBrowseLine() current heap: 630M
[HGS-VLT] RtBrowseLine() current heap: 632M
[HGS-VLT] RtBrowseLine() current heap: 633M
[HGS-VLT] RtBrowseLine() current heap: 635M
[HGS-VLT] RtBrowseLine() current heap: 637M
[HGS-VLT] RtBrowseLine() current heap: 638M
[HGS-WEB] could not find a work for tlg0021w001
[HGS-VLT] RtBrowseLine() current heap: 638M
[HGS-VLT] RtBrowseLine() current heap: 640M
[HGS-VLT] RtBrowseLine() current heap: 641M
[HGS-VLT] RtBrowseLine() current heap: 643M
[HGS-WEB] could not find a work for tlg0025w001
[HGS-VLT] RtBrowseLine() current heap: 643M
[HGS-VLT] RtBrowseLine() current heap: 644M
[HGS-VLT] RtBrowseLine() current heap: 646M
[HGS-VLT] RtBrowseLine() current heap: 648M
[HGS-VLT] RtBrowseLine() current heap: 648M
[HGS-VLT] RtBrowseLine() current heap: 648M
[HGS-VLT] RtBrowseLine() current heap: 650M
[HGS-VLT] RtBrowseLine() current heap: 651M
[HGS-VLT] RtBrowseLine() current heap: 653M
[HGS-VLT] RtBrowseLine() current heap: 654M
[HGS-VLT] RtBrowseLine() current heap: 654M
[HGS-VLT] RtBrowseLine() current heap: 655M
[HGS-VLT] RtBrowseLine() current heap: 656M
[HGS-VLT] RtBrowseLine() current heap: 656M
[HGS-VLT] RtBrowseLine() current heap: 657M
[HGS-VLT] RtBrowseLine() current heap: 657M
[HGS-VLT] RtBrowseLine() current heap: 658M
[HGS-VLT] RtBrowseLine() current heap: 658M
[HGS-VLT] RtBrowseLine() current heap: 658M
[HGS-VLT] RtBrowseLine() current heap: 658M
[HGS-VLT] RtBrowseLine() current heap: 658M
[HGS-VLT] RtBrowseLine() current heap: 659M
[HGS-VLT] RtBrowseLine() current heap: 659M
[HGS-VLT] RtBrowseLine() current heap: 660M
[HGS-VLT] RtBrowseLine() current heap: 660M
[HGS-VLT] RtBrowseLine() current heap: 660M
[HGS-VLT] RtBrowseLine() current heap: 660M
[HGS-VLT] RtBrowseLine() current heap: 660M
[HGS-VLT] RtBrowseLine() current heap: 662M
[HGS-VLT] RtBrowseLine() current heap: 664M
[HGS-VLT] RtBrowseLine() current heap: 665M
[HGS-VLT] RtBrowseLine() current heap: 665M
[HGS-VLT] RtBrowseLine() current heap: 668M
[HGS-VLT] RtBrowseLine() current heap: 669M
[HGS-VLT] RtBrowseLine() current heap: 671M
[HGS-SELFTEST] [D1: 30.811s][Δ: 0.215s] browse 50 passages
[HGS-VLT] RtLexFindByForm() current heap: 672M
[HGS-VLT] RtLexFindByForm() current heap: 672M
[HGS-VLT] RtLexFindByForm() current heap: 673M
[HGS-VLT] RtLexFindByForm() current heap: 673M
[HGS-VLT] RtLexFindByForm() current heap: 674M
[HGS-VLT] RtLexFindByForm() current heap: 678M
[HGS-VLT] RtLexFindByForm() current heap: 679M
[HGS-VLT] RtLexFindByForm() current heap: 680M
[HGS-VLT] RtLexFindByForm() current heap: 680M
[HGS-VLT] RtLexFindByForm() current heap: 683M
[HGS-VLT] RtLexFindByForm() current heap: 684M
[HGS-VLT] RtLexFindByForm() current heap: 684M
[HGS-VLT] RtLexFindByForm() current heap: 686M
[HGS-VLT] RtLexFindByForm() current heap: 686M
[HGS-VLT] RtLexFindByForm() current heap: 691M
[HGS-VLT] RtLexFindByForm() current heap: 692M
[HGS-VLT] RtLexFindByForm() current heap: 692M
[HGS-VLT] RtLexFindByForm() current heap: 692M
[HGS-VLT] RtLexFindByForm() current heap: 693M
[HGS-VLT] RtLexFindByForm() current heap: 693M
[HGS-VLT] RtLexFindByForm() current heap: 694M
[HGS-VLT] RtLexFindByForm() current heap: 695M
[HGS-VLT] RtLexFindByForm() current heap: 696M
[HGS-VLT] RtLexFindByForm() current heap: 696M
[HGS-VLT] RtLexFindByForm() current heap: 697M
[HGS-VLT] RtLexFindByForm() current heap: 698M
[HGS-VLT] RtLexFindByForm() current heap: 699M
[HGS-VLT] RtLexFindByForm() current heap: 488M
[HGS-VLT] RtLexFindByForm() current heap: 122M
[HGS-VLT] RtLexFindByForm() current heap: 122M
[HGS-VLT] RtLexFindByForm() current heap: 123M
[HGS-VLT] RtLexFindByForm() current heap: 124M
[HGS-VLT] RtLexFindByForm() current heap: 124M
[HGS-VLT] RtLexFindByForm() current heap: 125M
[HGS-VLT] RtLexFindByForm() current heap: 125M
[HGS-VLT] RtLexFindByForm() current heap: 126M
[HGS-VLT] RtLexFindByForm() current heap: 127M
[HGS-VLT] RtLexFindByForm() current heap: 129M
[HGS-VLT] RtLexFindByForm() current heap: 129M
[HGS-VLT] RtLexFindByForm() current heap: 130M
[HGS-VLT] RtLexFindByForm() current heap: 130M
[HGS-VLT] RtLexFindByForm() current heap: 131M
[HGS-VLT] RtLexFindByForm() current heap: 131M
[HGS-VLT] RtLexFindByForm() current heap: 132M
[HGS-VLT] RtLexFindByForm() current heap: 133M
[HGS-VLT] RtLexFindByForm() current heap: 137M
[HGS-VLT] RtLexFindByForm() current heap: 138M
[HGS-VLT] RtLexFindByForm() current heap: 138M
[HGS-SELFTEST] [D2: 33.536s][Δ: 2.725s] look up 48 specific words
[HGS-VLT] RtLexLookup() current heap: 145M
[HGS-DBI] FindProximateEntry() found no entry before '0.000000'
[HGS-VLT] RtLexLookup() current heap: 149M
[HGS-VLT] RtLexLookup() current heap: 154M
[HGS-VLT] RtLexLookup() current heap: 157M
[HGS-VLT] RtLexLookup() current heap: 171M
[HGS-VLT] RtLexLookup() current heap: 181M
[HGS-VLT] RtLexLookup() current heap: 187M
[HGS-DBI] FindProximateEntry() found no entry before '0.000000'
[HGS-VLT] RtLexLookup() current heap: 191M
[HGS-VLT] RtLexLookup() current heap: 196M
[HGS-VLT] RtLexLookup() current heap: 199M
[HGS-VLT] RtLexLookup() current heap: 213M
[HGS-VLT] RtLexLookup() current heap: 223M
[HGS-VLT] RtLexLookup() current heap: 230M
[HGS-DBI] FindProximateEntry() found no entry before '0.000000'
[HGS-VLT] RtLexLookup() current heap: 131M
[HGS-VLT] RtLexLookup() current heap: 136M
[HGS-VLT] RtLexLookup() current heap: 138M
[HGS-VLT] RtLexLookup() current heap: 151M
[HGS-VLT] RtLexLookup() current heap: 161M
[HGS-SELFTEST] [D3: 34.614s][Δ: 1.078s] look up 18 word substrings
[HGS-VLT] RtLexReverse() current heap: 209M
[HGS-VLT] RtLexReverse() current heap: 247M
[HGS-VLT] RtLexReverse() current heap: 153M
[HGS-VLT] RtLexReverse() current heap: 206M
[HGS-VLT] RtLexReverse() current heap: 250M
[HGS-VLT] RtLexReverse() current heap: 153M
[HGS-SELFTEST] [D4: 35.927s][Δ: 1.313s] reverse lookup for 6 word substrings
```

## workflow

![workflow](../gitimg/hipparchia_workflow.svg)

## code stats

```
 % cloc --exclude-dir=z --not-match-f="^jq*" --not-match-f="xml" .
     205 text files.
     190 unique files.
     288 files ignored.

github.com/AlDanial/cloc v 2.08  T=0.19 s (987.6 files/s, 187474.6 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                             124           3895           4463          17901
CSS                             13            477            144           2013
HTML                            13            162             18           1589
JavaScript                       8            323            218           1585
Markdown                        10            447              0           1150
JSON                             9              0              0            498
Text                             6            108              0            487
SVG                              1              1              1            392
Bourne Shell                     3             25              7             89
XML                              1              0              0              9
Python                           1              5              6              7
YAML                             1              8             34              4
-------------------------------------------------------------------------------
SUM:                           190           5451           4891          25724
-------------------------------------------------------------------------------

```