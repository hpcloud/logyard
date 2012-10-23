# getting started

```
# prepare your gopath
export GOPATH=$HOME/go
git clone <this-repo> $GOPATH/src/logyard
cd $GOPATH/src/logyard
make setup  # XXX: clone tail.git separately
make test  # optional
```

# run

```
make
$GOPATH/bin/systail &
$GOPATH/bin/apptail &
$GOPATH/bin/logyard
```
