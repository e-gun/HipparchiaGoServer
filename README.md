# HipparchiaGoServer

### STATUS (`v1.0.7`):

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

* room to improve result highlighting

### DISTANT FANTASIES
* vectors


![options](gitimg/hgscli.png)

![workflow](gitimg/hipparchia_workflow.svg)

```
% cloc *go

      29 text files.
      29 unique files.                              
       0 files ignored.

github.com/AlDanial/cloc v 1.94  T=0.03 s (1029.8 files/s, 413514.4 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              29           1729           1635           8281
-------------------------------------------------------------------------------
SUM:                            29           1729           1635           8281
-------------------------------------------------------------------------------

```