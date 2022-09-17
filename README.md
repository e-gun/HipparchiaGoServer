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

---

## minimal installation overview

* install postgresql: this is a hard requirement
..* various platforms have various installation options; see [HERE](https://www.postgresql.org/download/)
..* add a user named `hippa_wr` to postgresql
..* create a database named `hipparchiaDB`
* load an archive into the database: (as per the above) `cat /ARCHIVEFOLDER/TARGETDIR/hipparchiaDB-* | psql hipparchiaDB`
* download a binary of `HipparchiaGoServer`
* set up `config.json` (as per `sample_config.json`)
* run `HipparchiaGoServer` via a terminal

## some further notes

* ask google about adding a user and database to postgresql; but the easiest way is probably:
..* execute `psql` from the command line of a terminal. NB: `Postgres.app` on macOS will install `psql` somewhere where you can't just type `psql` to get into the database...
..* then send the next two commands from the postgres shell
..* `CREATE USER hippa_wr;`
..* `CREATE DATABASE hipparchiaDB;`
..* exit postgresql by telling it `\q`
* loading the archive is also easy if you are in the right directory; but you need to know where you are...
* loading can take several minutes
* if no configuration is present when you launch `HipparchiaGoServer`, you will be given a prompt about what it should look like. all you should really need to do is set the password (avoid adding "smart quotes" instead of straight quotes to this file)
