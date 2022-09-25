#!/bin/sh

go build -ldflags "-s -w"
P="HipparchiaGoServer"

V=$(./${P} -v | cut -d" " -f 5 | cut -d "(" -f 2 | cut -d ")" -f 1)
U=$(uname)
A=$(uname -p)
U="${U}-${A}"

cp ${P} ${P}-$U-${V}
bzip2 -f ${P}-${U}-${V}
