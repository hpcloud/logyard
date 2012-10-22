GOBIN    	= $(shell pwd)/bin
VM	    	= sf4r

default:	install

setup:	clean setup-repos setup-prepare

# temporary workaround until we have a generic replacement for `go get -u`
# with support for alternate repo paths.
# TODO: write a version of setup-repos that clones dev (git-rw)
# branches to ~/as/ and create a symlink to it from GOPATH/src/
setup-repos:
	tools/fetch-dependencies.sh

setup-prepare:
	mkdir -p bin
	# pull rest of the dependencies and build them
	go get -v logyard/...

install:	fmt installall

doozer:
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v github.com/ActiveState/doozer/cmd/doozer

installall:
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard/...

push:	fmt
	rsync -4 -rtv . stackato@stackato-$(VM).local:/s/go/src/logyard/ --exclude .git --exclude bin

fmt:
	gofmt -w .

clean: 
	GOPATH=$(GOPATH) go clean
	rm -rf ./bin
