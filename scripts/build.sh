#!/usr/bin/env bash
# build rmd

BASE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source $BASE/go-env
__proj_dir="$(dirname "$__dir")"
__repo_path="github.com/intel/rmd"
BUILD_DATE=${BUILD_DATE:-$( date +%Y%m%d-%H:%M:%S )}

# Use version number hardcoded in file
version=$(cat ./RMD_VERSION)
if [ -d ./.git ];
then
# Building from git working directory
# Use data from repo (ex. add build number and commit id)
    version+=$(git describe --tags --dirty --abbrev=14 | cut -f"3,4" -d"-" | sed -E 's/^g/+/' )
    revision=$(git rev-parse --short HEAD 2> /dev/null || echo 'unknown' )
    branch=$(git rev-parse --abbrev-ref HEAD 2> /dev/null || echo 'unknown' )
else
# Building from tarball/zip
    revision=""
    branch=""
fi


go_version=$( go version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/' )

GO_MINOR_VERSION=$( go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)

if [[ ${GO_MINOR_VERSION} -ge 11 ]]; then
	export GO111MODULE=on
fi

if [ -d ./.git ];
then
    echo "git commit: $(git log --pretty=format:"%H" -1)"
else
    echo "building from tarbal"
fi

# rebuild binaries:
export GOOS=${GOOS:-$(go env GOOS)}
export GOARCH=${GOARCH:-$(go env GOARCH)}

if [[ "${GOARCH}" == "amd64" ]]; then
    build_path="${__proj_dir}/build/${GOOS}/x86_64"
else
    build_path="${__proj_dir}/build/${GOOS}/${GOARCH}"
fi

mkdir -p "${build_path}"
echo "building rmd for ${GOOS}/${GOARCH}"

ldflags="
    -X ${__repo_path}/version.Version=${version}
    -X ${__repo_path}/version.Revision=${revision}
    -X ${__repo_path}/version.Branch=${branch}
    -X ${__repo_path}/version.BuildDate=${BUILD_DATE}
    -X ${__repo_path}/version.GoVersion=${go_version}"

# load code for setting GOBUILDOPTS variable based on command line params
source $BASE/build-opts-get

go build -i -v $GOBUILDOPTS -ldflags "$ldflags" -o ${build_path}/rmd ./cmd/rmd  || exit 1
go build -i -v $GOBUILDOPTS -ldflags=-linkmode=external -o ${build_path}/gen_conf $BASE/../cmd/gen_conf/gen_conf.go  || exit 1
