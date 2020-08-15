# Crius
TODO description

## Contributing

### Dependencies

* [go v1.14](https://golang.org/dl/)
* [Docker desktop](https://docs.docker.com/desktop/)
* [migrate v4](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

### Dev Workflow

```bash
# View available make targets
make help

# Spin up a DB, run the app against it
make db/run migrate/up service/run

# Tidy up code and run the unit tests
make tidy service/test
```
