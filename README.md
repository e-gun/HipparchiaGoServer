# HipparchiaGoServer

## late alpha status: mostly does most of what you most want...

### status as of `v0.5.0`:

* monolithic binary: no need for extra files/folders beyond `config.json`
* searching theoretically fine but needs more testing/checking
* browsing 
* selections 
* exclusions 
* not-near
* progress polling 
* dictionaries
* text maker
* indexing (needs more polish)
* vocabulary list generation (needs more polish)

### TODO (in vague order of priority)

* fix the less wonky lemma highlighting
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
