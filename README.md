# getting started

```
# prepare your gopath
export GOPATH=$HOME/go

git clone <this-repo> $GOPATH/src/logyard
cd $GOPATH/src/logyard
make dev-setup 
make dev-install
make dev-test  # optional
```

# run

```
make dev-install
$GOPATH/bin/systail &
$GOPATH/bin/apptail &
$GOPATH/bin/cloudevents &
$GOPATH/bin/logyard
```
