CREATE EXTENSION IF NOT EXISTS PostGIS;

DROP TABLE public.taxi_routes;
DROP TABLE public.taxi_route_steps;

CREATE TABLE public.taxi_routes
(
  id                    BIGSERIAL NOT NULL,
  taxi_id               INTEGER NOT NULL,
  pickup_time           TIMESTAMP WITHOUT TIME ZONE,
  dropoff_time          TIMESTAMP WITHOUT TIME ZONE,
  passenger_count       INTEGER,
  trip_distance         DOUBLE PRECISION,
  trip_duration         DOUBLE PRECISION,
  fare_amount           DOUBLE PRECISION,
  extra                 DOUBLE PRECISION,
  mta_tax               DOUBLE PRECISION,
  tip_amount            DOUBLE PRECISION,
  tolls_amount          DOUBLE PRECISION,
  ehail_fee             DOUBLE PRECISION,
  improvement_surcharge DOUBLE PRECISION,
  total_amount          DOUBLE PRECISION,
  payment_type          INTEGER,
  trip_type             INTEGER,
  geometry              GEOMETRY,
  CONSTRAINT taxi_routes_pkey PRIMARY KEY (id)
)
WITH (
OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.taxi_routes
  OWNER TO dobucher;

CREATE TABLE public.taxi_route_steps
(
  id         BIGINT NOT NULL,
  taxi_route BIGINT NOT NULL,
  distance   DOUBLE PRECISION,
  duration   DOUBLE PRECISION,
  mode       CHARACTER VARYING,
  name       CHARACTER VARYING,
  geometry   GEOMETRY,
  CONSTRAINT taxi_route_steps_pkey PRIMARY KEY (id)
)
WITH (
OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.taxi_route_steps
  OWNER TO dobucher;