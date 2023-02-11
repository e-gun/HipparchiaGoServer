## Installing HipparchiaGoServer

1. first install and configure `PostgreSQL`
1. next load `hipparchiaDB` into `PostgreSQL`
1. next acquire a binary for `HipparchiaGoServer`

### [A] install and configure `PostgreSQL`
1. download from http://postgresql.org/downloads (which will send you to enterprisedb.com...)

![dl](gitimg/windows/01_getpsql.png)

2. launch the installer `postgresql-15.1-1-windows-x64.exe` (vel sim)

![launch](gitimg/windows/02_setuppsql.png)

3. click through the installation options accepting *most* of the defaults...

![inst1](gitimg/windows/03_psqldir.png)

![inst2](gitimg/windows/04_psqlcomponents.png)

![inst3](gitimg/windows/05_psqldata.png)

4. stop mindlessly clicking 'Next >' and pick an ADMIN password; write it down somewhere; you will need to pick a different password later as a USER password

![inst4](gitimg/windows/06_db_adminpass.png)

5. return to accepting defaults...

![inst5](gitimg/windows/07_dbport.png)

6. this one is big: you must pick `C` as your `locale`

![inst6](gitimg/windows/08_locale.png)

7. back to just clicking forwards...

![inst7](gitimg/windows/09_summary.png)

8. wait...

![inst8](gitimg/windows/10_psqlinstalling.png)

9. done. click "Finish". Do not launch Stack Builder

![inst9](gitimg/windows/11_psqlinstallationends.png)

### [B] load `hipparchiaDB` into `PostgreSQL`
1. Launch `SQL Shell` which lives inside the `PostgreSQL 15` folder

![inst10](gitimg/windows/12_find_sqlshell.png)

2. Gain access to the `postgres` database by hitting `RETURN` 4x: you are accepting the default supplied values; 
at the fifth stop you will need to enter the ADMIN password you set earlier.

![inst11](gitimg/windows/13_insidesqlshell.png)

3. Now you will be creating a user (`hippa_wr`), creating a database (`hipparchiaDB`), giving the user 
permission to access the database, enabling fast indexing, and then quitting. You need to enter each line EXACTLY as
seen below but for the part where you enter a real password instead of `random_password`. All punctuation 
matters (a lot): quotation marks, semicolons, ...

![inst11](gitimg/windows/14_furtherinsidesqlshell.png)

4. Now you load the data into `PostgreSQL`. 
* First launch `PowerShell`. 
* Then `cd` to the directory that contains the 
data you will be loading. There is no need to `cd` if the data is in your home directory already. 
* Then set an alias to the `pg_restore.exe` application. You might need to change `15` in the example below to some
other number.
* Then execute `pg_restore`. The sample image has a typo. Make sure you enter `--username=hippa_wr`. 
You also need to set the name of the folder where the data lives properly. It might not be `hDB`.

![inst12](gitimg/windows/15_loaddata.png)

### [C] acquire `HipparchiaGoServer.exe` and launch it
1. You can build `HipparchiaGoServer.exe` yourself with the files in this repository. Or you can grab a pre-built binary.

![inst13](gitimg/windows/16_getbinary.png)

2. Double-click on the binary to launch. On the first launch you will be asked to enter the password for `hippa_wr`.

![inst13](gitimg/windows/17_firstlaunch.png)

3. A configuration file will be generated and now you are running.

![inst13](gitimg/windows/18_launch.png)

4. Now you can point a browser at http://127.0.0.1:8000
