If you put a file named `hgs-prolix-conf.json` into `.config` in your home folder, you can override a number of built-in defaults. 
You can also delete the very basic `hgo-conf.json` if you do this, but you will want to get `PGLogin` right.

One-time overrides are available by setting flags when launching `HipparchiaGoServer`. Try running `HipparchiaGoServer -h` to see them.

You can skip any of the items below and they will have defaults inserted instead. 

Items of most interest:

* `BrowserCtx` sets the default number of lines to show in the browser. The size of your monitor will determine the most convenient value for you.
* `DefCorp` sets which corpora are active on a reset. So if you almost never search Greek, you can set `gr` to `false`, for example. On a slow machine, this would significantly speed up `in every active author` searches.
* `Font` sets the interface font. `Noto` is embedded in the program. If you pick another name, you need to have it installed on your machine. 
* `QuietStart` spares you the copyright notice.
* `WorkerCount` sets the number of cores of your CPU to use when searching. You will be sorry if you pick a number that is larger than what the machine actually has installed. `WorkerCount` = `CoreCount` is probably the best choice unless you know why it is not.
* `ZapLunates` lets you see σ/ς instead of lunate sigmas. But why would you want that?


```
{
  "Authenticate": false,
  "BadChars": "\"'!@:,=_/",
  "BlackAndWhite": false,
  "BrowserCtx": 20,
  "DbDebug": false,
  "DefCorp":
    {"gr": true, "lt": true, "in": false, "ch": false, "dp": false},
  "EchoLog": 0,
  "Font": "Noto",
  "Gzip": false,
  "HostIP": "localhost",
  "HostPort": 8000,
  "LogLevel": 3,
  "ManualGC": true,
  "MaxText": 25000,
  "PGLogin":
    {"Pass": "" ,"Host": "127.0.0.1", "Port": 5432, "DBName": "hipparchiaDB" ,"User": "hippa_wr"},
  "QuietStart": true,
  "VocabByCt": false,
  "VocabScans": false,
  "WorkerCount": 6,
  "ZapLunates": false
}
```

Maniacs can edit `./emb/css/hgs.css` and/or `theconstants.go` and then build a custom binary: `go build`. That would really change things up. 
