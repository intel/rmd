# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.8

# Copy the local package files to the container's workspace.
WORKDIR /go/src/github.com/intel/rmd
COPY . .

RUN apt update && apt install openssl libpam0g-dev db-util -y && rm -rf /var/lib/apt/lists/*
# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get github.com/tools/godep
RUN go run /go/src/github.com/intel/rmd/cmd/get_vendor.go \
        --godeps /go/src/github.com/intel/rmd/Godeps/Godeps.json \
        --vendor /go/src/github.com/intel/rmd/vendor

RUN godep go install github.com/intel/rmd

RUN ./install-deps --skip-pam-userdb
RUN ./test.sh -u

# what etc should we use?
# log

# Run the outyet command by default when the container starts.
ENTRYPOINT [ "/go/bin/rmd" ]

# Document that the service listens on port 8080.
# EXPOSE 8080
