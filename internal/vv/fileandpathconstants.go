//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

// see also vectorconst.go

const (
	CONFIGLOCATION    = "."
	CONFIGALTAPTH     = "%s/.config/" // %s = os.UserHomeDir()
	CONFIGAUTH        = "hgs-users.json"
	CONFIGBASIC       = "hgs-conf.json"
	CONFIGPROLIX      = "hgs-prolix-conf.json"
	CUSTOMCSSFILENAME = "custom-hipparchiastyles.css"
	HDBFOLDER         = "hDB"
	LOGFILEEL         = "hgs-echo.log"
	LOGFILEML         = "hgs-msg.log" // circular import problem; need to edit "messaging.go" too if changing this
	SSLCERTDIR        = "./sslcerts/"
	SSLCPEM           = "cert.pem"
	SSLPPEM           = "privkey.pem"
	WRITEPERMS        = 0644
)
