#!/bin/bash
# a script to fetch Go dependencies.
# when necessary, we use backup repos with specific tag/branches.
# TODO: remove this script once deb packaging work is complete and
# used in nightly.

set -xe

function git_get {
    git clone -q $1 $2
    pushd $2
    git checkout $3
    popd
}

# fetched during install.sh. FIXME
# git clone -q gitolite@gitolite.activestate.com:tail 			$GOPATH/src/github.com/srid/tail
echo "WARNING: tail.git is expected to be cloned manually"


git_get https://github.com/ActiveState/doozer $GOPATH/src/github.com/ActiveState/doozer d3af058c2
git_get https://github.com/kr/pretty $GOPATH/src/github.com/kr/pretty 821b30f5
git_get https://github.com/nu7hatch/gouuid $GOPATH/src/github.com/nu7hatch/gouuid 0345199
git_get https://github.com/howeyc/fsnotify $GOPATH/src/github.com/howeyc/fsnotify d6220df
git_get https://github.com/apcera/nats $GOPATH/src/github.com/apcera/nats dd857f76
git_get https://github.com/srid/doozerconfig $GOPATH/src/github.com/srid/doozerconfig 49819652
git_get https://github.com/srid/log2 $GOPATH/src/github.com/srid/log2 889405d0
git_get https://github.com/fzzbt/radix $GOPATH/src/github.com/fzzbt/radix 7687e823
git_get https://github.com/alecthomas/gozmq $GOPATH/src/github.com/alecthomas/gozmq 965ec0982

hg -q clone https://code.google.com/p/goprotobuf/ $GOPATH/src/code.google.com/p/goprotobuf
pushd $GOPATH/src/code.google.com/p/goprotobuf
hg update -r 29db96746e5d
popd

# XXX: need to be mirrored (bug 95841)
mkdir -p $GOPATH/src/launchpad.net
bzr branch -q lp:tomb $GOPATH/src/launchpad.net/tomb
