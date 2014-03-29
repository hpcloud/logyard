PWD := $(shell pwd)
GOARGS=-v -tags zmq_3_x
# Because 'go ./...' doesn't like symlinks
TARGETS=$(shell find . -type f -name \*.go -not -path "./vendor/*"  | xargs -n 1 dirname | sort | uniq | sed -e 's/^\./logyard/')

# To be invoked from docker container to produce binary at ./vendor/bin/
# FIXME: tests
all:
	ln -s ${PWD} ${GOPATH}/src/logyard
	GOPATH=${PWD}/vendor depman
	echo ${GODIRS}
	cd ${GOPATH}/src/logyard && GOBIN=${PWD}/vendor/bin GOPATH=${GOPATH}:${PWD}/vendor \
		sh -ex -c "go install ${GOARGS} ${TARGETS}; go test ${GOARGS} ${TARGETS}"
	ls -l vendor/bin

docker:
	docker run -v `pwd`:/source:rw stackatobuild/go make
	docker build -t stackato/logyard .

fmt:
	gofmt -w .
