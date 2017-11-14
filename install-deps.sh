#!/usr/bin/env bash

if [ ! -f install-deps.sh ]; then
	echo 'This script must be run within its container folder' 1>&2
	exit 1
fi

BASE=$(pwd)
source $BASE/scripts/go-env

go get github.com/Masterminds/glide && glide install
go install github.com/intel/rmd && \
cp -r etc/rmd /etc

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
    ./setup_pam_files.sh $1
else
    ./setup_pam_files.sh
fi

DATA="\"logfile\":\"$LOGFILE\", \"dbtransport\":\"$DBFILE\", \"logtostdout\":false"
go run ./cmd/gen_conf.go -path /etc/rmd/rmd.toml -data "{$DATA}"
