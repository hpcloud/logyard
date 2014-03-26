[Logyard: The Why and How of Stackato's Logging System](http://www.activestate.com/blog/2013/04/logyard-why-and-how-stackatos-logging-system)

# getting started

get a Stackato VM running and then:


```
# install golang
wget http://stackato-pkg.nas1.activestate.com/repo-common/stackato-golang_1.2_amd64.deb
sudo dpkg -i stackato-golang*deb
export PATH=/usr/local/go/bin:$PATH

# prepare your gopath
export GOPATH=/s/go
rm -rf /s/go/src 

git clone <this-repo> $GOPATH/src/logyard
cd $GOPATH/src/logyard

# install dependencies
sudo apt-get install zeromq-dev>=3.2.2
go get -v github.com/vube/depman
depman

# build logyard binary
make i

make dev-test  # optional
```

# run

```
make i
# restart all of selected services
sup restart logyard systail apptail logyard_sieve docker_events
```

