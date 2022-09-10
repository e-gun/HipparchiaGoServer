# HipparchiaGoServer

## alpha status: does a lot of what you most want acceptably...

### status as of `v0.3.1`:

* searching mostly works 
* browsing mostly works
* selections mostly work 
* progress polling mostly works
* dictionaries mostly work
* basic text maker
* very basic indexing

### TODO (in vague order of priority)

* sort results by date
* saving/fetching searches
* vocabulary list generation
* search list info box
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
