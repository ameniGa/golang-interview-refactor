## Refactor Notes
* Fix tightly coupled components and mixed responsibilities:
    * Add interface for cart service.
    * Add repository to isolate database operations.
    * Make the controller depends on a cart service handler.
    * Inject interfaces as dependencies to make the code testable.
* Refactor calculator methods to make them more readable and concise.
* Move HTTP request check and forms to controller instead of calculator service.
* Handle HTTP responses in controller and keep the service Gin agnostic.
* Create one instance of DB connection and pass it for all components.
* Added unit test for cart service and other public functions.
* Added integration test for the repository.
* Added configuration management.
* Delete credentials from code base and docker files.
* Make renderTemplate function as a helper so can be used with different controllers.
* Added Makefile for common commands.

## Run unit tests
```
make test
```

## Run integration test
```
make integration_test
```

## Run application
If you don't have mysql on your system run docker first:
```
make run-docker
```
then start the app
```
make start 
```
NB: Make sure the cmd/web-api/config/.env and docker/.env have same variables

NB: env files were pushed for the sake of the assignment but never forget to include it in the .gitignore