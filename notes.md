# Distributed Services in Go

## Build a commit log

### HTTP Server using Json
an API with two endpoints, 
one to write and another to consume.

### HTTP Server with gRPC

## Building a Log

- Record: the data stored in the log
- Store : the file we store records in
- Index : the fie we store index entries in
- Segment : the abstraction that ties a store and index together
- Log : the abstraction that ties all the segments together

Memory-mapped file can improve read/write times


## Write a log package
Create a log library

## Network
Create a service that people can access via network. People do not pay for library but for services
Secure the services
- TLS to authenticate the server
- TLS mutual authentication, as 2-way auth is mostly common in machine-to-machine comm, i.e distributed systems

- For Authorization, use ACL

- use CFSSL to crete my own CA
- casbin for ACL

## Add observability
Metrics, logs and traces

# Distrubte

## Service discovery
using serf or consul or etcd or zookeeper
serf is simple and decentralized.
All instances of the server runs a serf node
## Data Replication
pull based vs push-based (raft, leader-follwer replication)
using grpc client, data is replicated
### Delivery
either run package or agent package

## Deploy

### Use Kubernetes with Kind for local deployment
for our service, we will use stateful sets to manage the application as the service persists in a log
* Stable, unique network identifiers—each node in our service requires
unique node names as identifiers.
* Stable, persistent storage—our service expects the data it's written to
persist across restarts.
* Ordered, graceful deployment and scaling—our service needs initial
node to bootstrap the cluster and join subsequent nodes to its cluster.
* Ordered, automated rolling updates—we always want our cluster to
have a leader, and when we roll the leader we want to give the cluster
enough time to elect a new leader before rolling the next node.
And by “stable,” I mean persisted across scheduling changes like restarts
and scaling

for stateless applications, use deployment.
For example, for a service that persists in a relational database, like Postgres, you'd run the API service with deployment and Postgres with a stateful set

# TODO
Graceful close raft log store when rafting shutdown
