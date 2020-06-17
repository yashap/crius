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
make help # View available make targets
make db/run service/run
```

## Creating Database Migration Scripts

```bash
migrate create -ext sql -dir scripts/db/migrations -seq $file_name
```

TODO:
Doing this tutorial:
  https://semaphoreci.com/community/tutorials/building-and-testing-a-rest-api-in-go-with-gorilla-mux-and-postgresql
Left off at:
  The test to check the response when fetching a nonexistent product can be implemented as follows
