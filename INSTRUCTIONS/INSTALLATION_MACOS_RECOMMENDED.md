## Installing HipparchiaGoServer

1. first install and configure `PostgreSQL`
1. next acquire a binary for `HipparchiaGoServer`
1. load `hipparchiaDB` into `PostgreSQL` on the first launch of `HipparchiaGoServer`

### [A] install and configure `Postgres.app`

1. You will install `Postgres.app` which will then manage `PostgreSQL` for you. So go to https://postgresapp.com.

   ![inst02](../gitimg/macos_posgresapp/00_getpostgres.png)
   
2. Download the application and drag it into your 'Applications' folder.

    ![inst02](../gitimg/macos_posgresapp/01_getpostgres.png)

3. Launch the application. And then click `Initialize`.

    ![inst02](../gitimg/macos_posgresapp/02_initialize.png)

4. After you have initizlized the server you might want to check that `PostgreSQL` is set to be constantly running. See `Server Settings...` and look for `Automatically start server`.

    ![inst02](../gitimg/macos_posgresapp/03_autostart.png)

### [B] acquire `HipparchiaGoServer` and launch it

1. You can build `HipparchiaGoServer` yourself with the files in this repository. Or you can grab a pre-built binary. Download the correct binary. Intel Macs: `-darwin-amd64-` Apple Silicon: `-darwin-arm64-`

![inst12](../gitimg/windows/16_getbinary.png)

2. If you download a file like `HipparchiaGoServer-darwin-arm64-1.0.18.zip`, it needs to be UNZIPPED. Double-clicking will do that. You will then see something like `HipparchiaGoServer-darwin-arm64-1.0.18` in the same folder.

3. This file needs to be RENAMED: `HipparchiaGoServer-darwin-arm64-1.0.18` --> `HipparchiaGoServer`

   ![inst12](../gitimg/macos_homebrew/12_renamea.png)

   ![inst13](../gitimg/macos_homebrew/13_renameb.png)

4. Double-click to launch. It is possible that you will get a complaint about an UNIDENTIFIED DEVELOPER. In that case you need to go to `System Settings` -> `Gatekeeper` -> `Security` and then allow this application to run.

   ![inst14](../gitimg/macos_homebrew/14_gatekeeper.png)

### [C] the first launch of `HipparchiaGoServer`: loading `hipparchiaDB` into `PostgreSQL`

1. The database load happens the first time you run `HipparchiaGoServer`. It will take *several minutes*.

2. On the first run you will be asked for the password for `hippa_wr`.

   ![inst15](../gitimg/macos_homebrew/15_firstrun.png)

3. Then you will be told that the self-load is about to begin.

   ![inst02](../gitimg/macos_posgresapp/05_selfload.png)

4. After several minutes the server will launch. The self-load process only has to happen once.

   ![inst02](../gitimg/macos_posgresapp/06_selfload_done.png)


### [D] Troubleshooting / Resetting

1. You will be working with `Terminal.app`. Launch it.

   ![inst01](../gitimg/macos_homebrew/01_terminal.png)

2. If you want to zap everything and start over:
- `rm ~/.config/hgs-conf.json`
- `/Applications/Postgres.app/Contents/Versions/15/bin/psql postgres` [note that `15` might change at some date]
- - inside of `psql` enter the following
- - `DROP DATABASE "hipparchiaDB"`
- - `DROP USER hippa_wr`
- - `\q`

3. The next time you run `HipparchiaGoServer` will be like a first launch.