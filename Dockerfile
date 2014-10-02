# Dockerfile for dev convenience only. Not used in production.

FROM golang

RUN apt-get -y update
RUN apt-get -y install build-essential pkg-config

# RUN apt-get -y install software-properties-common python-software-properties
# RUN add-apt-repository -y ppa:chris-lea/zeromq
# RUN apt-get -y update
# RUN apt-get -y install zeromq3

RUN go get github.com/vube/depman
RUN mkdir /s/ && ln -s /go /s/go

ADD http://download.zeromq.org/zeromq-3.2.4.tar.gz /
RUN cd / && tar xzf /zeromq-3.2.4.tar.gz
WORKDIR /zeromq-3.2.4
RUN ./configure && make && make install
RUN ldconfig

ADD . /go/src/logyard
WORKDIR /go/src/logyard

# Replace internal mirror URLs with the original github clone URI,
# and run depman. XXX: need to do this recursively on dependencies' deps.json
# Otherwise, need VPN access.
# RUN sed -i 's/git-mirrors.activestate.com\///' deps.json && depman
RUN depman

RUN go install -v logyard/...
