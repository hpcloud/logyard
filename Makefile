#
# Makefile for stackato-logyard
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
#   rm -rf .gopath .goroot; STACKATO_PKG_BRANCH=mybranch make
#

NAME=logyard

SRCDIR=src/$(NAME)

COMMON_REPO=git://gitolite.activestate.com/stackato-common.git

UPDATE=.stackato-pkg/update
COMMON_DIR=$(UPDATE)/stackato-common
PKGTMPDIR=$(COMMON_DIR)/go
BUILDGOROOT=$$PWD/.goroot

INSTALLHOME=/home/stackato
INSTALLROOT=$(INSTALLHOME)/stackato
GOBINDIR=$(INSTALLROOT)/go/bin

INSTDIR=$(DESTDIR)$(prefix)

INSTHOMEDIR=$(INSTDIR)$(INSTALLHOME)
INSTROOTDIR=$(INSTDIR)$(INSTALLROOT)
INSTGOPATH=$(INSTDIR)$(INSTALLROOT)/go
INSTBINDIR=$(INSTDIR)$(INSTALLHOME)/bin

BUILDGOPATH=$$PWD/.gopath

ifdef STACKATO_PKG_BRANCH
    BRANCH_OPT=-b $(STACKATO_PKG_BRANCH)
endif

all:	installgo repos compile

installgo:	$(BUILDGOROOT)

# Manually download and install the Go binary until Ubuntu updates its
# Go version.
$(BUILDGOROOT):
	wget -c https://go.googlecode.com/files/go1.1.linux-amd64.tar.gz
	tar zxf *.tar.gz
	mv go .goroot
	rm -f go*tar.gz

repos:	$(COMMON_DIR)
	mkdir -p $(BUILDGOPATH)/src/$(NAME)
	git archive HEAD | tar -x -C $(BUILDGOPATH)/src/$(NAME)
	GOPATH=$(BUILDGOPATH) $(PKGTMPDIR)/goget $(PKGTMPDIR)/goget.manifest

$(COMMON_DIR):	update

compile:	$(BUILDGOROOT)
	GOPATH=$(BUILDGOPATH) GOROOT=$(BUILDGOROOT) $(BUILDGOROOT)/bin/go install -tags zmq_3_x -v $(NAME)/...
	GOPATH=$(BUILDGOPATH) GOROOT=$(BUILDGOROOT) $(BUILDGOROOT)/bin/go install -v github.com/ActiveState/tail/cmd/gotail
	GOPATH=$(BUILDGOPATH) GOROOT=$(BUILDGOROOT) $(BUILDGOROOT)/bin/go test -v logyard/... confdis/go/confdis/...

install:	
	mkdir -p $(INSTGOPATH)/$(SRCDIR)
	rsync -a $(BUILDGOPATH)/$(SRCDIR)/etc/*.yml $(INSTGOPATH)/$(SRCDIR)/etc/
	rsync -a $(BUILDGOPATH)/bin $(INSTGOPATH)
	rsync -a etc $(INSTROOTDIR)
	mkdir -p $(INSTBINDIR)
	ln -sf $(GOBINDIR)/logyard-cli $(INSTBINDIR)
	chown -Rh stackato.stackato $(INSTHOMEDIR)

clean:	$(BUILDGOROOT)
	GOPATH=$(BUILDGOPATH) $(BUILDGOROOT)/bin/go clean

# For manual use.

update:
	rm -rf $(UPDATE)
	git clone $(BRANCH_OPT) $(COMMON_REPO) $(COMMON_DIR)

# For developer use.

dev-setup:	update
	cd .stackato-pkg/update/stackato-common/go && ./goget

dev-install:	fmt dev-installall

dev-installall:
	go install -tags zmq_3_x -v logyard/... github.com/ActiveState/tail/cmd/gotail

fmt:
	rm -rf .goroot
	gofmt -w .

dev-test:
	go test logyard/... confdis/go/confdis/...
