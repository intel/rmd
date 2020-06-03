#!/bin/bash

PACKAGE_NAME=rmd-extra.pkg.tar.gz

if [[ -z $1 || -z $2 ]]
then
    echo "Usage: $0 <download_path> <output_path>"
    exit 1
fi

DOWNLOAD_PATH=$1
OUTPUT_PATH=$2

echo "Downloading all dependencies to: $DOWNLOAD_PATH"

# create output directory if does not exists
mkdir -p $DOWNLOAD_PATH

# set GOPATH to download directory
export GOPATH=$DOWNLOAD_PATH
# force to enable Go modules (just in case)
export GO111MODULE=on


# check if current directory or parent directory contains go.mod
INPUT=""

if [ -f go.mod ]
then
    INPUT="./go.mod"
elif [ -f ../go.mod]
then
    INPUT="../go.mod"
else
    echo "No go.mod found in current neither parent directory"
    exit 1
fi

echo "Reading packages from: $INPUT"

for line in `grep "^\s\+" $INPUT | sed -e "s/\s\+\(\S\+\) \(\S\+\).*/\1@\2/"`
do
    echo "... fetching $line"
    go get $line
done

# All packages downloaded - create tarbal

START_POINT=`pwd`
# enter parent folder of download path
cd $(dirname $DOWNLOAD_PATH)
# create a .tar.gz file
tar czf $OUTPUT_PATH/$PACKAGE_NAME $(basename $DOWNLOAD_PATH)

echo Done
# show the tarball details (ex. human readable size)
ls -lh $OUTPUT_PATH/$PACKAGE_NAME

# go back to the directory where has been launched
cd $START_POINT

# remove downloaded packages (first change dir modes as go get set them to RO)
chmod -R u+rw $DOWNLOAD_PATH
rm -r $DOWNLOAD_PATH

