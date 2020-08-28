#!/bin/bash

# TODO add a simple script for functional test.
# All these are hardcode and it only support BDW platform.
# setup PAM files
BASE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
PAMSRCFILE="$BASE/../etc/rmd/pam/test/rmd"
PAMDIR="/etc/pam.d"
BERKELEYDBFILENAME="rmd_users.db"
__proj_dir="$(dirname "$__dir")"

GO_MINOR_VERSION=$(go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)

# rebuild binaries:
export GOOS=${GOOS:-$(go env GOOS)}
export GOARCH=${GOARCH:-$(go env GOARCH)}

if [[ "${GOARCH}" == "amd64" ]]; then
    build_path="${__proj_dir}/build/${GOOS}/x86_64"
else
    build_path="${__proj_dir}/build/${GOOS}/${GOARCH}"
fi

if [ ${GO_MINOR_VERSION} -ge 11 ]; then
        export GO111MODULE=on
fi

source $BASE/go-env

if [ -d $PAMDIR ]; then
    cp $PAMSRCFILE $PAMDIR
fi

# setup PAM test user
echo "user" >> users
openssl passwd -crypt "user1" >> users
echo "test" >> users
openssl passwd -crypt "test1" >> users

db_load -T -t hash -f users "/tmp/"$BERKELEYDBFILENAME
if [ $? -ne 0 ]; then
    rm -rf users
    echo "Failed to setup pam files"
    exit 1
fi

rm -rf users

# load code for setting GOBUILDOPTS variable based on command line params
source $BASE/build-opts-get

# replacement for old glide usage: glide novendor | grep -v /test
DIRS_TO_TEST=`for ff in \`find . -name "*.go" | cut -f2 -d"/" | grep -v '/test/' | grep -v '/vendor/' | sort -u\`; do echo "./$ff/..."; done`

cd $BASE/..
if [ "$1" == "-u" ]; then
    mount -t resctrl resctrl /sys/fs/resctrl
    go test $GOBUILDOPTS -short -v -cover $DIRS_TO_TEST
    exit $?
fi

if [ "$1" != "-i" -a "$1" != "-s" ]; then
    mount -t resctrl resctrl /sys/fs/resctrl
    go test $GOBUILDOPTS -short -v -cover $DIRS_TO_TEST
fi
cd -

RESDIR="/sys/fs/resctrl"
PID="/var/run/rmd.pid"
CONFFILE="/tmp/rmd.toml"

if [ -f "$PID" ]; then
    pid=`cat "$PID"`
    if [ -n "$pid" ]; then
        if [ -d "/proc/$pid" ]; then
            echo "RMD: $pid is already running. exit!"
            exit 1
        fi
    fi
fi

# clean up, force remove resctrl
if [ -d "$RESDIR" ] && mountpoint $RESDIR > /dev/null 2>&1 ; then
    umount /sys/fs/resctrl
    if [ $? -ne 0 ]; then
        echo "--------------------------------------------------"
        echo "Please unmount /sys/fs/resctrl manually"
        echo "It is used by these processes:"
        lsof "$RESDIR"
        exit 1
    fi
fi

# to have "clean" situation before functional tests
umount /sys/fs/resctrl

# Set a unused random port
CHECK="do while"
while [[ ! -z $CHECK ]]; do
    PORT=$(( ( RANDOM % 60000 )  + 1025 ))
    CHECK=$(netstat -ap | grep $PORT)
done

DATA=""

if [ "$1" == "-s" ]; then
    if [ "$2" == "-nocert" ]; then
        DATA="\"clientauth\":\"no\", \"tlsport\":$PORT"
    else
        DATA="\"tlsport\":$PORT"
    fi
else
    DATA="\"debugport\":$PORT"
fi

DATA="$DATA, \"policypath\":\"/tmp/policy.toml\", \"dbtransport\":\"/tmp/rmd.db\", \"stdout\":false, \"logfile\":\"/tmp/rmd.log\""

go run $GOBUILDOPTS $BASE/../cmd/gen_conf/gen_conf.go -path ${CONFFILE} -data "{$DATA}"

if [ $? -ne 0 ]; then
    echo "Failed to generate configure file. Exit."
    exit 1
fi

cp -r $BASE/../etc/rmd/policy.toml /tmp/policy.toml

cat $CONFFILE

cp ${build_path}/rmd .

if [ "$1" == "-s" ]; then
    ./rmd --conf-dir ${CONFFILE%/*} &
else
    ./rmd --conf-dir ${CONFFILE%/*} -d &
fi

sleep 1

if [ "$1" == "-s" ]; then
    if [ "$2" == "-nocert" ]; then
        CONF=$CONFFILE ${GOPATH}/bin/ginkgo -v -tags "integrationHTTPS" --focus="PAMAuth" ./test/integrationHTTPS
    else
        CONF=$CONFFILE ${GOPATH}/bin/ginkgo -v -tags "integrationHTTPS" --focus="CertAuth" ./test/integrationHTTPS
    fi
else
    CONF=$CONFFILE ${GOPATH}/bin/ginkgo -v -tags "integration" ./test/integration
fi

rev=$?

# cleanup
kill -TERM `cat $PID`
umount /sys/fs/resctrl


# cleanup PAM files
if [ "$1" == "-s" -a "$2" == "-nocert" ]; then
    rm -rf "/tmp/"$BERKELEYDBFILENAME
    rm -rf $PAMDIR"/rmd"
fi

if [[ $rev -ne 0 ]]; then
    echo ":( <<< Functional testing fail, retual value $rev ."
else
    echo ":) >>> Functional testing passed ."
fi
rm rmd
exit $rev
