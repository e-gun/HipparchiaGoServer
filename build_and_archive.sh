#!/bin/sh

go build -ldflags "-s -w"
P="HipparchiaGoServer"
T="../HipparchiaGoBinaries/cli_prebuilt_binaries"

V=$(./${P} -v | cut -d" " -f 5 | cut -d "(" -f 2 | cut -d ")" -f 1)
U=$(uname)
A=$(uname -p)
U="${U}-${A}"

mv ${P} ${P}-$U-${V}
bzip2 ${P}-${U}-${V}