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
	# treat the current project as a go import.
	ln -sf `pwd` $(GOPATH)/src/
	mkdir -p bin
	# pull rest of the dependencies and build them
	go get -v logyard 

install:	fmt installall

doozer:
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v github.com/ActiveState/doozer/cmd/doozer

# as i can never understand the behaviour of go build/install, it
# would be better take a painful shortcut for now and build all files
# without caring for modified files.
installall:
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard/drain
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard/stackato
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard/cmd/logyard
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard/cmd/send
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard/cmd/recv
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard/cmd/systail
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard/cmd/apptail

push:	fmt
	rsync -4 -rtv . stackato@stackato-$(VM).local:/s/vcap/logyard/ --exclude .git --exclude bin

fmt:
	gofmt -w .

clean: 
	GOPATH=$(GOPATH) go clean
	rm -rf ./bin
