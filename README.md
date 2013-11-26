[Logyard: The Why and How of Stackato's Logging System](http://www.activestate.com/blog/2013/04/logyard-why-and-how-stackatos-logging-system)

# getting started

```
# prepare your gopath
export GOPATH=$HOME/go

# install depman
go get -v github.com/vube/depman

git clone <this-repo> $GOPATH/src/logyard
cd $GOPATH/src/logyard

# install dependencies
depman

# build logyard binary
make i

make dev-test  # optional
```

# run

rsync your GOPATH to the VM's /s/go and:

```
make i
# restart all of selected services
sup restart logyard systail apptail logyard_sieve docker_events
```

