# HipparchiaGoServer

### STATUS (`v1.0.4`):

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
* multi-line span formatting needs more work

### DISTANT FANTASIES
* vectors


![options](gitimg/hgscli.png)

![workflow](gitimg/hipparchia_workflow.svg)

```
 % cloc *go
      26 text files.
      26 unique files.                              
       0 files ignored.

github.com/AlDanial/cloc v 1.94  T=0.03 s (1013.0 files/s, 449816.9 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              26           1714           1948           7883
-------------------------------------------------------------------------------
SUM:                            26           1714           1948           7883
-------------------------------------------------------------------------------

```