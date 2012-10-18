GOBIN    	= $(shell pwd)/bin
VM	    	= zzw5

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

install:	fmt bin/logyard bin/send bin/recv bin/systail

doozer:
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v github.com/ActiveState/doozer/cmd/doozer


# FIXME: the go tool should be doing this.

$(GOPATH)/pkg/*/logyard.a:	*.go
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard

bin/logyard:	cmd/logyard/*.go $(GOPATH)/pkg/*/logyard.a
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard/cmd/logyard

bin/send:	cmd/send/*.go $(GOPATH)/pkg/*/logyard.a
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard/cmd/send

bin/recv:	cmd/recv/*.go $(GOPATH)/pkg/*/logyard.a
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard/cmd/recv

bin/systail:	cmd/systail/*.go $(GOPATH)/pkg/*/logyard.a
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install -v logyard/cmd/systail

push:	fmt
	rsync -4 -rtv . stackato@stackato-$(VM).local:/s/logyard/ --exclude .git --exclude bin

fmt:
	gofmt -w .

clean: 
	GOPATH=$(GOPATH) go clean
	rm -rf ./bin
