#!/usr/bin/env bash
set -e

function usage {
  echo "Usage: $(basename "$0") <proto-dir>"
  exit 1
}
[[ ${1} = "" ]] && usage
protoDir="${1}"
basePackage="go.thethings.network/lorawan-stack"

# TODO figure out a way to traverse only files imported by protos using grpc-gateway (google.api.http option)
protos=( ${protoDir}/*.proto )
perlArgs=()
genPaths=()

IFS_BAK=${IFS}
IFS="
"

for f in "${protos[@]}"; do
  if grep -q '(google.api.http)' "${f}"; then
    path=${f%".proto"}".pb.gw.go"
    if grep -q 'option go_package' "${f}"; then
      goPackage=$(grep 'option go_package' "${f}" | perl \
        -pe 's![[:space:]]*option[[:space:]]+go_package[[:space:]]*=[[:space:]]*"'${basePackage}'/([[:alnum:]_.\-/]+)".*!\1!')
      path=${goPackage}/$(basename "${path}")
    fi
    genPaths+=( "${path}" )
  fi

  if grep -q '(google\.protobuf\.Empty)' "${f}"; then
    perlArgs+=( "-pe s!empty\\.Empty!types.Empty!g;" )
  fi

  if grep -q '(gogoproto.customname)' "${f}"; then
    for l in $(grep '(gogoproto.customname)' "${f}"); do
      from=$(echo "${l}" | perl \
        -pe 's![[:space:]]*(repeated[[:space:]]+)?([[:alnum:]_.]+|map<[[:space:]]*[[:alnum:]]+[[:space:]]*,[[:space:]]*[[:alnum:]_.]+[[:space:]]*>)[[:space:]]+([[:alnum:]_]+)[[:space:]]*=[[:space:]]*[0-9]+.*!\3!;' \
        -pe 's!(^[[:alnum:]])([[:alnum:]]*)|_([[:alnum:]])([[:alnum:]]*)!\U\1\3\E\2\4!g;')
      to=$(echo "${l}" | perl \
        -pe 's!.*\(gogoproto.customname\)[[:space:]]*=[[:space:]]*"([[:alnum:]_]+)".*!\1!')
      ! [ "${from}" = "${to}" ] && perlArgs+=("-pe s!([^[:alnum:]])${from}([^[:alnum:]])!\\1${to}\\2!g;")
    done
  fi
done
IFS=${IFS_BAK}

if [[ ${#perlArgs[@]} != 0 ]]; then
  for f in "${genPaths[@]}"; do
    perl -i -p ${perlArgs[*]} "${f}"
  done
fi
exit 0
