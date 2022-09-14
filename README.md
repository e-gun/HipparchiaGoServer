# HipparchiaGoServer

## alpha status: mostly does most of what you most want...

### status as of `v0.4.2`:

* searching mostly works 
* browsing mostly works
* selections mostly work 
* exclusions mostly work
* not-near ok
* progress polling mostly works
* dictionaries mostly work
* basic text maker
* very basic indexing
* very basic vocabulary list generation

### TODO (in vague order of priority)

* UI needs to reflect settings better
* fix the less wonky lemma highlighting
* index and vocab maker could use a lot of polishing
* saving/fetching searches
* maximize fixed length slices
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

an uncompressed archive is c. 3.6GB
