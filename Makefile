CONTAINER_RUNTIME ?= "podman"

.PHONY: up
up: keycloak-certificate keycloak generate-config webhook-certificate webhook cluster token

.PHONY: down
down:
	${CONTAINER_RUNTIME} stop ${KEYCLOAK_CONTAINER_NAME} || true
	${CONTAINER_RUNTIME} stop ${WEBHOOK_CONTAINER_NAME} || true
	kind delete cluster --name padlok-dev

KEYCLOAK_IMAGE ?= "quay.io/keycloak/keycloak:latest"
KEYCLOAK_CONTAINER_NAME ?= "keycloak"
KEYCLOAK_ADMIN_USERNAME ?= "admin"
KEYCLOAK_ADMIN_PASSWORD ?= "admin"
.PHONY: keycloak
keycloak:
	${CONTAINER_RUNTIME} stop ${KEYCLOAK_CONTAINER_NAME} || true
	${CONTAINER_RUNTIME} wait ${KEYCLOAK_CONTAINER_NAME} || true
	${CONTAINER_RUNTIME} run -d --rm --name ${KEYCLOAK_CONTAINER_NAME} -p 127.0.0.1:8443:8443 --network=kind \
		-e KC_BOOTSTRAP_ADMIN_USERNAME=${KEYCLOAK_ADMIN_USERNAME} \
		-e KC_BOOTSTRAP_ADMIN_PASSWORD=${KEYCLOAK_ADMIN_PASSWORD} \
		-e KC_HTTPS_CERTIFICATE_FILE=/certs/cert.pem \
		-e KC_HTTPS_CERTIFICATE_KEY_FILE=/certs/key.pem \
		-v $(shell pwd)/dev/certs:/certs/ \
		-v $(shell pwd)/dev/keycloak:/opt/keycloak/data/import/ \
		${KEYCLOAK_IMAGE} start-dev --import-realm

.PHONY: keycloak-certificate
keycloak-certificate:
	rm -rf dev/certs/keycloak/*
	mkdir -p dev/certs/keycloak
	openssl req -x509 -newkey rsa:4096 -keyout dev/certs/keycloak/key.pem -out dev/certs/keycloak/cert.pem -sha256 -days 365 -nodes -subj "/C=US/O=Padlok/OU=PadlokProject/CN=${KEYCLOAK_CONTAINER_NAME}" -addext "subjectAltName=DNS:${KEYCLOAK_CONTAINER_NAME}"

.PHONY: build
build:
	mkdir -p bin
	go build -o bin/padlok main.go

IMAGE_TAG ?= "quay.io/rh_ee_bpalmer/padlok:latest"
.PHONY: image
image:
	${CONTAINER_RUNTIME} build -t ${IMAGE_TAG} .

CLIENT_ID ?= "k8s-client"
ISSUER ?= "https://keycloak:8443/realms/k8s"
TOKEN_TOOL_IMAGE ?= "quay.io/rh_ee_bpalmer/oauth2cli:latest"
.PHONY: token
token:
	${CONTAINER_RUNTIME} run --rm --network=kind ${TOKEN_TOOL_IMAGE} device-code --issuer ${ISSUER} --client-id ${CLIENT_ID}

WEBHOOK_CONTAINER_NAME ?= padlok

.PHONY: webhook-certificate
webhook-certificate:
	rm -rf dev/certs/padlok/*
	mkdir -p dev/certs/padlok
	openssl req -x509 -newkey rsa:4096 -keyout dev/certs/padlok/key.pem -out dev/certs/padlok/cert.pem -sha256 -days 365 -nodes -subj "/C=US/O=Padlok/OU=PadlokProject/CN=${WEBHOOK_CONTAINER_NAME}" -addext "subjectAltName=DNS:${WEBHOOK_CONTAINER_NAME}"

.PHONY: webhook
webhook:
	${CONTAINER_RUNTIME} stop ${WEBHOOK_CONTAINER_NAME} || true
	${CONTAINER_RUNTIME} wait ${WEBHOOK_CONTAINER_NAME} || true
	${CONTAINER_RUNTIME} run -d --rm --name ${WEBHOOK_CONTAINER_NAME} --network=kind \
		-v $(shell pwd)/dev/cfg/config.yaml:/cfg/config.yaml:Z \
		-v $(shell pwd)/dev/certs/padlok/key.pem:/certs/key.pem \
		-v $(shell pwd)/dev/certs/padlok/cert.pem:/certs/cert.pem \
		${IMAGE_TAG} run --config=/cfg/config.yaml --tls-private-key-file=/certs/key.pem --tls-cert-file=/certs/cert.pem

define CONFIG_TEMPLATE
apiVersion: padlok.everettraven.github.io/v1alpha1
kind: AuthenticationConfiguration
jwt:
  - issuer:
      url: https://keycloak:8443/realms/k8s
      audiences:
        - k8s-client
      certificateAuthority: |
$${KEYCLOAK_CERTIFICATE_AUTHORITY}
    claimMappings:
      username:
       claim: "preferred_username"
       prefix: ""
      groups:
        expression: "claims.groups.split(',')"
    externalClaimsSources:
    - authentication:
        type: RequestProvidedToken
      tls:
        certificateAuthority: |
$${KEYCLOAK_CERTIFICATE_AUTHORITY}
      url:
        hostname: keycloak:8443
        pathExpression: "['realms', 'k8s', 'protocol', 'openid-connect', 'userinfo']"
      mappings:
        - name: groups
          expression: "response.body.groups.join(',')"
      conditions:
        - expression: "!has(claims.groups)"
endef
export CONFIG_TEMPLATE

.PHONY: generate-config
generate-config:
	rm -rf dev/cfg/*
	mkdir -p dev/cfg
	echo "$${CONFIG_TEMPLATE}" > dev/cfg/config-templ.yaml
	export KEYCLOAK_CERTIFICATE_AUTHORITY=$$(sed 's/^/          /' dev/certs/cert.pem); \
	envsubst < dev/cfg/config-templ.yaml > dev/cfg/config.yaml
	rm -f dev/cfg/config-templ.yaml

.PHONY: cluster
cluster:
	kind delete cluster --name padlok-dev
	kind create cluster --name padlok-dev --config dev/kind/config.yaml

.PHONY: unit
unit:
	go test $(shell go list ./... | grep -v "internal/third_party")
