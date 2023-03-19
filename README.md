# HipparchiaGoServer

### STATUS (v0.0.0):

* 25%-700% faster than HipparchiaServer depending on the function. Uses c. 60% as much RAM.
* monolithic binary: no need for extra files/folders beyond setting a password in `hgs-conf.json`

### FEATURES

* searching
* browsing 
* selections 
* exclusions 
* not-near syntax
* semantic vectors 
* dictionaries
* text maker
* indexing (+ flag unique forms)
* vocabulary list generation (+ flag unique headwords)
* morphology tables
<<<<<<< HEAD
* progress polling
* self-configuration first run
* self-installation on first run 
* vectors
=======
* self-installation on first run
* self-configuration first run
* vectors (but early tests of ML golang libraries are not promising)

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
