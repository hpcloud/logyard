# getting started

```
# prepare your gopath
export GOPATH=$HOME/go
git clone <this-repo> $GOPATH/src/logyard
cd $GOPATH/src/logyard
git clone -q gitolite@gitolite.activestate.com:tail $GOPATH/src/github.com/srid/tail  # XXX: until tail.git moves to github
make setup 
make
make test  # optional
```

# run

```
make
$GOPATH/bin/systail &
$GOPATH/bin/apptail &
$GOPATH/bin/cloudevents &
$GOPATH/bin/logyard
```
