#!/bin/sh
scriptdir=$(cd $(dirname $0); pwd)
cd $scriptdir/..

set -e
set -x

# building for host arch (x86_64), not adding spacemonkey. Adding "simulation"
# build tag.  This does file level conditional build and includes any .go files
# with //build +simulation comments at the beginning, before the package
# statement.

binaries=${binaries:-$(find cmd/* -type d)}
TAGS=${TAGS:-ec mockup}

for cmd in ${binaries}
do
    go build -tags "simulation $TAGS" "$@" github.com/superchalupa/sailfish/$cmd
done
echo

set +x
for b in ${binaries}
do
    echo -e "BUILD SUCCESS. binary ready: $(basename $b)"
done
