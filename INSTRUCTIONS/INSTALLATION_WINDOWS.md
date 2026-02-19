## Installing HipparchiaGoServer

1. first install and configure `PostgreSQL`
2. next acquire a binary for `HipparchiaGoServer`
3. then load `hgdb` into `PostgreSQL` when running `HipparchiaGoServer` for the first time
4. [fyi] how to archive and/or migrate the data
5. [fyi] how to reset the database and start over

---

### [A] install and configure `PostgreSQL`
1. download from http://postgresql.org/downloads (which will send you to enterprisedb.com...)

![dl](./gitimg/windows/01_getpsql.png)

1. launch the installer `postgresql-15.1-1-windows-x64.exe` (vel sim)

![launch](./gitimg/windows/02_setuppsql.png)

1. click through the installation options accepting *most* of the defaults...

![inst1](./gitimg/windows/03_psqldir.png)

![inst2](./gitimg/windows/04_psqlcomponents.png)

![inst3](./gitimg/windows/05_psqldata.png)

1. stop mindlessly clicking 'Next >' and pick an ADMIN password. Write it down somewhere. Do not lose it. 
You will need to pick a different password later as a USER password. You will need the ADMIN
password at `C.2` below. 

![inst4](./gitimg/windows/06_db_adminpass.png)

1. return to accepting defaults...
   (note that if you select a port other than `5432` the initial insertion of the data will become very tricky since the installer will default to `5342` and you will need to install with various command line options set)

![inst5](./gitimg/windows/07_dbport.png)

1. this one is *extremely important* and not a default value: you must pick `C` as your `locale`

![inst6](./gitimg/windows/08_locale.png)

1. back to just clicking forwards...

![inst7](./gitimg/windows/09_summary.png)

1. wait...

![inst8](./gitimg/windows/10_psqlinstalling.png)

1. done. click "Finish". Do not launch Stack Builder

![inst9](./gitimg/windows/11_psqlinstallationends.png)

---

### [B] acquire `HipparchiaGoServer.exe` and launch it
1. You can build `HipparchiaGoServer.exe` yourself with the files in this repository (https://github.com/e-gun/HipparchiaGoServer). Or you can grab a pre-built binary from the site pictured below.

![inst13](./gitimg/windows/16_getbinary.png)

1. Double-click on the binary to launch. 

---

### [C] the first launch of `HipparchiaGoServer`: loading `hgdb` into `PostgreSQL`
1. You need to have the DATA available. [The data needs to come from a `pg_dump` of a working `HipparchiaGoServer` installation.]
   The data *must* reside in a folder named `HGDBArchive`. This folder has to be in the same folder as `HipparchiaGoServer`. 
See the image and note that both are present in the same directory. You can (re)move the data folder after you
have successfully installed the data into the database.

NB: The data will already be available if you build the database yourself with `HipparchiaGoBuilder`. And several of the steps below will be skipped. But building and running `HipparchiaGoBuilder` is beyond the scope of these instructions.
  
![inst13](./gitimg/windows/16b_have_binary.png)

1. The first launch might cause `Microsoft Defender` to complain that the app is `unrecognized`. Click `More Info` and then `Run anyway`.

![inst13](./gitimg/windows/16c_smartscreen_01.png)

![inst13](./gitimg/windows/16c_smartscreen_02.png)

1. The database load happens the first time you run `HipparchiaGoServer`. This will take *several minutes*.

2. On the first run instruction files will be dropped into your current working directory. You will be asked for a fresh password for `hippa_wr` you will also need the 
PSQL administrator password you entered at `A.4` above.

![inst13](./gitimg/windows/17_firstlaunch.png)

1. A configuration file will be generated and now `HipparchiaGoServer` will attempt to build and load its database.

![inst13](./gitimg/windows/18_preparing_to_load.png)

1. When loading you will see thousands of messages in the console.

![inst13](./gitimg/windows/19_loading.png)

1. Now you can point a browser at http://127.0.0.1:8001. You can also leave the server running indefinitely. It does not consume many resources if not active: 0% CPU, <1% RAM.


![inst13](./gitimg/windows/19b_loaded.png)

---

### [D] Archiving / Migrating

1. If you lose/destroy the `HGDBArchive` folder with the original data and want it back, the data can be extracted and archived.

2. Move `HipparchiaGoServer` into your home directory. Launch `PowerShell`

3. Type `.\HipparchiaGoServer.exe -ex`. The data will be put into a new `HGDBArchive` folder in the current directory.

---

### [E] Troubleshooting / Resetting

#### [E1] easier

1. Move `HipparchiaGoServer` into your home directory. Launch `PowerShell`

2. Type `.\HipparchiaGoServer.exe -00`. If you say `YES` and give the ADMIN password to `PostgreSQL`, the database will reset itself.

![inst13](./gitimg/windows/22_selfreset.png)


#### [E2] less easy

1. Delete the `HipparchiaGoServer` folder in the `.config` folder of your home folder.

![inst13](./gitimg/windows/21_configfile.png)

1. Launch `SQL Shell` (which can be found inside the `PostgreSQL 17` folder).

2. Gain access to the `postgres` database by hitting `RETURN` 4x: you are accepting the default supplied values;
      at the fifth stop you will need to enter the ADMIN password you set earlier in `A.4`.

![inst11](./gitimg/windows/13_insidesqlshell.png)

1. Now enter the following:
- `DROP DATABASE "hgdb";`
- `DROP USER hgdbuser;`
- `DROP EXTENSION pg_trgm;`
- `\q`

![inst11](./gitimg/windows/22_reset.png)

1. The next time you run `HipparchiaGoServer` will be like a first launch as per the above.
