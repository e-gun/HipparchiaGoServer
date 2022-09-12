# HipparchiaGoServer

## alpha status: does a lot of what you most want acceptably...

### status as of `v0.3.9`:

* searching mostly works 
* browsing mostly works
* selections mostly work 
* exclusions mostly work
* progress polling mostly works
* dictionaries mostly work
* basic text maker
* very basic indexing

### TODO (in vague order of priority)

* not-near in progress: failing in the second pass
* UI needs to reflect settings better
* fix wonky lemma highlighting
* saving/fetching searches
* vocabulary list generation
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
