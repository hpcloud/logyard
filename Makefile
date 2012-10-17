GOBIN    	= $(shell pwd)/bin
VM	    	= k6jh

default:	install

# temporary workaround until we have a generic replacement for `go get -u`
# with support for alternate repo paths.
setup:	clean setup-repos setup-prepare

# TODO: write a version of setup-repos that clones dev (git-rw)
# branches to ~/as/ and create a symlink to it from GOPATH/src/
setup-repos:
	mkdir -p $(GOPATH)/src/launchpad.net
	bzr branch -q lp:tomb 										$(GOPATH)/src/launchpad.net/tomb
	hg -q clone http://mercurial.activestate.com/stackato-mirrors/goprotobuf/	 		$(GOPATH)/src/code.google.com/p/goprotobuf
	# XXX: requires ssh keys
	git clone -q git@github.com:ActiveState/tail.git 			$(GOPATH)/src/github.com/ActiveState/tail
	git clone -q https://github.com/ActiveState/doozer 			$(GOPATH)/src/github.com/ActiveState/doozer
	#git clone -q https://github.com/ActiveState/doozerconfig 	$(GOPATH)/src/github.com/ActiveState/doozerconfig
	git clone -q https://github.com/ActiveState/pretty 			$(GOPATH)/src/github.com/kr/pretty
	#git clone -q https://github.com/ActiveState/nats	 		$(GOPATH)/src/github.com/apcera/nats
	#git clone -q https://github.com/ActiveState/radix	 		$(GOPATH)/src/github.com/fzzbt/radix
	#git clone -q https://github.com/ActiveState/gouuid	 		$(GOPATH)/src/github.com/nu7hatch/gouuid
	#git clone -q https://github.com/ActiveState/colors.git      $(GOPATH)/src/github.com/wsxiaoys/colors
	#git clone -q https://github.com/ActiveState/fsnotify.git 	$(GOPATH)/src/github.com/howeyc/fsnotify

setup-prepare:
	ln -sf `pwd` $(GOPATH)/src/
	mkdir -p bin

install:	bin/logyard

doozer:
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v github.com/ActiveState/doozer/cmd/doozer

bin/logyard:	*.go cmd/*/*.go
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard/cmd/logyard

push:	fmt
	rsync -4 -rtv . stackato@stackato-$(VM).local:/s/vcap/logyard/ --exclude .git --exclude bin

fmt:
	gofmt -w .

clean: 
	GOPATH=$(GOPATH) go clean
	rm -rf ./bin
