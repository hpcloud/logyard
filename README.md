[Logyard: The Why and How of Stackato's Logging System](http://www.activestate.com/blog/2013/04/logyard-why-and-how-stackatos-logging-system)

# getting started

clone sources on a Stackato VM and then run once:

```
stackon-fetch http://stackato:suchDogeW0w@docker-internal.stackato.com master stackatobuild/go
```

hack, hack, hack ... and build:

```
make docker
```

restart logyard (or whatever) with the new docker image:

```
sup restart logyard  # or whatever
```
