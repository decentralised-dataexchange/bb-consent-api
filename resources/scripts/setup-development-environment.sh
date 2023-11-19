#!/bin/bash
set +e
set -u
set -x

cd "$(dirname "${BASH_SOURCE[@]}")"

CONTAINER_MONGO="mongo"
CONTAINER_KEYCLOAK="keycloak"
CONTAINER_POSTGRESQL="postgresql"

# Check for configuration file
CONFIG_PATH="${PWD}/../config"
CONFIG_FILE="${CONFIG_PATH}/config-development.json"

MONGODB_USER=$(jq -r .DataBase.username< "$CONFIG_FILE")
MONGODB_PASSWORD=$(jq -r .DataBase.password < "$CONFIG_FILE")
MONGODB_DATABASE=$(jq -r .DataBase.name < "$CONFIG_FILE")
KEYCLOAK_USER=$(jq -r .Iam.AdminUser< "$CONFIG_FILE")
KEYCLOAK_PASSWORD=$(jq -r .Iam.AdminPassword < "$CONFIG_FILE")

(docker ps -af "name=${CONTAINER_MONGO}" | grep "${CONTAINER_MONGO}" > /dev/null) && docker rm -f "${CONTAINER_MONGO}" > /dev/null
(docker ps -af "name=${CONTAINER_KEYCLOAK}" | grep "${CONTAINER_KEYCLOAK}" > /dev/null) && docker rm -f "${CONTAINER_KEYCLOAK}" > /dev/null
(docker ps -af "name=${CONTAINER_POSTGRESQL}" | grep "${CONTAINER_POSTGRESQL}" > /dev/null) && docker rm -f "${CONTAINER_POSTGRESQL}" > /dev/null

ARCH=$(uname -m)

if [ "$ARCH" == "arm64" ] ; then
    docker build --platform=linux/amd64 -t bb-consent/mongo:4.0.4 ../docker/development/
else
    docker build -t bb-consent/mongo:4.0.4 ../docker/development/
fi

docker run -d \
    --name "$CONTAINER_MONGO" \
    -e MONGODB_APPLICATION_DATABASE="$MONGODB_DATABASE" \
    -e MONGODB_APPLICATION_USER="$MONGODB_USER" \
    -e MONGODB_APPLICATION_PASSWORD="$MONGODB_PASSWORD" \
	-v mongo-datadir:/data/db \
    -p 27017:27017 \
    bb-consent/mongo:4.0.4

# Wait for Mongo DB
#echo "Wait for Mongo..."
#while ! docker exec mongo mysqladmin ping -h"127.0.0.1" -u"$MYSQL_USER" -p"$MYSQL_PASS" --silent &> /dev/null; do
#    sleep 1
#done

docker run -d \
    --name "$CONTAINER_POSTGRESQL" \
    -e POSTGRESQL_USERNAME="bn_keycloak" \
    -e POSTGRESQL_PASSWORD="bn_keycloak" \
    -e POSTGRESQL_DATABASE="bitnami_keycloak" \
    -v postgresql-datadir:/bitnami/postgresql \
    -p 5432:5432 \
    bitnami/postgresql:14.10.0  # Use the appropriate PostgreSQL image and ports

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to start..."
while ! docker logs "$CONTAINER_POSTGRESQL" 2>&1 | grep -q "database system is ready to accept connections"; do
    sleep 1
done

docker run -d \
    --name "$CONTAINER_KEYCLOAK" \
    -e KEYCLOAK_ADMIN_USER="$KEYCLOAK_USER" \
    -e KEYCLOAK_ADMIN_PASSWORD="$KEYCLOAK_PASSWORD" \
    -e KEYCLOAK_DATABASE_PASSWORD="bn_keycloak" \
    --link=${CONTAINER_POSTGRESQL} \
    -p 9090:8080 \
    bitnami/keycloak:22.0.2  # Use the appropriate Keycloak image and ports
