
# HipparchiaGoServer FYI

## CLI

![options](../gitimg/hgscli.png)

## self-test

``` 
[HGS] Hipparchia Golang Server (v1.1.6-pre) [git: 3cc8a3df] [gl=0; el=0]
[HGS] Running Selftest 1 of 1
	Built:	2023-03-07@12:43:44
	Go:	go1.20.1
⇨ http server started on 127.0.0.1:8000
[HGS] entering selftest mode (3 segments)
[HGS] [I] 6 search tests
[HGS] [A1: 0.222s][Δ: 0.222s] single word in corpus: 'vervex'
[HGS] [A2: 1.459s][Δ: 1.237s] phrase in corpus: 'plato omnem'
[HGS] [A3: 3.325s][Δ: 1.865s] phrase near phrase: 'καὶ δὴ καὶ' near 'εἴ που καὶ'
[HGS] [B1: 4.385s][Δ: 1.061s] lemma in corpus: 'φθορώδηϲ'
[HGS] [B2: 9.051s][Δ: 4.666s] lemma near phrase: 'γαῖα' near 'ἐϲχάτη χθονόϲ'
[HGS] [B3: 33.214s][Δ: 24.163s] lemma near lemma in corpus: 'Πόλιϲ' near 'ὁπλίζω
[HGS] [II] 3 text, index, and vocab maker tests
[HGS] [C1: 33.451s][Δ: 0.237s] build a text for 35000 arbitrary lines
[HGS] [C2: 36.849s][Δ: 3.397s] build an index to 35000 arbitrary lines
[HGS] [C3: 39.330s][Δ: 2.481s] build vocabulary list for 35000 arbitrary lines
[HGS] [III] 4 browsing and lexical tests
[HGS] could not find a work for gr0021w001
[HGS] could not find a work for gr0025w001
[HGS] [D1: 39.544s][Δ: 0.215s] browse 50 passages
[HGS] findbyform() found no results for 'Romani'
[HGS] [D2: 44.120s][Δ: 4.576s] look up 48 specific words
[HGS] [D3: 49.347s][Δ: 5.227s] look up 6 word substrings
[HGS] [D4: 64.088s][Δ: 14.741s] reverse lookup for 6 word substrings
[HGS] exiting selftest mode
```

### some selftest times

* 91s on 8 cores of apple silicon (m1) virtualizing rocky linux
* 111s on 6 cores of apple silicon (m1) virtualizing rocky linux
* 113s on 6 cores of an intel 9900k running a virtualized ubuntu
* 147s on 6 cores of apple silicon (m1) virtualizing freebsd13 w/ zfs
* 154s on 6 cores of apple silicon (m1) virtualizing macos
* 232s on 6 cores of an intel 9900k running a virtualized windows 10
* 349s on a mac mini 2018
* 483s on 6 cores of apple silicon (m1) virtualizing windows 11 arm
* 1144s (ouch) on a 2017 MacBook w/ 1.3GHz Core i5

### self-test: cpu-constrained vs i/o constrained

![workers vs time](../gitimg/workers_vs_time.png)

## profiling

`HipparchiaGoServer -st -st -st` called to generate
* `default.pgo`
* `MEMProfile.pdf`
* `CPUProfile.pdf` 

## memory use during self-test

