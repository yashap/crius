# Crius
Crius is a work-in-progress, open source project, that helps you manage dependencies between frontends, services and events, in a service oriented architecture. It will help you visualize your system, and find both direct and transitive dependencies of any service, or any service endpoint.

## Contributing

### Dependencies

* [go v1.14](https://golang.org/dl/)
* [Docker desktop](https://docs.docker.com/desktop/)
* [Node Version Manager](https://github.com/nvm-sh/nvm)

### Dev Workflow

```bash
# View available make targets
make help

# Run the DB and HTTP server (will wipe local DB)
make service/run

# Just run the HTTP server (assumes DB is already running, won't wipe the DB)
make service/postgres/run

# Run the UI. Will open in a browser, and you must have the backend running already (i.e. `make run`)
nvm use
make ui/run

# Tidy up code and run all unit and integration tests
make tidy test
```

### Modifying the DB Schema

TODO: explain the process

You'll need to have the following additional dependencies installed:

* [SQLBoiler](https://github.com/volatiletech/sqlboiler)
    * `GO111MODULE=off go get -u -t github.com/volatiletech/sqlboiler`
    * `GO111MODULE=off go get github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql`
    * `GO111MODULE=off go get github.com/volatiletech/sqlboiler/drivers/sqlboiler-mysql`
* TODO link to migration tool

## Special Thanks

Special thanks to [Jet Brains](https://www.jetbrains.com/?from=crius) for contributing a free IDE licence to this project via their [open source licence program](https://www.jetbrains.com/community/opensource/#support?from=crius).
