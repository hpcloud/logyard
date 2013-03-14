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

NAME=logyard

SRCDIR=src/$(NAME)

COMMON_REPO=git://gitolite.activestate.com/stackato-common.git

UPDATE=.stackato-pkg/update
COMMON_DIR=$(UPDATE)/stackato-common
TMPDIR=$(COMMON_DIR)/go

INSTALLHOME=/home/stackato
INSTALLROOT=$(INSTALLHOME)/stackato
GOBINDIR=$(INSTALLROOT)/go/bin

INSTDIR=$(DESTDIR)$(prefix)

INSTGOPATH=$(INSTDIR)/$(INSTALLROOT)/go
INSTBINDIR=$(INSTDIR)/$(INSTALLHOME)/bin

GOPATH=$$PWD/.gopath

ifdef STACKATO_PKG_BRANCH
    BRANCH_OPT=-b $(STACKATO_PKG_BRANCH)
endif

all:	repos compile

repos:	$(COMMON_DIR)
	mkdir -p $(GOPATH)/src/$(NAME)
	git archive HEAD | tar -x -C $(GOPATH)/src/$(NAME)
	GOPATH=$(GOPATH) $(TMPDIR)/goget $(TMPDIR)/goget.manifest

$(COMMON_DIR):	update

compile:	
	GOPATH=$(GOPATH) go install -tags zmq_3_x -v $(NAME)/...
	GOPATH=$(GOPATH) go install -v github.com/ActiveState/tail/cmd/gotail

install:	
	mkdir -p $(INSTGOPATH)/$(SRCDIR)
	rsync -a $(GOPATH)/$(SRCDIR)/config.yml $(INSTGOPATH)/$(SRCDIR)
	rsync -a $(GOPATH)/bin $(INSTGOPATH)
	mkdir -p $(INSTBINDIR)
	ln -sf $(GOBINDIR)/logyardctl $(INSTBINDIR)

clean: 
	GOPATH=$(GOPATH) go clean

# For manual use.

update:
	rm -rf $(UPDATE)
	git clone $(BRANCH_OPT) $(COMMON_REPO) $(COMMON_DIR)