``` 
[HGS] Hipparchia Golang Server (v1.1.6-pre) [git: 3cc8a3df] [gl=4; el=0]
[HGS] [B1: 0.198s][Δ: 0.198s] unnested lemma map built (152382 items)
[HGS] [A1: 0.264s][Δ: 0.264s] 236835 works built: map[string]DbWork
[HGS] [B2: 0.273s][Δ: 0.075s] nested lemma map built
[HGS] [A2: 0.302s][Δ: 0.037s] 3455 authors built: map[string]DbAuthor
[HGS] [A3: 0.374s][Δ: 0.073s] corpus maps built
[HGS] main() post-initialization runtime.GC() 239M --> 206M
[HGS] initialization took 0.395s before reaching StartEchoServer()
[HGS] Running Selftest 1 of 1
⇨ http server started on 127.0.0.1:8000
	Built:	2023-03-07@12:33:40
	Go:	devel go1.21-84609d874e Mon Mar 6 23:46:08 2023 +0000
[HGS] entering selftest mode (3 segments)
[HGS] [I] 6 search tests
[HGS] RtSearch() runtime.GC() 217M --> 210M
[HGS] [A1: 0.308s][Δ: 0.308s] single word in corpus: 'vervex'
[HGS] RtSearch() runtime.GC() 224M --> 215M
[HGS] [A2: 1.598s][Δ: 1.290s] phrase in corpus: 'plato omnem'
[HGS] [Δ: 1.360s]  WithinXLinesSearch(): 2307 initial hits
[HGS] [Δ: 0.003s]  SSBuildQueries() rerun
[HGS] [Δ: 0.023s]  WithinXLinesSearch(): 3 subsequent hits
[HGS] RtSearch() runtime.GC() 261M --> 219M
[HGS] [A3: 3.105s][Δ: 1.507s] phrase near phrase: 'καὶ δὴ καὶ' near 'εἴ που καὶ'
[HGS] RtSearch() runtime.GC() 231M --> 222M
[HGS] [B1: 4.504s][Δ: 1.399s] lemma in corpus: 'φθορώδηϲ'
[HGS] [Δ: 3.800s]  WithinXLinesSearch(): 86256 initial hits
[HGS] [Δ: 0.084s]  SSBuildQueries() rerun
[HGS] [Δ: 0.263s]  WithinXLinesSearch(): 4 subsequent hits
[HGS] RtSearch() runtime.GC() 386M --> 232M
[HGS] [B2: 8.799s][Δ: 4.295s] lemma near phrase: 'γαῖα' near 'ἐϲχάτη χθονόϲ'
[HGS] [Δ: 7.254s]  WithinXLinesSearch(): 99300 initial hits
[HGS] [Δ: 0.164s]  SSBuildQueries() rerun
[HGS] [Δ: 16.444s]  WithinXLinesSearch(): 101 subsequent hits
[HGS] RtSearch() runtime.GC() 535M --> 267M
[HGS] [B3: 33.207s][Δ: 24.409s] lemma near lemma in corpus: 'Πόλιϲ' near 'ὁπλίζω
[HGS] [II] 3 text, index, and vocab maker tests
[HGS] [C1: 33.474s][Δ: 0.267s] build a text for 35000 arbitrary lines
[HGS] [C2: 40.796s][Δ: 7.322s] build an index to 35000 arbitrary lines
[HGS] [C3: 43.873s][Δ: 3.077s] build vocabulary list for 35000 arbitrary lines
[HGS] [III] 4 browsing and lexical tests
[HGS] could not find a work for gr0021w001
[HGS] could not find a work for gr0025w001
[HGS] [D1: 44.085s][Δ: 0.212s] browse 50 passages
[HGS] findbyform() found no results for 'Romani'
[HGS] [D2: 48.748s][Δ: 4.663s] look up 48 specific words
[HGS] RtLexLookup() runtime.GC() 805M --> 398M
[HGS] RtLexLookup() runtime.GC() 403M --> 394M
[HGS] RtLexLookup() runtime.GC() 401M --> 394M
[HGS] RtLexLookup() runtime.GC() 398M --> 395M
[HGS] RtLexLookup() runtime.GC() 415M --> 400M
[HGS] [D3: 54.151s][Δ: 5.403s] look up 6 word substrings
[HGS] RtLexLookup() runtime.GC() 411M --> 399M
[HGS] RtLexReverse() runtime.GC() 458M --> 406M
[HGS] RtLexReverse() runtime.GC() 464M --> 413M
[HGS] RtLexReverse() runtime.GC() 426M --> 408M
[HGS] RtLexReverse() runtime.GC() 487M --> 417M
[HGS] RtLexReverse() runtime.GC() 463M --> 419M
[HGS] [D4: 69.234s][Δ: 15.083s] reverse lookup for 6 word substrings
[HGS] exiting selftest mode
[HGS] RtLexReverse() runtime.GC() 452M --> 414M
```

## workflow

![workflow](../gitimg/hipparchia_workflow.svg)

## code stats

```
% cloc *go
      37 text files.
      37 unique files.                              
       1 file ignored.

github.com/AlDanial/cloc v 1.96  T=0.03 s (1116.0 files/s, 445406.1 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              37           2223           2007          10537
-------------------------------------------------------------------------------
SUM:                            37           2223           2007          10537
-------------------------------------------------------------------------------

```