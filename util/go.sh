#!/usr/bin/env bash

ARCH=`uname | tr '[:upper:]' '[:lower:]'`-amd64
VERSION=${1}

GOBIN=`echo ${PATH} | grep /usr/local/go/bin`
if test -z ${GOBIN}; then
    echo Go Path Notfound. Set 'export PATH=$PATH:/usr/local/go/bin'
    exit 1
fi

if type go >/dev/null 2>&1; then
    CURRENT_VERSION=`go version | awk -F" " '{print $3}'`
fi

if test "${CURRENT_VERSION}" = "go${VERSION}"; then
    echo "Version ${VERSION} is already installed."
    exit 0
fi

# Download and move to /usr/local
curl -L -O -s https://storage.googleapis.com/golang/go${VERSION}.${ARCH}.tar.gz
tar xzf go${VERSION}.${ARCH}.tar.gz
sudo mv go go-${VERSION}
sudo rm -fr /usr/local/go-${VERSION}
sudo mv go-${VERSION} /usr/local
rm go${VERSION}.${ARCH}.tar.gz

# Create symbolic links
sudo rm -fr /usr/local/go
sudo ln -s /usr/local/go-${VERSION} /usr/local/go
go version
