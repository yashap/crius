# Crius
Crius is a work-in-progress, open source project, that helps you manage dependencies between frontends, services and events, in a service oriented architecture. It will help you visualize your system, and find both direct and transitive dependencies of any service, or any service endpoint.

## Contributing

### Dependencies

* [go v1.14](https://golang.org/dl/)
* [Docker desktop](https://docs.docker.com/desktop/)

### Dev Workflow

```bash
# View available make targets
make help

# Run the DB and HTTP server (will wipe local DB)
make run

# Just run the HTTP server (assumes DB is already running)
make run-server

# Tidy up code and run all unit and integration tests
make tidy test
```
