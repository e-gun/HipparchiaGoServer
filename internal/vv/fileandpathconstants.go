//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

// see also vectorconst.go

const (
	CONFIGLOCATION    = "."
	CONFIGALTAPTH     = "%s/.config/HipparchiaGoServer" // %s = os.UserHomeDir()
	CONFIGAUTH        = "users.json"
	CONFIGOPTIONS     = "options.json"
	CONFIGCOLORS      = "colors.json"
	CUSTOMCSSFILENAME = "custom-hipparchiastyles.css"
	HDBFOLDER         = "HGDBArchive"
	LOGFILEEL         = "hgs-echo.log"
	LOGFILEML         = "hgs-msg.log" // circular import problem; need to edit "messaging.go" too if changing this
	SSLCERTDIR        = "./sslcerts/"
	SSLCPEM           = "cert.pem"
	SSLPPEM           = "privkey.pem"
	WRITEPERMS        = 0644
)
