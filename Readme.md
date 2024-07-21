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

### DB migrations
using migrate do it in CLI
golang\migrate to run db migrations at startup

### Rate limiting
Chosen algo : token bucket
with burst of b and on average r requests per s

### Dependencies
Go uses module proxies to ensure package longevity.
Package are located at  https://proxy.golang.org.
If needed(firewall issue, own module mirror), one can change env GOPROXY to add the needed proxy.
Go proxy does not guarantee that the sources will be present forever(unlikely, but possible)
An alternative is to use vendor to keep all the sources of in a vendor folder.
It can be useful for long standing applications.
Metrics are handled with prometheus and vizualisation is done with grafana

### How to run
The application is deployed using docker containers and docker compose.
Use .env.template to create your .env file with your credentials
then install docker on your machine.
use 

make docker/compose/up/rebuild to launch the app


### User Flow
1. Register by hitting /v1/users with {"name": "Bob Jones", "email": "bob@example.com", "password": "pa55word"}
2. Using the token recieved by email to confirm your registration v1/users/activated
3. Get your token for others API call by hitting /v1/tokens/authentication with {"email": "bob@example.com", "password": "pa55word"}

### Deploying with kubernetes
1. setup a local env with either kind or activate k8 in docker desktop
2. create .env with your environment variables
3. run the helm_install_postgres.sh to set postgres and redis
4. deploy using helm install 
