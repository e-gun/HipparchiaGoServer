# HipparchiaGoServer

### STATUS (v1.2.1):

* 25%-700% faster than HipparchiaServer depending on the function. Uses c. 60% as much RAM.
* monolithic binary: no need for extra files/folders

### FEATURES

* self-configuration first run
* self-installation on first run 
* searching
  * search scope selections 
  * search scope exclusions 
  * near/not-near syntax
  * progress polling
* semantic vectors
  * search for terms and yield an interactive graph of nearest neighbors
  * choice of graphing styles: "neighbors" or "neighbors + neighbors of neighbors"
  * choice of modeling algorithms
  * configurable model parameters
  * configurable lists of stop-words
* text browsing
* text maker
* indexing (+ flag unique forms)
* vocabulary list generation (+ flag unique headwords)
* morphology tables
* dictionaries

### INSTRUCTIONS
* for installation, see the `INSTRUCTIONS` folder
* when running: hover over UI elements to get quick tips
* when running: click on `?` at bottom corner for detailed help
* see also `antisigma` for video demos

### NOTES

* c. 325MB resident RAM on launch (note that a simple application like iTerm2 uses 354MB and Firefox uses 855MB)
* add about 125MB of RAM per postgres client connection in the pool
* a very heavy search like "all forms of Πόλιϲ near..." can kick you up into 660MB

### MINIMUM VERSION OF GOLANG REQUIRED TO BUILD
* go 1.19 because `atomic.Int32` and `atomic.Int64` are called
