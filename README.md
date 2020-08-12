# Crius
TODO description

## Contributing

### Dependencies

* [go v1.14](https://golang.org/dl/)
* [Docker desktop](https://docs.docker.com/desktop/)
* [migrate v4](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
* If you want to run with "rebuild on save", [gin](https://github.com/codegangsta/gin)

### Dev Workflow

```bash
# View available make targets
make help

# Spin up a DB, run the app against it
make db/run migrate/up service/run

# Format code and run the unit tests
make fmt service/test
```

## Creating Database Migration Scripts

```bash
migrate create -ext sql -dir scripts/db/migrations -seq $file_name
```
