package structs

type CurrentConfiguration struct {
	Authenticate    bool
	BadChars        string
	BlackAndWhite   bool
	BrowserCtx      int
	CustomCSS       bool
	DbDebug         bool
	DefCorp         map[string]bool
	EchoLog         int // 0: "none", 1: "terse", 2: "prolix", 3: "prolix+remoteip"
	Font            string
	Gzip            bool
	HostIP          string
	HostPort        int
	LdaGraph        bool
	LdaTopics       int
	LogLevel        int
	ManualGC        bool // see messenger.LogPaths()
	MaxText         int
	MaxSrchIP       int
	MaxSrchTot      int
	PGLogin         PostgresLogin
	ProfileCPU      bool
	ProfileMEM      bool
	ResetVectors    bool
	QuietStart      bool
	SelfTest        int
	TickerActive    bool
	VectorsDisabled bool
	VectorBot       bool
	VectorChtHt     string
	VectorChtWd     string
	VectorMaxlines  int
	VectorModel     string
	VectorNeighb    int
	VectorTextPrep  string
	VectorWebExt    bool // "simple" when false; "expanded" when true
	VocabByCt       bool
	VocabScans      bool
	WorkerCount     int
	ZapLunates      bool
}
