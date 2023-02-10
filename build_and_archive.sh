#!/bin/sh

# go get -u ./...

oss=(linux windows darwin)
archs=(amd64 arm64)

P="HipparchiaGoServer"
SUFF=""
OUT="./bin"

DT=$(date "+%Y-%m-%d@%H:%M:%S")
GC=$(git rev-list -1 HEAD | cut -c-8)

go build -ldflags "-s -w -X main.GitCommit=${GC} -X main.BuildDate=${DT}"
V=$(./${P} -v)

if [ ! -d "${OUT}" ]; then
  mkdir ${OUT}
fi

rm ${OUT}/*.zip

for arch in ${archs[@]}
do
  for os in ${oss[@]}
  do
    echo "building ${os}-${arch}"
    if [ ${os} == "windows" ]; then
      SUFF=".exe"
    else
      SUFF=""
    fi
	  env GOOS=${os} GOARCH=${arch} go build -ldflags "-s -w -X main.GitCommit=${GC} -X main.BuildDate=${DT}" -o ${P}-${os}-${arch}-${V}${SUFF}
	  zip -q ${P}-${os}-${arch}-${V}${SUFF}.zip ${P}-${os}-${arch}-${V}${SUFF}
	  mv ${P}-${os}-${arch}-${V}${SUFF}.zip ${OUT}/
	  rm ${P}-${os}-${arch}-${V}${SUFF}
	done
done


