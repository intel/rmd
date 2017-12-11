# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.8

# Copy the local package files to the container's workspace.
WORKDIR /go/src/github.com/intel/rmd
COPY . .

RUN apt update && apt install openssl libpam0g-dev db-util -y && \
        rm -rf /var/lib/apt/lists/*
RUN make install && make clean
RUN make test-unit

# what etc should we use?
# log

# Run the outyet command by default when the container starts.
ENTRYPOINT [ "/go/bin/rmd" ]

# Document that the service listens on port 8080.
# EXPOSE 8080
