#!/bin/bash

# Load environment variables from .env file
set -o allexport
source ../.env
set +o allexport

helm uninstall postgres
kubectl delete pvc data-greenlight-db-postgresql-0

helm install greenlight-db bitnami/postgresql -f postgres-values.yaml \
  --set global.postgresql.auth.username=$POSTGRES_USER \
  --set global.postgresql.auth.password=$POSTGRES_PASSWORD \
  --set global.postgresql.auth.database=$POSTGRES_DB

helm install greenlight-redis bitnami/redis -f redis-values.yaml