# HipparchiaGoServer

## alpha status: does a lot of what you most want acceptably...

### status as of `v0.2.2`:

* searching mostly works 
* browsing mostly works 
* dictionaries mostly work 
* selections mostly work 
* progress polling mostly works 
* parallelized hit context fetching


### TODO (in vague order of priority)
* rationalize selections
* within X words 
* statistics for dictionary entries
* bracket highlighter
* better browser formatting
* saving/fetching searches
* indexing
* simple text generation
* vocabulary list generation
* author info box
* search list info box
* dictionary searches
* reverse lexical searches
* morphology searches
* raw input style
* onehit searching

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

an uncompressed archive is c. 3.6GB
