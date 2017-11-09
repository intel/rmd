#!/bin/bash

# TODO add a simple script for functional test.
# All these are hardcode and it only support BDW platform.
# setup PAM files

PAMSRCFILE="etc/rmd/pam/test/rmd"
PAMDIR="/etc/pam.d"
BERKELEYDBFILENAME="rmd_users.db"

BASE=$(pwd)

source $BASE/scripts/go-env

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

if [ "$1" == "-u" ]; then
    godep go test -short -v -cover $(go list ./... | grep -v /vendor/ | grep -v /test/ | grep -v /cmd)
    exit $?
fi

if [ "$1" != "-i" -a "$1" != "-s" ]; then
    godep go test -short -v -cover $(go list ./... | grep -v /vendor/ | grep -v /test/ | grep -v /cmd)
fi

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

# not support -o cdp
mount -t resctrl resctrl /sys/fs/resctrl

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

godep go run ./cmd/gen_conf.go -path ${CONFFILE} -data "{$DATA}"

if [ $? -ne 0 ]; then
    echo "Failed to generate configure file. Exit."
    exit 1
fi

cp -r etc/rmd/policy.toml /tmp/policy.toml

cat $CONFFILE

# Use godep to build rmd binary instead of using dependicies of user's
# GOPATH
godep go install github.com/intel/rmd
if [ $? -ne 0 ]; then
    echo "Failed to build rmd, please correct build issue."
    exit 1
fi

if [ "$1" == "-s" ]; then
    ${GOPATH}/bin/rmd --conf-dir ${CONFFILE%/*} &
else
    ${GOPATH}/bin/rmd --conf-dir ${CONFFILE%/*} --debug &
fi

sleep 1

if [ "$1" == "-s" ]; then
    if [ "$2" == "-nocert" ]; then
        CONF=$CONFFILE ginkgo -v -tags "integrationHTTPS" --focus="PAMAuth" ./test/integrationHTTPS/...
    else
        CONF=$CONFFILE ginkgo -v -tags "integrationHTTPS" --focus="CertAuth" ./test/integrationHTTPS/...
    fi
else
    CONF=$CONFFILE ginkgo -v -tags "integration" ./test/integration/...
fi

rev=$?

# cleanup
kill -TERM `cat $PID`
umount /sys/fs/resctrl
rm ${GOPATH}/bin/rmd

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
exit $rev
