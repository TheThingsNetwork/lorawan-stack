CMD="$1"
HEADER="$2"
FILE="$3"
COMMENT="$4"

LINES=`echo -e ${HEADER} | wc -l`

function check () {
  exit 1
}

function fix () {
  file="$1"
  extension="${file##*.}"
  comment=//

  case "$file" in
      *.make)
        comment=\#
        ;;
      *.sh)
        comment=\#
        ;;
      Makefile)
        comment=\#
        ;;
  esac

  echo -e ${HEADER} | sed 's:^\\(.*) :'"${comment} :g" > file.new && echo >> file.new && cat "$file" >> file.new && mv file.new "$file"
  exit 2
}

ok=1
for i in `seq 1 ${LINES}`; do
  hline=`echo -e ${HEADER} | sed $i'q;d'`
  sed $i'q;d' "$FILE" | grep -q ^$hline$ || ok=0
done
if [[ $ok -ne 1 ]]; then
  case "$CMD" in
    check)
      check "$FILE"
      ;;
    fix)
      fix "$FILE"
      ;;
  esac
fi
