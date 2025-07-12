//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
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

	PROJYEAR = "2022-25"
	PROJAUTH = "E. Gunderson"
	PROJMAIL = "Department of Classics, 125 Queen’s Park, Toronto, ON  M5S 2C7 Canada"
	PROJURL  = "https://github.com/e-gun/HipparchiaGoServer"

	HELPTEXTTEMPLATE = `S3command line optionsS0:
   C1-auC0          toggle authentication relative to the config file [C6currentC0: C3{{.authstatus}}C0]
                   "C3trueC0" requires a properly configured "C3{{.confauth}}C0" 
   C1-avC0          automatically generate vector models for every author in the database
                   default settings will consume c. C11.3GBC0 of extra disk space
   C1-bcC0 C2{num}C0    default lines of browser context to display [C6currentC0: C3{{.ctxlines}}C0]
   C1-bwC0          disable color output in the console
   C1-cdC0          redefine a color scheme: C4name hue sat lum1 lum2C0 [does not work for C3LightC0 or C3DarkC0]
                   example: C4"./HipparchiaGoServer -cm Monochrome -cd 0 40 95 8 Monochrome"C0
   C1-cmC0          select a default css color mode [C6currentC0: C3{{.defcol}}C0]
                   [C6built-in:C0 C3{{.knowncolors}}C0]
   C1-crC0          report color scheme names and values
   C1-csC0          use a custom CSS file; will try to read "C3{{.home}}{{.css}}C0"
   C1-dbC0          debug database: show internal references in browsed passages
   C1-dvC0          disable semantic vector searching
   C1-elC0 C2{num}C0    set echo server log level (C10-3C0) [C6currentC0: C3{{.echoll}}C0]
   C1-exC0          extract the data to an archive folder in the same directory as the application
                   data sent to: "C3{{.cwd}}C0"
   C1-ftC0 C2{string}C0 change the font [C6currentC0: C3{{.deffnt}}C0]
                   [C6known fonts:C0 C3{{.knownfnts}}C0]
                   C4BrillC0, C4IosevkaC0 and C4NotoC0 have the broadest support for rare characters.
   C1-glC0 C2{num}C0    set golang log level (C10-5C0) [C6currentC0: C3{{.hgsll}}C0]
   C1-gzC0          enable gzip compression of the server's output
   C1-hC0           print this help information
   C1-lfC0          toggle logging to textfiles in C1os.UserHomeDir()C0 relative to the config file: "C3{{.loge}}C0" and "C3{{.logm}}C0" [C6currentC0: C3{{.lfstatus}}C0]
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
   C1-sdC0 C2{string}C0 ssl certificate directory [C6currentC0: C3{{.ssldir}}C0] (requires "C4{{.sslcert}}C0" and "C4{{.sslpriv}}C0")
   C1-spC0 C2{num}C0    server http port [C6currentC0: C3{{.port}}C0]
   C1-ssC0 C2{num}C0    server https port [C6currentC0: C3{{.sslport}}C0]
   C1-stC0          run the self-test suite at vv; repeat the flag to iterate: e.g., "C1-st -stC0" will run twice
   C1-tkC0          turn on the uptime UptimeTicker [unavailable if OS is Windows]
   C1-tlC0          log lines to display in UptimeTicker [C6currentC0: C3{{.tlines}}C0]
   C1-uiC0 C2{string}C0 unacceptable input characters; be hesitant to remove items from it [C6currentC0: C3{{.badchars}}C0]
   C1-uvC0          toggle the forced formatting of u for v in the output relative to the config file [C6currentC0: C3{{.uv}}C0]
   C1-vC0           print version info and exit
   C1-vvC0          print full version info and exit
   C1-wcC0 C2{int}C0    number of workers [C1cpu_countC0 is C3{{.cpus}}C0][C6currentC0: C3{{.workers}}C0]
   C1-zlC0          zap lunate sigmas and replace them with C1σ/ςC0 (automatically set if a built-in font lacks lunates)
   C1-00C0          completely erase the database and reset the tables
                   the application cannot run again until you restore its data from an archive 
                   you probably want to run with the "C1-exC0" flag before you try this.
     (*) S3exampleS0: 
         C4"{\"Pass\": \"YOURPASSWORDHERE\" ,\"Host\": \"127.0.0.1\", \"Port\": 5432, \"DBName\": \"hipparchiaDB\" ,\"User\": \"hippa_wr\"}"C0
     
     S1NB:S0 a properly formatted version of "C3{{.conffile}}C0" in "C3{{.home}}C0" configures everything for you. 
         See "C3sample_hgs-prolix.jsonC0"" as well as other sample configuration files at
             C3{{.projurl}}C0
`
)
