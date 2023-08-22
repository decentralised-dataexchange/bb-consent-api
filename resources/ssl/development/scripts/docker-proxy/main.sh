#!/bin/bash
# set -x
set -e
set -u

## Settings
readonly RESOLVER_TLD="dev"
readonly DOCKER_VM_NAME="docker-vm"
readonly CONTAINER_NGINX="proxy_nginx"
readonly CONTAINER_DNSMASQ="proxy_dns"

readonly ROOT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

## Flags
readonly FLAG_OSX=1
readonly FLAG_LINUX=2
readonly FLAG_DOCKER_MACHINE=4
readonly FLAG_DOCKER_BETA=8
readonly FLAG_DOCKER_UNKNOWN=16
readonly FLAG_PROXY_NGINX=32
readonly FLAG_PROXY_DNS=64

readonly BB_CONSENT_DOMAIN="bb-consent.dev"
readonly BB_CONSENT_SUBDOMAINS="api.${BB_CONSENT_DOMAIN} dashboard.${BB_CONSENT_DOMAIN} privacy.${BB_CONSENT_DOMAIN} docs.${BB_CONSENT_DOMAIN}"
ENVIRONMENT_FLAGS=0
DOCKER_IP=127.0.0.1


function detect_flags {
  which docker > /dev/null || (echo "docker command not found"; exit 1)

  ENVIRONMENT_FLAGS=$FLAG_PROXY_NGINX
  ## OS
  case $(uname -s) in
    "Darwin")
      ENVIRONMENT_FLAGS=$ENVIRONMENT_FLAGS+$FLAG_OSX
      ENVIRONMENT_FLAGS=$ENVIRONMENT_FLAGS+$FLAG_PROXY_DNS
    ;;
    "Linux")
      ENVIRONMENT_FLAGS=$ENVIRONMENT_FLAGS+$FLAG_LINUX
    ;;
    *)
      echo "Unsupported OS"
      exit 1
    ;;
  esac

  if (((ENVIRONMENT_FLAGS & FLAG_OSX)>0)); then
      case $(docker info 2>/dev/null | sed -n "s/Operating System: \(.*\)/\1/p") in
        "Alpine Linux v"*)
          # Docker beta
          ENVIRONMENT_FLAGS=$ENVIRONMENT_FLAGS+$FLAG_DOCKER_BETA
          ;;
        "Boot2Docker "*)
          # Docker machine
          ENVIRONMENT_FLAGS=$ENVIRONMENT_FLAGS+$FLAG_DOCKER_MACHINE
          ;;
        *)
          ENVIRONMENT_FLAGS=$ENVIRONMENT_FLAGS+$FLAG_DOCKER_UNKNOWN
          ;;
      esac
  fi
}

function create_proxy_dns {
  local host_ip=$1

  if [[ -z "${host_ip}" ]]; then
    echo "misssing host ip"
    exit 1
  fi

  (docker ps -af "name=${CONTAINER_DNSMASQ}" | grep "${CONTAINER_DNSMASQ}"> /dev/null) && docker rm -f "${CONTAINER_DNSMASQ}" > /dev/null
  docker run \
    -d \
    -p "${host_ip}:53:53/tcp" \
    -p "${host_ip}:53:53/udp" \
    --cap-add NET_ADMIN \
    --restart "always" \
    --name "${CONTAINER_DNSMASQ}" \
    andyshinn/dnsmasq \
    "--address=/${RESOLVER_TLD}/${host_ip}" \
    > /dev/null

  echo "Create resolver directory"
  sudo mkdir -p /etc/resolver && sudo chmod 755 /etc/resolver

  echo "Create ${RESOLVER_TLD} resolver file at /etc/resolver/${RESOLVER_TLD}"
  echo "nameserver ${host_ip}" | sudo tee "/etc/resolver/${RESOLVER_TLD}" > /dev/null

  set +e
  sleep 1
  ping -c1 "undefined.${RESOLVER_TLD}" > /dev/null 2>&1
  pingExitCode=$?
  set -e
  if [[ $pingExitCode -gt 0 ]]; then
    echo "Problem resolving .${RESOLVER_TLD} TLD"
    exit 1
  fi

  echo "TLD ${RESOLVER_TLD} points to ${host_ip}"
  echo
}

function create_proxy_nginx {
  local host_ip=$1

  if [[ -z "${host_ip}" ]]; then
    echo "misssing host ip"
    exit 1
  fi
  (docker ps -af "name=${CONTAINER_NGINX}" | grep "${CONTAINER_NGINX}" > /dev/null) && docker rm -f "${CONTAINER_NGINX}" > /dev/null
  docker run \
    -d \
    -p "${host_ip}:443:443" \
    -v /var/run/docker.sock:/tmp/docker.sock:ro \
    -v "${ROOT_DIR}/certs":/etc/nginx/certs:ro \
    -v "${ROOT_DIR}/vhost.d":/etc/nginx/vhost.d:ro \
    -v "${ROOT_DIR}/proxy.conf":/etc/nginx/proxy.conf:ro \
    --restart "always" \
    --name "${CONTAINER_NGINX}" \
    jwilder/nginx-proxy \
    > /dev/null
}

function set_hosts {
  local host_ip=$1

  if [[ -z "${host_ip}" ]]; then
    echo "misssing host ip"
    exit 1
  fi
  sed "/# bb-consent/d" /etc/hosts | sudo tee /etc/hosts > /dev/null
  echo -e "${host_ip}\t${BB_CONSENT_SUBDOMAINS} # bb-consent" | sudo tee -a /etc/hosts > /dev/null
}

detect_flags

## Prepare environment
if ! docker info >/dev/null 2>&1; then
  echo "Docker daemon is not running"
  if (( (ENVIRONMENT_FLAGS & FLAG_OSX)>0 )); then
    echo "Options:"
    echo

    set +e
    docker-machine inspect "${DOCKER_VM_NAME}" > /dev/null 2>&1
    machineExitCode=$?
    set -e

    if [[ $machineExitCode -gt 0 ]]; then
      echo " - start docker beta"
      echo " - create docker machine: run 'docker-machine create --driver virtualbox --virtualbox-memory 2048 ${DOCKER_VM_NAME}'"
    elif [[ -z "${DOCKER_HOST:-}" ]]; then
      echo " - set docker env: run 'eval \$(docker-machine env ${DOCKER_VM_NAME})'"
    else
      echo " - start docker machine: run 'docker-machine start ${DOCKER_VM_NAME}'"
    fi

    echo
  fi
  exit 1
fi

if (( (ENVIRONMENT_FLAGS & FLAG_DOCKER_MACHINE)>0 )); then
  DOCKER_IP=$(docker-machine ip "${DOCKER_VM_NAME}")
fi

if (( (ENVIRONMENT_FLAGS & FLAG_PROXY_NGINX)>0 )); then
  create_proxy_nginx "${DOCKER_IP}"
fi

set_hosts "${DOCKER_IP}"

echo "Done"
exit 0
