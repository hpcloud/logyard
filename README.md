# getting started

```
# prepare your gopath
export GOPATH=$HOME/go

# retrieve goget
sudo wget "http://gitolite.activestate.com/?p=goget.git;a=blob_plain;f=goget;hb=refs/heads/master" -O /usr/local/bin/goget 
sudo chmod +x /usr/local/bin/goget

git clone <this-repo> $GOPATH/src/logyard
cd $GOPATH/src/logyard
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
