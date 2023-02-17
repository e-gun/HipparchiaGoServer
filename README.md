# HipparchiaGoServer

### STATUS (v1.1.3):

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

### INSTRUCTIONS
* for installation, see the `INSTRUCTIONS` folder
* for UI items, hover to get quick tips
* click on `?` at bottom corner for detailed help
* see `antisigma` for video demos

### NOTES

* c. 325MB resident RAM on launch (note that a simple application like iTerm2 uses 354MB and Firefox uses 855MB)
* add about 125MB of RAM per postgres client connection in the pool
* a very heavy search like "all forms of Πόλιϲ near..." can kick you up into 660MB

### MINIMUM VERSION OF GOLANG REQUIRED TO BUILD
* go 1.19
* `searchstructs.go` uses `atomic.Int64` from `sync/atomic`. `atomic.Int64` was added in go 1.19.
