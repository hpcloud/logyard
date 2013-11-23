[Logyard: The Why and How of Stackato's Logging System](http://www.activestate.com/blog/2013/04/logyard-why-and-how-stackatos-logging-system)

# getting started

```
# prepare your gopath
export GOPATH=$HOME/go

git clone <this-repo> $GOPATH/src/logyard
cd $GOPATH/src/logyard
make dev-setup 
make dev-install  # or, `make i`
make dev-test  # optional
```

# run

```
make dev-install
$GOPATH/bin/logyard
# $GOPATH/bin/systail &
# $GOPATH/bin/apptail &
# $GOPATH/bin/logyard_sieve &
```

note: it is best to run these on the Stackato dev VM.

