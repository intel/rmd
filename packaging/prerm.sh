#!/bin/sh -xe

if [ -f "/lib/systemd/system/rmd.service" ]; then
    systemctl stop rmd.service
    systemctl disable rmd.service
fi
