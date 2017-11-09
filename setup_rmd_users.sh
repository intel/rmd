#!/usr/bin/env bash

if [[ $EUID != 0 ]]; then
  echo "This script can be run only by root"
  exit 1
fi

BERKELEYDBDIR="/etc/rmd/pam"
BERKELEYDBFILENAME="rmd_users.db"
echo 'Setup Berkeley db users'
while true
do
    echo 'Enter username or 0 to stop'
    read u
    if [ $u == "0" ]; then
        break
    fi
    echo $u >> users
    echo 'Enter password:'
    read -s p
    openssl passwd -crypt $p >> users
done

# If input file was created
if [ -f "users" ]; then
    mkdir -p $BERKELEYDBDIR
    # Berkeley DB is access restricted to root only
    db_load -T -t hash -f users $BERKELEYDBDIR"/"$BERKELEYDBFILENAME
    rm -rf users
fi
