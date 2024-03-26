//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

const (
	MINCONFIG = `
{"PostgreSQLPassword": "YOURPASSWORDHERE"}
`

	TERMINALTEXT = `Copyright (C) %s / %s
      %s

      This program comes with ABSOLUTELY NO WARRANTY; without even the  
      implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

      This is free software, and you are welcome to redistribute it and/or 
      modify it under the terms of the GNU General Public License version 3.`

	PROJYEAR = "2022-24"
	PROJAUTH = "E. Gunderson"
	PROJMAIL = "Department of Classics, 125 Queen’s Park, Toronto, ON  M5S 2C7 Canada"
	PROJURL  = "https://github.com/e-gun/HipparchiaGoServer"

	HELPTEXTTEMPLATE = `S3command line optionsS0:
   C1-auC0          require authentication; also implies "C3{{.confauth}}C0" exists and has been properly configured
   C1-avC0          automatically generate vector models for every author in the database
                   default settings will consume c. C11.3GBC0 of extra disk space
   C1-bcC0 C2{num}C0    default lines of browser context to display [C6currentC0: C3{{.ctxlines}}C0]
   C1-bwC0          disable color output in the console
   C1-csC0          use a custom CSS file; will try to read "C3{{.home}}{{.css}}C0"
   C1-dbC0          debug database: show internal references in browsed passages
   C1-dvC0          disable semantic vector searching
   C1-elC0 C2{num}C0    set echo server log level (C10-3C0) [C6currentC0: C3{{.echoll}}C0]
   C1-exC0          extract the data to an archive folder in the same directory as the application; data sent to: "C3{{.cwd}}C0"
   C1-ftC0 C2{string}C0 change the font [C6built-in:C0 C3{{.knownfnts}}C0][C6currentC0: C3{{.deffnt}}C0]
                   names with spaces need quotes around them: "C4Gentium Plus CompactC0"
   C1-glC0 C2{num}C0    set golang log level (C10-5C0) [C6currentC0: C3{{.hgsll}}C0]
   C1-gzC0          enable gzip compression of the server's output
   C1-hC0           print this help information
   C1-mdC0 C2{string}C0 set the default vector model type; available: C3gloveC0, C3lexvecC0, and C3w2vC0 [C6currentC0: C3{{.vmodel}}C0]
   C1-miC0 C2{num}C0    maximum number of concurrent searches per IP address [C6currentC0: C3{{.maxipsrch}}C0]
   C1-msC0 C2{num}C0    maximum total number of concurrent searches [C6currentC0: C3{{.maxtotscrh}}C0]
   C1-pcC0          enable CPU profiling run
   C1-pdC0          write a copy of the embedded PDF instructions to the current directory
   C1-pmC0          enable MEM profiling run
   C1-pgC0 C2{string}C0 supply full PostgreSQL credentials C4(*)C0
   C1-qC0           quiet startup: suppress copyright notice
   C1-rlC0          reload the database tables; data will be read from: "C3{{.dbf}}C0" in "C3{{.cwd}}C0"
   C1-rvC0          reset the stored semantic vector table
   C1-saC0 C2{string}C0 server IP address [C6currentC0: C3{{.host}}C0]
   C1-spC0 C2{num}C0    server port [C6currentC0: C3{{.port}}C0]
   C1-stC0          run the self-test suite at vv; repeat the flag to iterate: e.g., "C1-st -stC0" will run twice
   C1-tkC0          turn on the uptime UptimeTicker [unavailable if OS is Windows]
   C1-uiC0 C2{string}C0 unacceptable input characters [C6currentC0: C3{{.badchars}}C0]
   C1-vC0           print version info and exit
   C1-vvC0          print full version info and exit
   C1-wcC0 C2{int}C0    number of workers [C1cpu_countC0 is C3{{.cpus}}C0][C6currentC0: C3{{.workers}}C0]
   C1-zlC0          zap lunate sigmas and replace them with C1σ/ςC0
   C1-00C0          completely erase the database and reset the tables
                   the application cannot run again until you restore its data from an archive 
                   you probably want to run with the "C1-exC0" flag before you try this.
     (*) S3exampleS0: 
         C4"{\"Pass\": \"YOURPASSWORDHERE\" ,\"Host\": \"127.0.0.1\", \"Port\": 5432, \"DBName\": \"hipparchiaDB\" ,\"User\": \"hippa_wr\"}"C0
     
     S1NB:S0 a properly formatted version of "C3{{.conffile}}C0" in "C3{{.home}}C0" configures everything for you. 
         See "C3sample_hgs-prolix-vv.jsonC0"" as well as other sample configuration files at
             C3{{.projurl}}C0
`
)
