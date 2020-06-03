#!/usr/bin/env bash
GO_MINOR_VERSION=$(go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)

if [ ${GO_MINOR_VERSION} -lt 11 ]; then
	echo "unsupported go version. require >= go1.11"
	exit 1
fi

export GO111MODULE=on
BASE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source $BASE/go-env
#Copy only for git version. For package, this will be done from spec file
if [ -d $BASE/../.git ];
then
cp -r $BASE/../etc/rmd /etc
fi

USER="rmd"
useradd $USER || echo "User rmd already exists."

LOGFILE="/var/log/rmd/rmd.log"
if [ ! -d ${LOGFILE%/*} ]; then
    mkdir -p ${LOGFILE%/*}
    chown $USER:$USER ${LOGFILE%/*}
fi

DBFILE="/var/run/rmd/rmd.db"
if [ ! -d ${DBFILE%/*} ]; then
    mkdir -p ${DBFILE%/*}
    chown $USER:$USER  ${DBFILE%/*}
fi


if [ "$1" == "--skip-pam-userdb" ]; then
    $BASE/setup_pam_files.sh $1
else
    $BASE/setup_pam_files.sh
fi

DATA="\"logfile\":\"$LOGFILE\", \"dbtransport\":\"$DBFILE\", \"logtostdout\":false"
gen_conf -path /etc/rmd/rmd.toml -data "{$DATA}"

if [ -f "/lib/systemd/system/rmd.service" ]; then
    systemctl daemon-reload
    systemctl enable rmd.service
    systemctl start rmd.service
fi