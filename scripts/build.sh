#!/usr/bin/env bash
# build rmd
BASE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source $BASE/go-env
__proj_dir="$(dirname "$__dir")"

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
go build -o "${build_path}/rmd" . || exit 1
