# getting started

```
# prepare your gopath
export GOPATH=$HOME/go

git clone <this-repo> $GOPATH/src/logyard
cd $GOPATH/src/logyard
# TODO: merge Makefile.dev to Makefile
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
