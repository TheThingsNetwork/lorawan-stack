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
sedArgs=()
lines=()
genFiles=()

IFS_BAK=${IFS}
IFS="
"
for f in ${protos[@]}; do
  grep -q '(google.api.http)' ${f} && genFiles+=(${f%".proto"}".pb.gw.go")
  grep -q '(gogoproto.customname)' ${f} && lines+=(`cat ${f} | grep '(gogoproto.customname)'`) || continue

  for l in ${lines[@]}; do
    from=`echo $l | sed 's/[ ]*\(repeated[ ]\+\)\?[[:alnum:]_.]\+[ ]\+\([[:alnum:]_]\+\)[ ]*=[ ]*[0-9]\+.*/\2/' | sed 's/_\([a-z]\)/\u\1/g' | tr -d ' ' | sed 's/\(^[:a-z:]\)\(.*\)/\u\1\2/' | tr -d '_' `
    to=`echo $l | sed 's/.*(gogoproto.customname)[  ]*=[ ]*"\([[:alnum:]]\+\)".*/\1/' | tr -d ' '`
    sedArgs+=("-e s/\([^[:alnum:]]\)${from}\([^[:alnum:]]\)/\1${to}\2/")
  done
done
IFS=${IFS_BAK}

if [[ ${#genFiles[@]} != 0 ]] && [[ ${#sedArgs[@]} != 0 ]]; then
    sed -i ${sedArgs[*]} ${genFiles[*]}
fi
exit 0
