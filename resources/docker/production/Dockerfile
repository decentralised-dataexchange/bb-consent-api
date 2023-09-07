# https://github.com/GoogleContainerTools/distroless/tree/master/base
# Predecessor to scratch + ca-certificates
FROM frolvlad/alpine-glibc:glibc-2.30

# copy artifacts
COPY dist/linux_amd64/bb-consent-api /app/bin/bb-consent-api
COPY resources/config/config-development.json /app
COPY resources/config/auth_model.conf /app/config/ 
COPY resources/config/rbac_policy.csv /app/config/ 

# start API
EXPOSE 80
ENTRYPOINT ["./app/bin/bb-consent-api", "-config", "config-production.json"]
