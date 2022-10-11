# HipparchiaGoServer
## beta quality: (allegedly) solid core

### STATUS (`v1.0.0`):

* 25%-700% faster than HipparchiaServer depending on the function. Uses c. 60% as much RAM.
* monolithic binary: no need for extra files/folders beyond setting a password in `hgs-conf.json`
* searching seems to be fine (but more testing/checking would not hurt)
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

github.com/AlDanial/cloc v 1.94  T=0.02 s (1023.3 files/s, 444254.5 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              25           1659           1661           7534
-------------------------------------------------------------------------------
SUM:                            25           1659           1661           7534
-------------------------------------------------------------------------------
```