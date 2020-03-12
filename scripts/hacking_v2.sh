#!/bin/bash
function hacking() {
BASE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
cd $BASE/..
echo $BASE
echo " "
set -x
LINTOUT=$(golint ./...)
set +x
RETGOLINT=0
if [[ ! -z $LINTOUT ]]; then
#	echo " <<< golint errors: "
#	echo -e ${LINTOUT}
#	echo " <<< "
	echo ":( <<< Please address golint errors."
	RETGOLINT=1
else
	echo ":) >>> No errors from golint"
fi

echo " "
set -x
FMTOUT=$(gofmt -l ./)
set +x
RETGOFMT=0
if [[ ! -z $FMTOUT ]]; then
	echo " <<< gofmt: files that have formatting issues: "
	gofmt -d ./
	echo " <<< "
	echo ":( <<< Please address gofmt errors."
	RETGOFMT=1
else
	echo ":) >>> No errors from gofmt"
fi

echo " "
set -x
VETOUT=$(go vet ./...)
set +x
RETGOVET=0
if [[ $? -ne 0 ]]; then
	echo " <<< go vet errors: "
	echo -e ${VETOUT}
	echo " <<< "
	echo ":( <<< Please address go vet errors."
	RETGOVET=1
else
	echo ":) >>> No errors from go vet"
fi

echo " "
echo " >>> Done checking"
echo " "

RET=0
if [[ $RETGOLINT -ne 0 ]] || [[ $RETGOFMT -ne 0 ]] || [[ $RETGOVET -ne 0 ]]; then
	RET=1
fi
echo "return $RET"
return $RET
}

hacking

exit $?

