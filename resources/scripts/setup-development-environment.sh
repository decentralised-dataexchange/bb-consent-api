#!/bin/bash
set +e
set -u
set -x

cd "$(dirname "${BASH_SOURCE[@]}")"

CONTAINER_MONGO="mongo"
CONTAINER_KEYCLOAK="keycloak"

# Check for configuration file
CONFIG_PATH="${PWD}/../config"
CONFIG_FILE="${CONFIG_PATH}/config-development.json"

MONGODB_USER=$(jq -r .DataBase.username< "$CONFIG_FILE")
MONGODB_PASSWORD=$(jq -r .DataBase.password < "$CONFIG_FILE")
MONGODB_DATABASE=$(jq -r .DataBase.name < "$CONFIG_FILE")
KEYCLOAK_USER=$(jq -r .Iam.AdminUser< "$CONFIG_FILE")
KEYCLOAK_PASSWORD=$(jq -r .Iam.AdminPassword < "$CONFIG_FILE")

(docker ps -af "name=${CONTAINER_MONGO}" | grep "${CONTAINER_MONGO}" > /dev/null) && docker rm -f "${CONTAINER_MONGO}" > /dev/null

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
    --name "$CONTAINER_KEYCLOAK" \
    -e KEYCLOAK_USER="$KEYCLOAK_USER" \
    -e KEYCLOAK_PASSWORD="$KEYCLOAK_PASSWORD" \
    -e DB_VENDOR="h2" \
    -v keycloak-datadir:/opt/jboss/keycloak/standalone/data \
    -p 8080:8080 \
    jboss/keycloak:latest  # Use the appropriate Keycloak image and ports
