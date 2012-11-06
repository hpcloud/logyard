#!/bin/bash
# a script to fetch Go dependencies.
# when necessary, we use backup repos with specific tag/branches.

set -xe

PBF=http://mercurial.activestate.com/stackato-mirrors/goprotobuf/
mkdir -p $GOPATH/src/launchpad.net

# fetched during install.sh. FIXME
# git clone -q gitolite@gitolite.activestate.com:tail 			$GOPATH/src/github.com/srid/tail
echo "WARNING: tail.git is expected to be cloned manually"

# XXX: need to be mirrored (bug 95841)
bzr branch -q lp:tomb 										$GOPATH/src/launchpad.net/tomb

hg -q clone $PBF  $GOPATH/src/code.google.com/p/goprotobuf

git clone -q https://github.com/ActiveState/doozer 			$GOPATH/src/github.com/ActiveState/doozer
git clone -q https://github.com/ActiveState/pretty 			$GOPATH/src/github.com/kr/pretty
git clone -q https://github.com/ActiveState/gouuid	 		$GOPATH/src/github.com/nu7hatch/gouuid

# XXX: our mirrors are outdated; we trust the owners of these repo.
#git clone -q https://github.com/ActiveState/fsnotify.git 	$GOPATH/src/github.com/howeyc/fsnotify
#git clone -q https://github.com/ActiveState/nats    $GOPATH/src/github.com/apcera/nats
#git clone -q https://github.com/ActiveState/doozerconfig 	$GOPATH/src/github.com/srid/doozerconfig
#git clone -q https://github.com/ActiveState/radix	 		$GOPATH/src/github.com/fzzbt/radix
