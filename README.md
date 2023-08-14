# HipparchiaGoServer

### STATUS (v1.2.10):

* 25%-700% faster than HipparchiaServer depending on the function. Uses c. 60% as much RAM.
* monolithic binary: no need for extra files/folders

### FEATURES

* default configuration files generated on first run
* database self-installation on first run
* embedded help files
* text browsing
* text maker
* indexing (+ flag unique forms)
* vocabulary list generation (+ flag unique headwords)
* morphology tables
* dictionaries
* searching
  * search scope selections 
  * search scope exclusions 
  * near/not-near syntax
  * progress polling
* semantic vectors 
  * general
    * configurable model parameters
    * configurable lists of stop-words
    * choice of text building strategies
  * nearest neighbors
    * search for terms and yield a table and an interactive graph of nearest neighbors
    * choice of graphing styles: "neighbors" or "neighbors + neighbors of neighbors"
    * configure how many neighbors to track
    * choice of modeling algorithms
  * topic models
    * configure number of topics to track
    * option to build an interactive graph the LDA model via t-SNE reduction (graphs can be 2d or 3d)

### INSTRUCTIONS
* for installation, see the `INSTRUCTIONS` folder
* when running: hover over UI elements to get quick tips
* when running: click on `?` at bottom corner for detailed help
* see also `antisigma` for video demos

### NOTES

* c. 325MB resident RAM on launch (note that a simple application like iTerm2 uses 354MB and Firefox uses 855MB)
* add about 125MB of RAM per postgres client connection in the pool
* a very heavy search like "all forms of πόλιϲ near..." can kick you up into 660MB

### MINIMUM VERSION OF GOLANG REQUIRED TO BUILD
* go 1.19 because `atomic.Int32` and `atomic.Int64` are called
