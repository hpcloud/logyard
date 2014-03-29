FROM stackato/base

# Runtime dependency on libzmq.so
RUN wget -q http://stackato-pkg.nas1.activestate.com/repo-common/zeromq-dev_3.2.2_amd64.deb && \
    dpkg -i zeromq-dev*deb && \
    rm -f zeromq-dev*deb

ADD vendor/bin /logyard/bin
ADD etc /logyard/etc
ADD stackon.json /

ENV PATH /logyard/bin:$PATH
