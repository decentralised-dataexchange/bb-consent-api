# https://github.com/GoogleContainerTools/distroless/tree/master/base
# Predecessor to scratch + ca-certificates
FROM frolvlad/alpine-glibc:glibc-2.30

# copy artifacts
COPY dist/linux_amd64/bb-consent-api /app/bin/bb-consent-api
COPY resources/config/config-development.json /app
COPY resources/assets/logo.jpeg /opt/bb-consent/api/assets/
COPY resources/assets/cover.jpeg /opt/bb-consent/api/assets/
COPY resources/assets/avatar.jpeg /opt/bb-consent/api/assets/

# start API
EXPOSE 80
ENTRYPOINT ["./app/bin/bb-consent-api", "start-api", "--config", "config-production.json"]
