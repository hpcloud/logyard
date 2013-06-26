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
$GOPATH/bin/logyard
# $GOPATH/bin/systail &
# $GOPATH/bin/apptail &
# $GOPATH/bin/cloudevents &
```

note: it is best to run these on the VM in a proper stackato
requirement.

