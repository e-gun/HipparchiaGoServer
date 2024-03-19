## Installing HipparchiaGoServer

1. first install and configure `PostgreSQL`
1. next acquire a binary for `HipparchiaGoServer`
1. load `hipparchiaDB` into `PostgreSQL` on the first launch of `HipparchiaGoServer`
1. [fyi] how to archive and/or migrate the data
1. [fyi] how to reset the database and start over

---

### [A] install and configure `Postgres.app`

1. You will install `Postgres.app` which will then manage `PostgreSQL` for you. So go to https://postgresapp.com.

   ![inst02](./gitimg/macos_posgresapp/00_getpostgres.png)
   
2. Download the application and drag it into your 'Applications' folder.

    ![inst02](./gitimg/macos_posgresapp/01_getpostgres.png)

3. Launch the application. And then click `Initialize`.

    ![inst02](./gitimg/macos_posgresapp/02_initialize.png)

4. After you have initialized the server you might want to check that `PostgreSQL` is set to be constantly running. See `Server Settings...` and look for `Automatically start server`.

    ![inst02](./gitimg/macos_posgresapp/03_autostart.png)

5. `Postgres.app` will add an item to your menu bar.

   ![inst02](./gitimg/macos_posgresapp/03b_menu_item.png)

---

### [B] acquire `HipparchiaGoServer` and launch it

1. You can build `HipparchiaGoServer` yourself with the files in this repository (https://github.com/e-gun/HipparchiaGoServer). Or you can grab a pre-built binary from the site pictured below. Download the correct binary. Intel Macs: `darwin-amd64` Apple Silicon: `darwin-arm64`

![inst12](./gitimg/windows/16_getbinary.png)

2. If you download a file like `HipparchiaGoServer-1.1.0-darwin-amd64.zip`, it needs to be UNZIPPED. Double-clicking will do that. You will then see something like `HipparchiaGoServer-1.1.0-darwin-amd64` in the same folder.

3. This file *might* need to be RENAMED: `HipparchiaGoServer-1.1.0-darwin-amd64` --> `HipparchiaGoServer`

---

### [C] the first launch of `HipparchiaGoServer`: loading `hipparchiaDB` into `PostgreSQL`

0. You need to have the DATA available. [The data needs to come from a `pg_dump` of a working `HipparchiaGoServer` installation. If a working installation executes `HipparchiaGoServer -ex`, it will generate a valid `hDB` folder.]
   The data *must* reside in a folder named `hDB`. This folder has to be in the same folder as `HipparchiaGoServer`. Note that `hdb` ≠ `hDB`.
   You can (re)move the data folder after you have successfully installed the data into the database.

![inst02](./gitimg/macos_posgresapp/04_data_is_ready.png)

1. Double-click to launch. It is possible that you will get a complaint about an UNIDENTIFIED DEVELOPER.
   In that case you need to go to `System Settings` -> `Gatekeeper` -> `Security` and then allow this application to run
   by clicking OPEN ANYWAY and then entering a password to confirm this.

![inst02](./gitimg/macos_posgresapp/04a_firstrun_unidentified.png)

![inst02](./gitimg/macos_posgresapp/04b_firstrun_gatekeeper.png)

![inst02](./gitimg/macos_posgresapp/04c_firstrun_permission.png)

2. The database load happens the first time you run `HipparchiaGoServer`. This will take *several minutes*.

3. On the first run instruction files will be dropped into your current working directory. You will be asked for the password for `hippa_wr`.

   ![inst15](./gitimg/macos_posgresapp/04_firstrun.png)

4. Then you will be told that the self-load is about to begin.

   ![inst02](./gitimg/macos_posgresapp/05_selfload.png)

5. Thousands of messages will fly across the screen.

   ![inst02](./gitimg/macos_posgresapp/05b_loading_in_progress.png)

6. Eventually the server will launch. The self-load process only has to happen once.

   ![inst02](./gitimg/macos_posgresapp/06_selfload_done.png)

NB: `hippa_rd` errors are safe to ignore.

![inst02](./gitimg/macos_posgresapp/06b_selfload_done.png)

7. When you see `http server started on 127.0.0.1:8000` you are up and running. Now you can point a browser at http://127.0.0.1:8000.
Whenever the server is running, your browser can interact with http://127.0.0.1:8000. You can also leave the server 
running indefinitely. It does not consume many resources if not active: 0% CPU, <1% RAM.


---

### [D] [FYI] Archiving / Migrating

1. If you lose/destroy the `hDB` folder with the original data and want it back, the data can be extracted and archived.

2. Move `HipparchiaGoServer` into your home directory. Launch `Terminal.App`

3. Type `./HipparchiaGoServer -ex`. The data will be put into a new `hDB` folder in the current directory.

---

### [E] [FYI] Troubleshooting / Resetting

#### [E1] easier

1. Move `HipparchiaGoServer` into your home directory. Launch `Terminal.App`

2. Type `./HipparchiaGoServer -00`. If you say `YES`, the database will reset itself.

![inst13](./gitimg/macos_posgresapp/07a_selfreset.png)

![inst13](./gitimg/macos_posgresapp/07b_selfreset_done.png)

#### [E2] less easy

1. You will be working with `Terminal.app`. Launch it.

   ![inst01](./gitimg/macos_homebrew/01_terminal.png)

2. If you want to zap everything and start over, then open `Terminal.app` and enter the following:
- `rm ~/.config/hgs-*.json`
- `/Applications/Postgres.app/Contents/Versions/15/bin/psql postgres` [note that `15` might change at some date]
- - inside of `psql` enter the following
- - `DROP DATABASE "hipparchiaDB";`
- - `DROP USER hippa_wr;`
- - `DROP EXTENSION pg_trgm;`
- - `\q`

![inst01](./gitimg/macos_posgresapp/07_reset.png)


3. The next time you run `HipparchiaGoServer` will be like a first launch as per the above.
