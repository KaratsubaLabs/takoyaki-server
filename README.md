
# takoyaki backend

go backend for からつばLABS's **project takoyaki** - the vps platform.

## RUNNING ALL CONTAINERS

To be able to run the stack, **docker** and **docker-compose** are required.
Consult the relevant documentation based on your system on how to get these set
up.

First, make your own copy of `.env` by copying the provided `dotenv.example`
file.
```
$ cp dotenv.example .env
```

Next we can start the containers
```
$ docker-compose up
```

## RESETING THE DATABASE

During testing, if it happens that you wish to reset the database, simply
remove the directory:
```
$ sudo rm -rf db/data/
```

## TODO

- [ ] database initialization migration (+ shell interface to init db)
- [ ] validation for requests (as a middleware if possible)
- [ ] possibly error middleware
- [ ] figure out where to put temp files (cidata.iso etc) for when creating vps
- [ ] jwt auth

