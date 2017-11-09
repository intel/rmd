#!/usr/bin/env bash

PAMSRCFILE="etc/rmd/pam/rmd"
PAMDIR="/etc/pam.d"
if [ -d $PAMDIR ]; then
    cp $PAMSRCFILE $PAMDIR
fi

BERKELEYDBFILE="/etc/rmd/pam/rmd_users.db"

function SetupRMDUsersByResponse {
    if [ $1 == "y" -o $1 == "Y" ]; then
        ./setup_rmd_users.sh
    elif [ $1 != "n" -a $1 != "N" ]; then
        echo "Invalid input. No action taken."
    fi
}

if [ "$1" == "--skip-pam-userdb" ]; then
    SetupRMDUsersByResponse 'n'
    exit 0
fi

if [ -f $BERKELEYDBFILE ]; then
    echo "Do you want to create/update users in RMD Berkeley DB file?(y/n)"
    read -r a
    SetupRMDUsersByResponse $a
else
    SetupRMDUsersByResponse 'y'
fi
