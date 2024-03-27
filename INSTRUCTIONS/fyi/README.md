
# HipparchiaGoServer FYI

## CLI

![options](../gitimg/hgscli.png)

## self-test

self-test without vectors is now `HipparchiaGoServer -st -dv`

``` 
% /Users/erik/Applications/net/HipparchiaGoServer -st -wc 20
[HGS] Hipparchia Golang Server (v1.3.1-pre) [git: ed35669c] [default.pgo] [gl=3; el=0]
	Built:	2024-03-22@14:20:00		Golang:	go1.22.1
	System:	darwin-arm64			WKvCPU:	20/20
[HGS-VEC] Number of stored vector models: 12
[HGS] [A1: 0.154s][Δ: 0.154s] 7461 works built: map[string]DbWork
[HGS] [A2: 0.166s][Δ: 0.011s] 2185 authors built: map[string]DbAuthor
[HGS] [A3: 0.168s][Δ: 0.002s] corpus maps built
[HGS] [B1: 0.205s][Δ: 0.205s] unnested lemma map built (158817 items)
[HGS] [B2: 0.266s][Δ: 0.061s] nested lemma map built
[HGS] initialization took 0.280s
[HGS] to stop the server press Control-C or close this window
[HGS-SELFTEST] Running Selftest 1 of 1
[HGS-SELFTEST] entering selftestsuite mode (4 segments)
[HGS-SELFTEST] [I] 6 search tests
⇨ http server started on 127.0.0.1:8000
[HGS-SELFTEST] [A1: 0.782s][Δ: 0.782s] single word in corpus: 'vervex'
[HGS-SELFTEST] [A2: 1.842s][Δ: 1.060s] phrase in corpus: 'plato omnem'
[HGS-SELFTEST] [A3: 3.344s][Δ: 1.502s] phrase near phrase: 'καὶ δὴ καὶ' near 'εἴ που καὶ'
[HGS-SELFTEST] [B1: 4.570s][Δ: 1.226s] lemma in corpus: 'φθορώδηϲ'
[HGS-SELFTEST] [B2: 6.034s][Δ: 1.464s] lemma near phrase: 'γαῖα' near 'ἐϲχάτη χθονόϲ'
[HGS-SELFTEST] [B3: 27.734s][Δ: 21.700s] lemma near lemma in corpus: 'πόλιϲ' near 'ὁπλίζω'
[HGS-SELFTEST] [II] 3 text, index, and vocab maker tests
[HGS-SELFTEST] [C1: 27.986s][Δ: 0.251s] build a text for 35000 arbitrary lines
[HGS-SELFTEST] [C2: 29.176s][Δ: 1.190s] build an index to 35000 arbitrary lines
[HGS-SELFTEST] [C3: 32.071s][Δ: 2.895s] build vocabulary list for 35000 arbitrary lines
[HGS-SELFTEST] [III] 4 browsing and lexical tests
[HGS-WEB] could not find a work for gr0021w001
[HGS-WEB] could not find a work for gr0025w001
[HGS-SELFTEST] [D1: 32.236s][Δ: 0.165s] browse 50 passages
[HGS-WEB] findbyform() found no results for 'Romani'
[HGS-SELFTEST] [D2: 36.343s][Δ: 4.107s] look up 48 specific words
[HGS-SELFTEST] [D3: 50.825s][Δ: 14.482s] look up 18 word substrings
[HGS-SELFTEST] [D4: 62.966s][Δ: 12.141s] reverse lookup for 6 word substrings
[HGS-VEC] VectorDBReset() dropped semantic_vectors_nn
[HGS-SELFTEST] [IV] nearest neighbor vectorization tests
[HGS-VEC] VectorDBInitNN(): success
[HGS-SELFTEST] [E1: 84.478s][Δ: 21.513s] semantic vector model test: w2v - 1 author(s) with 4 text preparation modes per author
[HGS-SELFTEST] [E2: 107.762s][Δ: 23.284s] semantic vector model test: lexvec - 1 author(s) with 4 text preparation modes per author
[HGS-SELFTEST] [E3: 146.166s][Δ: 38.404s] semantic vector model test: glove - 1 author(s) with 4 text preparation modes per author
[HGS-SELFTEST] [V] lda vectorization tests
[HGS-SELFTEST] [F: 165.912s][Δ: 19.745s] lda vector model test - 1 author(s) with 4 text preparation modes per author
[HGS-SELFTEST] exiting selftestsuite mode

```

