#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE hosting_service_db;
    CREATE DATABASE resources_service_db;
    CREATE DATABASE kratos_db;
EOSQL