# HipparchiaGoServer

### STATUS (`v1.0.17`):

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

### SOME DAY

* vectors

### NOTES

* c. 325MB resident RAM on launch (note that a simple application like iTerm2 uses 354MB and Firefox uses 855MB)
* add about 125MB of RAM per postgres client connection in the pool
* a very heavy search like "all forms of Πόλιϲ near..." can kick you up into 660MB

### MINIMUM VERSION OF GOLANG REQUIRED TO BUILD
* go 1.19
* `searchstructs.go` uses `atomic.Int64` from `sync/atomic`. `atomic.Int64` was added in go 1.19.

![options](gitimg/hgscli.png)

![selftest](gitimg/selftest.png)

![workflow](gitimg/hipparchia_workflow.svg)

```
% cloc *go

      32 text files.
      32 unique files.                              
       0 files ignored.

github.com/AlDanial/cloc v 1.96  T=0.03 s (1130.1 files/s, 443933.3 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              32           1884           1727           8959
-------------------------------------------------------------------------------
SUM:                            32           1884           1727           8959
-------------------------------------------------------------------------------

```