self-test with vectors can be deceptive because `-wc` flag will not override config json.

### some vectorless selftest times

* 91s on 8 cores of apple silicon (m1) virtualizing rocky linux
* 111s on 6 cores of apple silicon (m1) virtualizing rocky linux
* 113s on 6 cores of an intel 9900k running a virtualized ubuntu
* 147s on 6 cores of apple silicon (m1) virtualizing freebsd13 w/ zfs
* 101 on 6 cores of apple silicon (m1) virtualizing macos
* 232s on 6 cores of an intel 9900k running a virtualized windows 10
* 349s on a mac mini 2018
* 483s on 6 cores of apple silicon (m1) virtualizing windows 11 arm
* 1144s (ouch) on a 2017 MacBook w/ 1.3GHz Core i5

### self-test: cpu-constrained vs i/o constrained

![workers vs time](../gitimg/workers_vs_time.png)

## profiling

`HipparchiaGoServer -st` called to generate
* `default.pgo`
* `MEMProfile.pdf`
* `CPUProfile.pdf`

## memory use during self-test

``` 
% /Users/erik/Applications/net/HipparchiaGoServer -st -wc 20 -gl 4
[HGS] Hipparchia Golang Server (v1.3.1-pre) [git: ed35669c] [default.pgo] [gl=4; el=0]
	Built:	2024-03-22@14:20:00		Golang:	go1.22.1
	System:	darwin-arm64			WKvCPU:	20/20
[HGS-VEC] Number of stored vector models: 12
[HGS-MPS] MapNewWorkCorpus() added 6625 works from 'gr'
[HGS-MPS] MapNewWorkCorpus() added 836 works from 'lt'
[HGS] [A1: 0.152s][Δ: 0.152s] 7461 works built: map[string]DbWork
[HGS-MPS] MapNewAuthorCorpus() added 1823 authors from 'gr'
[HGS-MPS] MapNewAuthorCorpus() added 362 authors from 'lt'
[HGS] [A2: 0.160s][Δ: 0.008s] 2185 authors built: map[string]DbAuthor
[HGS] [A3: 0.162s][Δ: 0.002s] corpus maps built
[HGS] [B1: 0.210s][Δ: 0.210s] unnested lemma map built (158817 items)
[HGS] [B2: 0.269s][Δ: 0.059s] nested lemma map built
[HGS] main() post-initialization current heap: 80M
[HGS] initialization took 0.309s
[HGS] to stop the server press Control-C or close this window
[HGS-SELFTEST] Running Selftest 1 of 1
[HGS-SELFTEST] entering selftestsuite mode (4 segments)
[HGS-SELFTEST] [I] 6 search tests
⇨ http server started on 127.0.0.1:8000
[HGS-WEB] RtSearch() current heap: 92M
[HGS-SELFTEST] [A1: 0.700s][Δ: 0.700s] single word in corpus: 'vervex'
[HGS-WEB] RtSearch() current heap: 106M
[HGS-SELFTEST] [A2: 1.992s][Δ: 1.292s] phrase in corpus: 'plato omnem'
[HGS-SEA] [Δ: 1.337s]  WithinXLinesSearch(): 2307 initial hits
[HGS-SEA] [Δ: 0.004s]  SSBuildQueries() rerun
[HGS-SEA] [Δ: 0.033s]  WithinXLinesSearch(): 3 subsequent hits
[HGS-WEB] RtSearch() current heap: 103M
[HGS-SELFTEST] [A3: 3.391s][Δ: 1.400s] phrase near phrase: 'καὶ δὴ καὶ' near 'εἴ που καὶ'
[HGS-WEB] RtSearch() current heap: 115M
[HGS-SELFTEST] [B1: 4.457s][Δ: 1.066s] lemma in corpus: 'φθορώδηϲ'
[HGS-STR] SwapPhraseAndLemma() was called: lemmatized 'γαῖα' swapped with 'ἐϲχάτη χθονόϲ'
[HGS-SEA] [Δ: 1.584s]  WithinXLinesSearch(): 11 initial hits
[HGS-SEA] [Δ: 0.002s]  SSBuildQueries() rerun
[HGS-SEA] [Δ: 0.004s]  WithinXLinesSearch(): 4 subsequent hits
[HGS-WEB] RtSearch() current heap: 132M
[HGS-SELFTEST] [B2: 6.071s][Δ: 1.614s] lemma near phrase: 'γαῖα' near 'ἐϲχάτη χθονόϲ'
[HGS-SEA] PickFastestLemma() is NOT swapping πόλιϲ for ὁπλίζω: possible hits 125274 vs 2547; known forms 50 vs 191
[HGS-SEA] [Δ: 6.100s]  WithinXLinesSearch(): 99350 initial hits
[HGS-SEA] [Δ: 0.189s]  SSBuildQueries() rerun
[HGS-SEA] [Δ: 14.943s]  WithinXLinesSearch(): 101 subsequent hits
[HGS-WEB] RtSearch() current heap: 196M
[HGS-SELFTEST] [B3: 27.552s][Δ: 21.481s] lemma near lemma in corpus: 'πόλιϲ' near 'ὁπλίζω'
[HGS-SELFTEST] [II] 3 text, index, and vocab maker tests
[HGS-WEB] RtTextMaker() current heap: 338M
[HGS-SELFTEST] [C1: 27.757s][Δ: 0.205s] build a text for 35000 arbitrary lines
[HGS-WEB] RtIndexMaker() current heap: 302M
[HGS-SELFTEST] [C2: 28.961s][Δ: 1.204s] build an index to 35000 arbitrary lines
[HGS-DBI] ArrayToGetRequiredMorphObjects() will search among 153612 words
[HGS-WEB] RtVocabMaker() current heap: 432M
[HGS-SELFTEST] [C3: 31.379s][Δ: 2.418s] build vocabulary list for 35000 arbitrary lines
[HGS-SELFTEST] [III] 4 browsing and lexical tests
...
[HGS-WEB] RtBrowseLine() current heap: 221M
[HGS-SELFTEST] [D1: 31.557s][Δ: 0.178s] browse 50 passages
...
[HGS-WEB] RtLexFindByForm() current heap: 264M
[HGS-SELFTEST] [D2: 35.687s][Δ: 4.130s] look up 48 specific words
...
[HGS-WEB] RtLexLookup() current heap: 190M
[HGS-SELFTEST] [D3: 50.739s][Δ: 15.052s] look up 18 word substrings
...
[HGS-WEB] RtLexReverse() current heap: 285M
[HGS-SELFTEST] [D4: 62.603s][Δ: 11.864s] reverse lookup for 6 word substrings
...
```

## workflow

![workflow](../gitimg/hipparchia_workflow.svg)

## code stats

```
cloc --exclude-dir=z --not-match-f="^jq*" .
     164 text files.
     148 unique files.                                          
     190 files ignored.

github.com/AlDanial/cloc v 2.00  T=0.12 s (1270.0 files/s, 257706.0 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              95           3113           3524          14277
JavaScript                       8            310            193           1595
CSS                              1            349             89           1567
HTML                            12            160             18           1532
Markdown                        11            452              0           1198
Text                             5             90              0            412
SVG                              1              1              1            392
JSON                             9              0              0            386
XML                              4              0              0            272
Bourne Shell                     1             13              7             63
Python                           1              5              6              7
-------------------------------------------------------------------------------
SUM:                           148           4493           3838          21701
-------------------------------------------------------------------------------
```