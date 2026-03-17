#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE USER lynx WITH ENCRYPTED PASSWORD 'lynx';
	CREATE DATABASE lynx OWNER lynx;
EOSQL
