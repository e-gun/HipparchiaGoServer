# HipparchiaGoServer
## beta quality: (allegedly) solid core

### STATUS (`v0.9.7`):

* now feature-complete for `1.0.0` reslease
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

* morphology tables need lots of polish; dialects need addressing
* WithinXWordsSearch() needs to have more corner cases checked
* room to improve lemma highlighting

### DISTANT FANTASIES
* vectors


![options](gitimg/hgscli.png)

![workflow](gitimg/hipparchia_workflow.svg)