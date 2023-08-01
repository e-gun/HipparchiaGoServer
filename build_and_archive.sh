#!/bin/sh

go get -u ./...
go mod tidy

# CopyInstructions() wants these PDFs, but it can survive without them
# RtEmbPDFHelp() really wants these too; awkward to 404 help files...
# $ pip install mdpdf
# $ npm install mdpdf

if hash mdpdf &> /dev/null
  then
    mdpdf INSTRUCTIONS/INSTALLATION_MACOS_RECOMMENDED.md emb/pdf/HGS_INSTALLATION_MacOS.pdf
    mdpdf INSTRUCTIONS/INSTALLATION_WINDOWS.md emb/pdf/HGS_INSTALLATION_Windows.pdf
    mdpdf INSTRUCTIONS/INSTALLATION_NIX.md emb/pdf/HGS_INSTALLATION_Nix.pdf
    mdpdf INSTRUCTIONS/CUSTOMIZATION.md emb/pdf/HGS_CUSTOMIZATION.pdf
    mdpdf INSTRUCTIONS/SEMANTICVECTORS.md emb/pdf/HGS_SEMANTICVECTORS.pdf
    mdpdf fyi/README.md emb/pdf/HGS_FYI.pdf
    mdpdf INSTRUCTIONS/BASIC_USE.md emb/pdf/HGS_BASIC_USE.pdf
  else
    cp emb/pdf/oops.pdf emb/pdf/HGS_INSTALLATION_MacOS.pdf
    cp emb/pdf/oops.pdf emb/pdf/HGS_INSTALLATION_Windows.pdf
    cp emb/pdf/oops.pdf emb/pdf/HGS_INSTALLATION_Nix.pdf
    cp emb/pdf/oops.pdf emb/pdf/HGS_CUSTOMIZATION.pdf
    cp emb/pdf/oops.pdf emb/pdf/HGS_SEMANTICVECTORS.pdf
    cp emb/pdf/oops.pdf emb/pdf/HGS_FYI.pdf
    cp emb/pdf/oops.pdf emb/pdf/HGS_BASIC_USE.pdf
fi

oss=(linux windows darwin freebsd)
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