#!/bin/sh

go get -u ./...
go mod tidy

# CopyInstructions() wants these PDFs, but it can survive without them
if hash mdpdf &> /dev/null
  then
    mdpdf INSTRUCTIONS/INSTALLATION_MACOS_RECOMMENDED.md emb/pdf/HGS_INSTALLATION_MacOS.pdf
    mdpdf INSTRUCTIONS/INSTALLATION_WINDOWS.md emb/pdf/HGS_INSTALLATION_Windows.pdf
    mdpdf INSTRUCTIONS/CUSTOMIZATION.md emb/pdf/HGS_Customization.pdf
    mdpdf fyi/README.md emb/pdf/HGS_FYI.pdf
fi

oss=(linux windows darwin)
archs=(amd64 arm64)

P="HipparchiaGoServer"
SUFF=""
OUT="./bin"

VS=""
if [ `git branch --show-current` != "stable" ]; then
  VS="-pre"
fi

DT=$(date "+%Y-%m-%d@%H:%M:%S")
GC=$(git rev-list -1 HEAD | cut -c-8)

go build -pgo=default.pgo -ldflags "-s -w -X main.GitCommit=${GC} -X main.BuildDate=${DT} -X main.VersSuppl=${VS}"
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
    EXE=${P}-${V}-${os}-${arch}${SUFF}
	  env GOOS=${os} GOARCH=${arch} go build -pgo=default.pgo -ldflags "-s -w -X main.GitCommit=${GC} -X main.BuildDate=${DT}" -o ${P}${SUFF}
	  zip -q ${EXE}.zip ${P}${SUFF}
	  mv ${EXE}.zip ${OUT}/
	  rm ${P}${SUFF}
	done
done


