#!/usr/bin/env bash
set -e

function usage {
  echo "Usage: `basename $0` <proto-dir>"
  exit 1
}
[[ ${1} = "" ]] && usage
protoDir=${1}

# TODO figure out a way to traverse only files imported by protos using grpc-gateway (google.api.http option)
protos=(${protoDir}/*.proto)
perlArgs=()
lines=()
genPaths=()

IFS_BAK=${IFS}
IFS="
"
for f in ${protos[@]}; do
  if grep -q '(google.api.http)' ${f}; then
    path=${f%".proto"}".pb.gw.go"
    if grep -q 'option go_package' ${f}; then
      goPackage=`grep 'option go_package' ${f} | perl -pe 's/[[:space:]]*option[[:space:]]\+go_package[[:space:]]\+=[[:space:]]*"\([[:alnum:]./]\+\)".*/\1/'`
      newPath=${GOPATH:-"${HOME}/go"}"/src/"${goPackage}/`basename ${path}`
      mv ${path} ${newPath}
      path=${newPath}
    fi
    genPaths+=(${path})
  fi

  if grep -q '(gogoproto.customname)' ${f}; then
    for l in `grep '(gogoproto.customname)' ${f}`; do
      from=`echo $l | perl -pe 's/[[:space:]]*\(repeated[[:space:]]\+\)\?[[:alnum:]_.]\+[[:space:]]\+\([[:alnum:]_]\+\)[[:space:]]*=[[:space:]]*[0-9]\+.*/\2/' | perl -pe 's/_\([a-z]\)/\u\1/g' | tr -d ' ' | perl -pe 's/\(^[:a-z:]\)\(.*\)/\u\1\2/' | tr -d '_' `
      to=`echo $l | perl -pe 's/.*(gogoproto.customname)[[:space:]]*=[[:space:]]*"\([[:alnum:]]\+\)".*/\1/' | tr -d ' '`
      perlArgs+=("-e s/\([^[:alnum:]]\)${from}\([^[:alnum:]]\)/\1${to}\2/")
    done
  fi
done
IFS=${IFS_BAK}

if [[ ${#perlArgs[@]} != 0 ]]; then
  for f in ${genPaths[@]}; do
    perl -i -p ${perlArgs[*]} ${f}
  done
fi
exit 0
