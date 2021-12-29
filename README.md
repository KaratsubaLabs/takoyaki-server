
# takoyaki backend

go backend for からつばLABS' **project takoyaki** - the vps platform

## RUNNING FOR DEVELOPMENT

If you wish to run **takoyaki** without having to keep building docker
containers, you can run it locally instead. Make sure you have a working go
installation.

Make your own copy of `.env` by copying the provided `dotenv.example`
file.
```
$ cp dotenv.example .env
```

Install packages
```
$ go mod download
```

Export environment variables from `.env` file
```
$ export $(grep -v '^#' .env | xargs)
```

Run the database container
```
$ docker-compose up -d db
```

Initialize/migrate the database
```
$ go run *.go db migrate
```

Run takoyaki
```
$ go run *.go server
```

There are also a couple of utility scripts for rapid development in `dev/`. Be
very careful when running these as they will purge all of your containers and
vms.

## RUNNING FOR PRODUCTION

To be able to run the stack, **docker** and **docker-compose** are required.
Consult the relevant documentation based on your system on how to get these set
up.

First, make your own copy of `.env` by copying the provided `dotenv.example`
file.
```
$ cp dotenv.example .env
```

**takoyaki-pipe**, the systemd service that executes libvirt commands for the
**takoyaki** container must also be installed and ran:
```
$ cd takoyaki-pipe
$ make install
$ systemctl enable takoyaki-pipe
$ systemctl start takoyaki-pipe
```

Next we can start the containers
```
$ docker-compose up --build
```

## RESETING THE DATABASE

During testing, if it happens that you wish to reset the database, simply
remove the directory:
```
$ sudo rm -rf db/data/
```

## TODO

- [x] database initialization migration (+ shell interface to init db)
- [x] validation for requests (as a middleware if possible)
- [x] possibly error middleware
- [x] figure out where to put temp files (cidata.iso etc) for when creating vps
- [x] jwt auth
- [x] look at database transactions (+ are they really needed)
- [x] write the routes
- [x] allow optional args when configuring vps (ie ssh key)
- [ ] add tests? (might be overkill + annoying)
- [ ] get progress of creating vps to show to frontend (possibly)
- [ ] snapshot requests
- [ ] look into using RLS
- [ ] return ip address and state of vm as well in vps info endpoint + look into vm networking
- [x] execute vm commands on host
- [ ] refactor project into multiple modules
- [ ] rewrite api

