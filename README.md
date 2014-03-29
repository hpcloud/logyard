[Logyard: The Why and How of Stackato's Logging System](http://www.activestate.com/blog/2013/04/logyard-why-and-how-stackatos-logging-system)

# getting started

run once:

```
docker login -u stackato -e s@s.com -p suchDogeW0w docker-internal.stackato.com
docker pull docker-internal.stackato.com/stackatobuild/go:master
docker tag docker-internal.stackato.com/stackatobuild/go:master stackatobuild/go
```

hack, hack, hack ... and build:

```
make docker
```

restart logyard (or whatever) with the new docker image:

```
sup restart logyard  # or whatever
```
