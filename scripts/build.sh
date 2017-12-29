#!/usr/bin/env bash
# build rmd

BASE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source $BASE/go-env
__proj_dir="$(dirname "$__dir")"
__repo_path="github.com/intel/rmd"
BUILD_DATE=${BUILD_DATE:-$( date +%Y%m%d-%H:%M:%S )}

version=$( git describe --tags --dirty --abbrev=14 | sed -E 's/-([0-9]+)-g/.\1+/' )
revision=$( git rev-parse --short HEAD 2> /dev/null || echo 'unknown' )
branch=$( git rev-parse --abbrev-ref HEAD 2> /dev/null || echo 'unknown' )
go_version=$( go version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/' )

echo "git commit: $(git log --pretty=format:"%H" -1)"

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

go build -i -v -ldflags "$ldflags" -o ${build_path}/rmd .  || exit 1
