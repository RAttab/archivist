#! /bin/bash -x
set -o errexit -o nounset -o pipefail

go fmt ./...
go build -mod=vendor ./...
go vet ./...

rm -rf /tmp/archivist
mkdir -p /tmp/archivist/usr/local/bin
mkdir -p /tmp/archivist/var/www
cp archivist /tmp/archivist/usr/local/bin
cp assets/* /tmp/archivist/var/www
tar -cf archivist.tar -C /tmp/archivist .

