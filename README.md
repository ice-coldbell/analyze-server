# Analyze Project
This is a project that receives events through a specific protocol, and asynchronously processes them by a worker. 

In the server, the events are received and published to a message queue, while in the worker, the events are retrieved from the message queue and stored in a database. 

In the future, the worker or cron job may generate various metrics such as DAU, MAU, retention, and event user groups, or move unused events to storage.

```
The purpose of this repository is for the owner to create a simple project while also learning the basics of Golang project structure, Kafka, RabbitMQ, and Cassandra.
```
## Project Tree
```
.
├── application         : Entrypoint
│   ├── analyze-server  : Server entrypoint
│   └── analyze-worker  : Worker entrypoint
├── config              : Config file for Local environment
├── docker              : docker setting file
│   ├── build           : Dockerfile for docker image build
│   ├── config          : Config file for Docker environment
│   └── environment     : Docker environment file
├── core                : application core package
│   ├── config          : Support to load the configuration file
│   ├── infra           : Infrastructure interface used in the service
│   ├── model           : Defines the model used in the service
│   └── service         : Implements the core application logic
├── pkg
│   ├── errorx          : Custom errors that wrap the standard `errors` package
│   ├── logger          : Wrap the Uber Zap package logger
│   ├── database        : Implementation of a Database interface
│   │    └── cassandra
└── └── queue           : Implementation of a Queue interface
        ├── kafka
        └── rabbitmq
```

## How to run
`To run this project, you can follow the steps below`

```bash
$docker compose up -d rabbitmq
or
$docker compose up -d zookeeper kafka kafka-ui
```

```bash
# You need to verify that the Cassandra node is running properly before starting the next node.
$docker compose up -d cassandra-node-0
$docker compose up -d cassandra-node-1
$docker compose up -d cassandra-node-2
```

```bash
# You need to create a table and keyspace that will be used in the application.
$docker compose exec cqlsh < pkg/database/cassandra/cql/table.cql
```

```bash
# To build and run the server, use the following command
$docker compose up -d --force-recreate --build analyze-server analyze-worker
```