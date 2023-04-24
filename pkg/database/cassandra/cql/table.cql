CREATE KEYSPACE IF NOT EXISTS event 
WITH replication = {'class': 'NetworkTopologyStrategy', 'dc1': '3'} 
AND durable_writes = true;

CREATE TABLE IF NOT EXISTS event.event (
    id                  UUID,
    user_id             varchar,
    identifier          varchar,
    event_timestamp     bigint,
    type                int,
    PRIMARY KEY ((id), user_id, identifier, event_timestamp, type)
);

CREATE TABLE IF NOT EXISTS event.event_data(
    id                  UUID,
    data                TEXT,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS event.event_date(
    event_date          text,
    event_timestamp     bigint,
    id                  UUID,
    PRIMARY KEY ((event_date), event_timestamp, id)
) WITH CLUSTERING ORDER BY (event_timestamp DESC);

CREATE TABLE IF NOT EXISTS event.event_user_id (
    user_id             varchar,
    identifier          varchar,
    id                  UUID,
    PRIMARY KEY ((user_id), identifier, id)
);

