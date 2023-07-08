# Doing advanced REST API with Go.

I am working on API from scratch that is similar to Open Movie Database API

## Routing
| Method | URL Pattern      | Handler              | Action                              |
|--------|------------------|----------------------|-------------------------------------|
| GET    | /v1/healthcheck  | healthcheckHandler   | Show application information        |
| GET    | /v1/movies       | listMoviesHandler    | Show the details of all movies      |
| POST   | /v1/movies       | createMovieHandler   | Create a new movie                  |
| GET    | /v1/movies/:id   | showMovieHandler     | Show the details of a specific movie|
| PUT    | /v1/movies/:id   | editMovieHandler     | Update the details of a specific movie |
| DELETE | /v1/movies/:id   | deleteMovieHandler   | Delete a specific movie             |


### Encoding responses as JSON
- use json.Marshal as it allows set headers
- just a bit less performant that json.NewEncoder

MarshalIndent to improve readability, but it is more expensie that Marshal.
Keep it mind if needed.