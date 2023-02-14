# HipparchiaGoServer

### STATUS (v1.1.1):

* 25%-700% faster than HipparchiaServer depending on the function. Uses c. 60% as much RAM.
* monolithic binary: no need for extra files/folders beyond setting a password in `hgs-conf.json`

### FEATURES

* searching
* browsing 
* selections 
* exclusions 
* not-near syntax
* progress polling 
* dictionaries
* text maker
* indexing
* vocabulary list generation
* morphology tables
* self-configuration first run
* self-installation on first run 

### SOME DAY

* vectors

### NOTES

* c. 325MB resident RAM on launch (note that a simple application like iTerm2 uses 354MB and Firefox uses 855MB)
* add about 125MB of RAM per postgres client connection in the pool
* a very heavy search like "all forms of Πόλιϲ near..." can kick you up into 660MB

### MINIMUM VERSION OF GOLANG REQUIRED TO BUILD
* go 1.19
* `searchstructs.go` uses `atomic.Int64` from `sync/atomic`. `atomic.Int64` was added in go 1.19.

### FYI

![options](gitimg/hgscli.png)

![selftest](gitimg/selftest.png)
#### some selftest times
* 64s on a mac studio
* 113s on 6 cores of an intel 9900k running a virtualized ubuntu
* 154s on 6 cores of apple silicon (m1) virtualizing macos
* 232s on 6 cores of an intel 9900k running a virtualized windows 10
* 349s on a mac mini 2018
* 483s on 6 cores of apple silicon (m1) virtualizing windows 11 arm
* 1144s (ouch) on a 2017 MacBook w/ 1.3GHz Core i5

![workflow](gitimg/hipparchia_workflow.svg)


```
% cloc *go

      34 text files.
      34 unique files.                              
       0 files ignored.

github.com/AlDanial/cloc v 1.96  T=0.03 s (1038.5 files/s, 395145.0 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              34           1952           1776           9209
-------------------------------------------------------------------------------
SUM:                            34           1952           1776           9209
-------------------------------------------------------------------------------

```