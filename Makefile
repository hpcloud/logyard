VM	    	= ny8r

default:	install

setup:	clean setup-repos

# temporary workaround until we have a generic replacement for `go get -u`
# with support for alternate repo paths.
# TODO: write a version of setup-repos that clones dev (git-rw)
# branches to ~/as/ and create a symlink to it from GOPATH/src/
setup-repos:
	tools/fetch-dependencies.sh
	# pull rest of the dependencies and build them
	go get -v logyard/...

install:	fmt installall

doozer:
	GOPATH=$(GOPATH) go install -v github.com/ActiveState/doozer/cmd/doozer

installall:
	GOPATH=$(GOPATH) go install -v logyard/...

push:	fmt
	rsync -4 -rtv . stackato@stackato-$(VM).local:/s/go/src/logyard/ --exclude .git

# compile and push; best used from within emacs (M-x compile)
cpush:	install push

fmt:
	gofmt -w .

test:
	go test -v logyard/...

clean: 
	GOPATH=$(GOPATH) go clean
