#!/bin/sh

go get -u ./...
go mod tidy

VS=""
if [ `git branch --show-current` != "stable" ]; then
  VS="-pre"
fi

DT=$(date "+%Y-%m-%d@%H:%M:%S")
GC=$(git rev-list -1 HEAD | cut -c-8)
PG="default.pgo"
PGF="./pgo/${PG}"

LDF="-s -w -X main.GitCommit=${GC} -X main.BuildDate=${DT} -X main.VersSuppl=${VS} -X main.PGOInfo=${PG}"

# i.e., call with anything at all after the script name and you will just stop here: one and done
# but note that INSTRUCTIONS will not have been generated
if test -n "$1"; then
  go build -pgo=${PGF} -ldflags "${LDF}"
  exit
fi

# CopyInstructions() wants these PDFs, but it can survive without them
# RtEmbPDFHelp() really wants these too; awkward to 404 help files...
# $ pip install mdpdf
# $ npm install mdpdf

if hash mdpdf &> /dev/null
  then
    mdpdf INSTRUCTIONS/INSTALLATION_MACOS_RECOMMENDED.md web/emb/pdf/HGS_INSTALLATION_MacOS.pdf
    mdpdf INSTRUCTIONS/INSTALLATION_WINDOWS.md web/emb/pdf/HGS_INSTALLATION_Windows.pdf
    mdpdf INSTRUCTIONS/INSTALLATION_NIX.md web/emb/pdf/HGS_INSTALLATION_Nix.pdf
    mdpdf INSTRUCTIONS/CUSTOMIZATION.md web/emb/pdf/HGS_CUSTOMIZATION.pdf
    mdpdf INSTRUCTIONS/SEMANTICVECTORS.md web/emb/pdf/HGS_SEMANTICVECTORS.pdf
    mdpdf INSTRUCTIONS/fyi/README.md web/emb/pdf/HGS_FYI.pdf
    mdpdf INSTRUCTIONS/BASIC_USE.md web/emb/pdf/HGS_BASIC_USE.pdf
    cp web/emb/pdf/*.pdf internal/lnch/efs/pdf/
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

# need this build just to set ${V} in the next line
go build -pgo=${PGF} -ldflags "${LDF}"
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
	  env GOOS=${os} GOARCH=${arch} go build -pgo=${PGF} -ldflags "${LDF}" -o ${P}${SUFF}
	  zip -q ${EXE}.zip ${P}${SUFF}
	  mv ${EXE}.zip ${OUT}/
	  rm ${P}${SUFF}
	done
done