#!/bin/sh

GO="go"
DT=$(date "+%Y-%m-%d@%H:%M:%S")
GC=$(git rev-list -1 HEAD | cut -c-8)

PG="default.pgo"
PGF="./pgo/${PG}"

VS=""
if [ `git branch --show-current` != "stable" ]; then
  VS="-pre"
fi

LDF="-s -w -X main.GitCommit=${GC} -X main.BuildDate=${DT} -X main.VersSuppl=${VS} -X main.PGOInfo=${PG}"

GOEXPERIMENT=greenteagc,jsonv2 ${GO} build -pgo=${PGF} -ldflags "${LDF}"

./HipparchiaGoServer  -gl 3 -el 1 -sp 8001 -cm Triadic -ft Fira
