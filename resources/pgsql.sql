CREATE SCHEMA IF NOT EXISTS public;

CREATE EXTENSION fuzzystrmatch;
CREATE EXTENSION postgis;
CREATE EXTENSION postgis_topology;
CREATE EXTENSION postgis_tiger_geocoder;
CREATE EXTENSION pg_trgm;

ALTER SCHEMA topology OWNER TO rds_superuser;
ALTER SCHEMA tiger OWNER TO rds_superuser;

CREATE FUNCTION exec(TEXT)
  RETURNS TEXT LANGUAGE plpgsql VOLATILE AS $f$ BEGIN EXECUTE $1;
  RETURN $1;
END; $f$;

SELECT exec('ALTER TABLE ' || quote_ident(s.nspname) || '.' || quote_ident(s.relname) || ' OWNER TO rds_superuser')
FROM (
  SELECT
    nspname,
    relname
  FROM pg_class c JOIN pg_namespace n ON (c.relnamespace = n.oid)
  WHERE nspname IN ('tiger', 'topology') AND
        relkind IN ('r', 'S', 'v')
  ORDER BY relkind = 'S') AS "s";

CREATE OR REPLACE FUNCTION to_text(text)
  RETURNS text AS
$func$
SELECT to_char(to_timestamp($1, 'YYYY-MM-DD"T"HH24:MI:SS.US')  -- adapt to your pattern
            AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US')
$func$ LANGUAGE sql IMMUTABLE;

CREATE SCHEMA tg;

CREATE TABLE tg.accounts (
  id SERIAL PRIMARY KEY NOT NULL,
  json_data JSONB NOT NULL
);

CREATE TABLE tg.account_users (
  id SERIAL PRIMARY KEY NOT NULL,
  account_id INT NOT NULL,
  json_data JSONB NOT NULL
);

CREATE TABLE tg.account_user_sessions (
  account_id INT NOT NULL,
  account_user_id INT NOT NULL,
  session_id CHAR(40) NOT NULL,
  created_at TIMESTAMP DEFAULT now() NOT NULL
);

CREATE TABLE tg.applications (
  id SERIAL PRIMARY KEY NOT NULL,
  account_id INT NOT NULL,
  json_data JSONB NOT NULL,
  enabled INT DEFAULT 1 NOT NULL
);

CREATE TABLE tg.consumers
(
  consumer_name TEXT NOT NULL,
  consumer_position TEXT NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE INDEX on tg.accounts USING GIN (json_data);
CREATE INDEX on tg.account_users USING GIN (json_data);
CREATE INDEX on tg.applications USING GIN (json_data);

