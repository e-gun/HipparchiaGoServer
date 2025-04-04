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

${GO} build -pgo=${PGF} -ldflags "${LDF}"

./HipparchiaGoServer  -gl 3 -el 2 -sp 8001 -cm Triadic -ft Fira -pg "{\"Pass\": \"YOURPASSWORDHERE\" ,\"Host\": \"127.0.0.1\", \"Port\": 5432, \"DBName\": \"hgdb\" ,\"User\": \"hgdbuser\"}"
