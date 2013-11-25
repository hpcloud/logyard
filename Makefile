#
# Makefile for stackato-logyard-ng
#
# Used solely by packaging systems.
# Must support targets "all", "install", "uninstall".
#
# During the packaging install phase, the native packager will
# set either DESTDIR or prefix to the directory which serves as
# a root for collecting the package files.
#
# Additionally, stackato-pkg sets STACKATO_PKG_BRANCH to the
# current git branch of this package, so that we may use it to
# fetch other git repos with the corresponding branch.
#
# The resulting package installs in /home/stackato,
# is not intended to be relocatable.
#
# This package depends on external data.  Run "make update" to update the
# local copy of that data.  Push any resulting changes to the git repo
# in order to trigger generation of a new package.
#
# To locally test this Makefile, run:
#
#   rm -rf .gopath; STACKATO_PKG_BRANCH=mybranch make
#

NAME=logyard

SRCDIR=src/$(NAME)

COMMON_REPO=git://gitolite.activestate.com/stackato-common.git

UPDATE=.stackato-pkg/update
COMMON_DIR=$(UPDATE)/stackato-common
PKGTMPDIR=$(COMMON_DIR)/go

INSTALLHOME=/home/stackato
INSTALLROOT=$(INSTALLHOME)/stackato
GOBINDIR=$(INSTALLROOT)/go/bin

INSTDIR=$(DESTDIR)$(prefix)

INSTHOMEDIR=$(INSTDIR)$(INSTALLHOME)
INSTROOTDIR=$(INSTDIR)$(INSTALLROOT)
INSTGOPATH=$(INSTDIR)$(INSTALLROOT)/go
INSTBINDIR=$(INSTDIR)$(INSTALLHOME)/bin

BUILDGOPATH=$(shell pwd)/.gopath

GOARGS=-v -tags zmq_3_x

GOARGS_TEST=-race

export PATH := /usr/local/go/bin:$(BUILDGOPATH)/bin/:$(PATH)

all:	repos compile

repos:
	mkdir -p $(BUILDGOPATH)/src/$(NAME)
	git archive HEAD | tar -x -C $(BUILDGOPATH)/src/$(NAME)
	GOPATH=$(BUILDGOPATH) GOROOT=/usr/local/go go get -v github.com/vube/depman
	GOPATH=$(BUILDGOPATH) GOROOT=/usr/local/go depman
	rm -f $(BUILDGOPATH)/bin/depman

$(COMMON_DIR):	update

compile:	$(BUILDGOROOT)
	GOPATH=$(BUILDGOPATH) GOROOT=/usr/local/go go install $(GOARGS) $(NAME)/...
	GOPATH=$(BUILDGOPATH) GOROOT=/usr/local/go go install $(GOARGS) github.com/ActiveState/tail/cmd/gotail
	GOPATH=$(BUILDGOPATH) GOROOT=/usr/local/go go test $(GOARGS) $(GOARGS_TEST) logyard/... confdis/go/confdis/...

install:	
	mkdir -p $(INSTGOPATH)/$(SRCDIR)
	rsync -a $(BUILDGOPATH)/$(SRCDIR)/etc/*.yml $(INSTGOPATH)/$(SRCDIR)/etc/
	mkdir -p $(INSTGOPATH)/$(SRCDIR)/clients/apptail/
	rsync -a $(BUILDGOPATH)/$(SRCDIR)/clients/apptail/*.conf $(INSTGOPATH)/$(SRCDIR)/clients/apptail/
	rsync -a $(BUILDGOPATH)/bin $(INSTGOPATH)
	rsync -a etc $(INSTROOTDIR)
	mkdir -p $(INSTBINDIR)
	ln -sf $(GOBINDIR)/logyard-cli $(INSTBINDIR)
	chown -Rh stackato.stackato $(INSTHOMEDIR)

clean:	$(BUILDGOROOT)
	GOPATH=$(BUILDGOPATH) GOROOT=/usr/local/go go clean

# For developer use.

dev-setup:	update
	cd .stackato-pkg/update/stackato-common/go && ./goget

dev-install:	fmt dev-installall

# convenient alias
i:	dev-install

dev-installall:
	go install $(GOARGS) logyard/... github.com/ActiveState/tail/cmd/gotail

fmt:
	gofmt -w .

dev-test:
	go test $(GOARGS) $(GOARGS_TEST) logyard/... confdis/go/confdis/...
