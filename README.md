# HipparchiaGoServer

### STATUS (`v1.0.5`):

* 25%-700% faster than HipparchiaServer depending on the function. Uses c. 60% as much RAM.
* monolithic binary: no need for extra files/folders beyond setting a password in `hgs-conf.json`
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

* WithinXWordsSearch() needs triple-checking
* room to improve lemma highlighting

### DISTANT FANTASIES
* vectors


![options](gitimg/hgscli.png)

![workflow](gitimg/hipparchia_workflow.svg)

```
% cloc *go

      27 text files.
      27 unique files.                              
       0 files ignored.

github.com/AlDanial/cloc v 1.94  T=0.03 s (998.6 files/s, 444320.3 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              27           1770           1990           8254
-------------------------------------------------------------------------------
SUM:                            27           1770           1990           8254
-------------------------------------------------------------------------------

```