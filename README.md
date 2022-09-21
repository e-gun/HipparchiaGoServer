# HipparchiaGoServer

## early beta status: rough edges, (allegedly) solid core

### status as of `v0.7.0`:

* monolithic binary: no need for extra files/folders beyond `hgs-conf.json`
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

* update the help files so that they reflect HGS vs the python server
* fix the less wonky lemma highlighting
* keying for multiple author/work indexing (this will also show that not all works of an author are being indexed ATM)

### DISTANT FANTASIES
* vectors

---

the password for `hippa_wr` will be requested when archiving/loading

archive:
```
pg_dump --clean "hipparchiaDB" --user hippa_wr | split -b 100m - /ARCHIVEFOLDER/TARGETDIR/hipparchiaDB-
tar -jcf /ARCHIVEFOLDER/TARGETDIR.tar.bz2 /ARCHIVEFOLDER/TARGETDIR
rm -rf /ARCHIVEFOLDER/TARGETDIR
```

load:
```
tar -jxf /ARCHIVEFOLDER/TARGETDIR.tar.bz2
cat /ARCHIVEFOLDER/TARGETDIR/hipparchiaDB-* | psql "hipparchiaDB" --user hippa_wr
```

an uncompressed archive is c. `3.6GB`

---

## minimal installation overview

- install PostgreSQL: this is a hard requirement; various platforms have various installation options; see [HERE](https://www.postgresql.org/download/)
- after installing PostgreSQL
  - add a user named `hippa_wr` to postgresql
  - create a database named `hipparchiaDB`
- load an archive into the database: (as per the above) `cat /ARCHIVEFOLDER/TARGETDIR/hipparchiaDB-* | psql hipparchiaDB`
- compile or download a binary of `HipparchiaGoServer`
- set up `config.json` (as per `sample_config.json`)
- run `HipparchiaGoServer` via a terminal

## some further notes

- ask google about adding a user and database to postgresql; but the easiest way is probably:
  - from the command line of a terminal execute `psql postgres` (or `sudo su postgres psql postgres` if you do not have permission to do the first). NB: `Postgres.app` on macOS will install `psql` somewhere where you can't just type `psql postgres` to get into the database...
  - then send the next two commands from the postgres shell:
    - `CREATE USER hippa_wr WITH PASSWORD 'some_random_password';`
    - `CREATE DATABASE "hipparchiaDB";`
  - exit postgresql by telling it `\q`
- loading the archive is also easy if you are in the right directory; but you need to know where you are...
  - loading can take several minutes
- if no configuration is present when you launch `HipparchiaGoServer`, you will be given a prompt about what it should look like; all you should really need to do is set the password (avoid adding “smart quotes” instead of straight quotes to this file)


## notes on installing homebrew in order to acquire postgresql
- on a mac this is probably the quickest/easiest way in the end even if `homebrew` is so much more than just `psql`
- go to [https://docs.brew.sh/Installation](https://docs.brew.sh/Installation) and follow the instructions
- you will be using `Terminal.app` throughout
- you will be installing `Xcode` from apple first; then you can get `homebrew`; then you can ask for `postgresql`
  - so: `xcode-select --install` (say "yes" to all the stuff you have to agree to; it might take a while for this to download and install) 
  - then: `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install.sh)"` (will require your admin password)
  - then one of the following: 
    - `/opt/homebrew/bin/brew install postgresql` (M1 macs) 
    - `/usr/local/bin/brew install postgresql` (intel macs)
  - then `brew services restart postgresql`

## notes on the EDB PostgreSQL Setup Wizard

- accept all defaults until you get to the "password" option: where to install the files, what tools to install, etc. 
- the first password you create will be for and admin user named `posgres` and *NOT* for Hipparchia's user
- pick a good/strong password and write it down somewhere
- continue accepting defaults: `port 5432`; `default locale`
- no need for "Stack Builder" at end
- the wizard will claim that it installed `pgAdmin`, but it might well be broken; if so, grab it as per the following

## notes about pgAdmin

- get `pgAdmin` from [https://www.pgadmin.org/download](https://www.pgadmin.org/download)
- you might have to install `Rosetta` to run pgAdmin on a mac; just click "yes" when asked to do this
- the first launch of `pgAdmin` can take a while
- you will be asked to set *another* password for `pdAdmin` itself; pick a good/strong password and write it down somewhere
- the command line prompt can be reached via the `_>` button on the top row of icons