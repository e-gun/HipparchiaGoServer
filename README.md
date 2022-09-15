# HipparchiaGoServer

## late alpha status: mostly does most of what you most want...

### status as of `v0.4.5`:

* searching mostly works 
* browsing mostly works
* selections mostly work 
* exclusions mostly work
* not-near ok
* progress polling mostly works
* dictionaries mostly work
* basic text maker
* basic indexing
* basic vocabulary list generation

### TODO (in vague order of priority)

* fix the less wonky lemma highlighting
* get the sorting right (and fast) on the index/vocab code
* properly associate elided words when indexing
* index by frequency
* fix js so index will do status updates w/out needing to reload the page
* saving/fetching searches
* search list info box

### DISTANT FANTASIES
* vectors

---

archive:
```
pg_dump -cC hipparchiaDB | split -b 100m - /ARCHIVEFOLDER/TARGETDIR/hipparchiaDB-
tar -jcf /ARCHIVEFOLDER/TARGETDIR.tar.bz2 /ARCHIVEFOLDER/TARGETDIR
rm -rf /ARCHIVEFOLDER/TARGETDIR
```

load:
```
tar -jxf /ARCHIVEFOLDER/TARGETDIR.tar.bz2
cat /ARCHIVEFOLDER/TARGETDIR/hipparchiaDB-* | psql hipparchiaDB
```

an uncompressed archive is c. `3.6GB`
