# Customizing an instillation

If you edit the file named `hgs-prolix-conf.json` in the folder `.config` within your home folder, you can override a number of built-in defaults. 
You need to get `PGLogin` right. The password is the most important part. The other values should be correct at their defaults.

One-time overrides are available by setting flags when launching `HipparchiaGoServer`. Try running `HipparchiaGoServer -h` to see them.

Items of most interest in the configuration file:

* `BrowserCtx` sets the default number of lines to show in the browser. The size of your monitor will determine the most convenient value for you.
* `CustomCSS` when set to `true` tells the server to load custom CSS. The server looks for a file named `custom-hipparchiastyles.css` inside `~/.config/`. You should use `hipparchiastyles.css` as a starting template.
* `DefCorp` sets which corpora are active on a reset. So if you almost never search Greek, you can set `gr` to `false`, for example. On a slow machine, this would significantly speed up `in every active author` searches.
* `Font` sets the interface font. `Noto`, `Fira`, and `Roboto` are embedded in the program. If you pick another name, you need to have it installed on your machine. 
* `QuietStart` spares you the copyright notice.
* `VectorChtHt` and `VectorChWd` will set the height and width of the vector charts. Note that `px` is required as are the quotation marks: `"1000px"`, e.g.
* `VectorMaxlines` sets the maximum scope of a vector search. `1000000` will let you model all of Latin. All of Greek is about 10x larger.
* `WorkerCount` sets the number of cores of your CPU to use when searching. You will be sorry if you pick a number that is larger than what the machine actually has installed. `WorkerCount` = `CoreCount` is probably the best choice unless you know why it is not.
* `ZapLunates` lets you see σ/ς instead of lunate sigmas. But why would you want that?


```
{
    "Authenticate": false,
    "BadChars": "\"'!@:,=_/",
    "BlackAndWhite": false,
    "BrowserCtx": 14,
    "CustomCSS": false,
    "DbDebug": false,
    "DefCorp": {
      "ch": false,
      "dp": false,
      "gr": true,
      "in": false,
      "lt": true
    },
    "EchoLog": 0,
    "Font": "Noto",
    "Gzip": false,
    "HostIP": "127.0.0.1",
    "HostPort": 8000,
    "LdaGraph": false,
    "LdaTopics": 8,
    "LogLevel": 0,
    "ManualGC": true,
    "MaxText": 35000,
    "MaxSrchIP": 2,
    "MaxSrchTot": 4,
    "PGLogin": {
      "Host": "127.0.0.1",
      "Port": 5432,
      "User": "hippa_wr",
      "Pass": "myrandompassword",
      "DBName": "hipparchiaDB"
    },
    "ProfileCPU": false,
    "ProfileMEM": false,
    "ResetVectors": false,
    "QuietStart": false,
    "SelfTest": 0,
    "TickerActive": false,
    "VectorsDisabled": false,
    "VectorBot": false,
    "VectorChtHt": "1200px",
    "VectorChtWd": "1500px",
    "VectorMaxlines": 1000000,
    "VectorModel": "w2v",
    "VectorNeighb": 16,
    "VectorTextPrep": "winner",
    "VectorWebExt": false,
    "VocabByCt": false,
    "VocabScans": false,
    "WorkerCount": 20,
    "ZapLunates": false
  }
```

Maniacs can edit `./emb/css/hgs.css` and/or `theconstants.go` and then build a custom binary: `go build`. That would really change things up. 

It is not such a good idea to expose `HipparchiaGoServer` to the internet as a whole. But you can. If you do, 
you almost certainly want to set `Authentication` to `true`. If you do that, then you need a `hgs-users.json` file to be 
placed in `~/.config/`. It is formatted as follows:

```
[
  {
    "User": "user1",
    "Pass": "pass1"
  },
  {
    "User":"user2",
    "Pass":"pass2"
  }
]
```

The full list of configuration files follows. Note that many of them are specific to `SEMATNTICVECTORS`. See that help file for more.

```
~/.config/ % ls -1 hgs-*
hgs-prolix-conf.json
hgs-users.json
hgs-vector-conf-glove.json
hgs-vector-conf-lda.json
hgs-vector-conf-lexvec.json
hgs-vector-conf-w2v.json
hgs-vector-stops-greek.json
hgs-vector-stops-latin.json
```