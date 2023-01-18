#!/bin/sh

GC=$(git rev-list -1 HEAD | cut -c-8)
go build -ldflags "-s -w -X main.GitCommit=$GC"
P="HipparchiaGoServer"

V=$(./${P} -vv)
U=$(uname)
A=$(uname -p)
U="${U}-${A}"

cp ${P} ${P}-$U-${V}
bzip2 -f ${P}-${U}-${V}
