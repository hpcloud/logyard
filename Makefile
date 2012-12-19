VM	    	= kny7

default:	install

setup:	clean setup-repos

setup-repos:
	GOPATH=$(GOPATH) goget

install:	fmt installall

doozer:
	GOPATH=$(GOPATH) go install -v github.com/ActiveState/doozer/cmd/doozer

installall:
	GOPATH=$(GOPATH) go install -v logyard/... github.com/ActiveState/tail/cmd/gotail

install_doozerd:
	GOPATH=$(GOPATH) go get -v github.com/ActiveState/doozerd

push:	fmt
	rsync -4 -rtv . stackato@stackato-$(VM).local:/s/go/src/logyard/ --exclude .git

pushdeps:
	rsync -4 -rtv $(GOPATH)/src/github.com stackato@stackato-$(VM).local:/s/go/src/ --exclude .git

# compile and push; best used from within emacs (M-x compile)
cpush:	install push

fmt:
	gofmt -w .

test:
	go test -v logyard/... github.com/ActiveState/log

clean: 
	GOPATH=$(GOPATH) go clean
