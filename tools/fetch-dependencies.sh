#!/bin/bash
# a script to fetch Go dependencies.
# when necessary, we use backup repos with specific tag/branches.

set -xe

PBF=http://mercurial.activestate.com/stackato-mirrors/goprotobuf/
mkdir -p $GOPATH/src/launchpad.net

# fetched during install.sh. FIXME
# git clone -q gitolite@gitolite.activestate.com:tail 			$GOPATH/src/github.com/srid/tail

bzr branch -q lp:tomb 										$GOPATH/src/launchpad.net/tomb
git clone -q https://github.com/ActiveState/fsnotify.git 	$GOPATH/src/github.com/howeyc/fsnotify
hg -q clone $PBF  $GOPATH/src/code.google.com/p/goprotobuf
git clone -q https://github.com/ActiveState/doozer 			$GOPATH/src/github.com/ActiveState/doozer
git clone -q https://github.com/ActiveState/pretty 			$GOPATH/src/github.com/kr/pretty
git clone -q https://github.com/ActiveState/gouuid	 		$GOPATH/src/github.com/nu7hatch/gouuid

#git clone -q https://github.com/ActiveState/nats    $GOPATH/src/github.com/apcera/nats
git clone -q https://github.com/apcera/nats	 		$GOPATH/src/github.com/apcera/nats



#git clone -q https://github.com/ActiveState/doozerconfig 	$GOPATH/src/github.com/ActiveState/doozerconfig
#git clone -q https://github.com/ActiveState/colors.git      $GOPATH/src/github.com/wsxiaoys/colors

git clone -q https://github.com/fzzbt/radix	 		$GOPATH/src/github.com/fzzbt/radix
#git clone -q https://github.com/ActiveState/radix	 		$GOPATH/src/github.com/fzzbt/radix
