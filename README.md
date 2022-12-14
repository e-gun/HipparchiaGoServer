# HipparchiaGoServer

### STATUS (`v1.0.11`):

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

### TODO (in vague order of priority)

* room to improve result highlighting

### DISTANT FANTASIES

* vectors

### NOTES

* c. 325MB resident RAM on launch (note that a simple application like iTerm2 uses 354MB and Firefox uses 855MB)
* add about 125MB of RAM per postgres client connection in the pool
* a very heavy search like "all forms of Πόλιϲ near..." can kick you up into 660MB

![options](gitimg/hgscli.png)

![workflow](gitimg/hipparchia_workflow.svg)

```
% cloc *go

      30 text files.
      30 unique files.                              
       0 files ignored.

github.com/AlDanial/cloc v 1.96  T=0.03 s (960.4 files/s, 387851.6 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              30           1824           1698           8593
-------------------------------------------------------------------------------
SUM:                            30           1824           1698           8593
-------------------------------------------------------------------------------

```