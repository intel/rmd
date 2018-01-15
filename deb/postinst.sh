#!/bin/sh -xe

echo "Post install of rmd"

# todo
# create a soft link of rmd
ln -s /usr/local/sbin/x86_64/rmd  /usr/local/sbin/rmd

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

if [ -f "/lib/systemd/system/rmd.service" ]; then
    systemctl daemon-reload
    systemctl enable rmd.service
    systemctl start rmd.service
fi
