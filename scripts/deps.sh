#!/usr/bin/env bash

BASE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
. $BASE/go-env

go get github.com/Masterminds/glide && glide install
