#!/bin/bash
# install zeromq-2.2 into /usr/local

set -xe

sudo apt-get -yq install uuid-dev

rm -rf /tmp/zeromq-2.2.0*
pushd /tmp
wget http://download.zeromq.org/zeromq-2.2.0.tar.gz
tar zxf zeromq-2.2.0.tar.gz
pushd zeromq-2.2.0
./configure && make
sudo make install

popd; popd

