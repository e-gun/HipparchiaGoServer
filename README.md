# HipparchiaGoServer

### STATUS (`v1.0.3`):

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

* WithinXWordsSearch() needs to have more corner cases checked
* room to improve lemma highlighting

### DISTANT FANTASIES
* vectors


![options](gitimg/hgscli.png)

![workflow](gitimg/hipparchia_workflow.svg)

```
 % cloc *go
      25 text files.
      25 unique files.
       0 files ignored.

github.com/AlDanial/cloc v 1.94  T=0.03 s (1032.0 files/s, 448622.8 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              26           1710           1923           7851
-------------------------------------------------------------------------------
SUM:                            26           1710           1923           7851
-------------------------------------------------------------------------------


```