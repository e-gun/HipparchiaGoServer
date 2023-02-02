#!/bin/sh

DT=$(date "+%Y-%m-%d@%H:%M:%S")
GC=$(git rev-list -1 HEAD | cut -c-8)
go build -ldflags "-s -w -X main.GitCommit=${GC} -X main.BuildDate=${DT}"
P="HipparchiaGoServer"

V=$(./${P} -v)
U=$(uname)
A=$(uname -p)
U="${U}-${A}"

cp ${P} ${P}-$U-${V}
bzip2 -f ${P}-${U}-${V}
