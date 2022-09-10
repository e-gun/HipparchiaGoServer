# HipparchiaGoServer

## alpha status: does a lot of what you most want acceptably...

### status as of `v0.2.8`:

* searching mostly works 
* browsing mostly works 
* dictionaries mostly work 
* selections mostly work 
* progress polling mostly works
* basic text maker
* very basic indexing

### TODO (in vague order of priority)

* sort results by date
* saving/fetching searches
* vocabulary list generation
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
