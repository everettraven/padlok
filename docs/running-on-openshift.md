# Running on OpenShift

This page documents the process of installing `padlok` on an OpenShift cluster and configuring OpenShift to use it.

## Prerequisites

- A running OpenShift cluster.
- An identity provider you'd like to configure `padlok` to use.

## Installing `padlok` on the OpenShift cluster

This section breaks down the different pieces you'll need to install `padlok` on an OpenShift cluster.

It provides examples of the raw YAML to apply for each step.

### 1: Create the `padlok` namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: padlok
```

### 2: Create a Secret with your `padlok` configuration

> [!NOTE]
> This is just an example configuration that was used for documentation purposes.
> It is your responsibility to create a valid configuration for your identity provider.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: padlok-config
  namespace: padlok
type: Opaque
stringData:
  config.yaml: |
    apiVersion: padlok.everettraven.github.io/v1alpha1
    kind: AuthenticationConfiguration
    jwt:
      - issuer:
          url: https://idp.example.com
          audiences:
            - openshift
          certificateAuthority: ${CERTIFICATE_AUTHORITY}
        claimMappings:
          username:
            expression: "claims.preferred_username"
          groups:
            expression: "claims.groups.split(',')"
        externalClaimsSources:
        - authentication:
            type: RequestProvidedToken
          tls:
            certificateAuthority: ${CERTIFICATE_AUTHORITY}
          url:
            hostname: idp.example.com
            pathExpression: "['realms', 'openshift', 'protocol', 'openid-connect', 'userinfo']"
          mappings:
            - name: groups
              expression: "response.body.groups.join(',')"
          conditions:
            - expression: "!has(claims.groups)"
```

### 3: Create the `padlok` Service

```yaml
apiVersion: v1
kind: Service
metadata:
  namespace: padlok
  name: padlok-server
  annotations:
    service.beta.openshift.io/serving-cert-secret-name: serving-cert
  labels:
    app: padlok
spec:
  selector:
    app: padlok
  ports:
    - name: https
      port: 443
      targetPort: 6443
```

### 4: Create the `padlok` Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: padlok-server
  namespace: padlok
  labels:
    app: padlok
spec:
  selector:
    matchLabels:
      app: padlok
  replicas: 3
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/default-container: padlok-server
      labels:
        app: padlok
    spec:
      containers:
        - name: padlok-server
          image: quay.io/rh_ee_bpalmer/padlok:latest
          imagePullPolicy: Always
          command: ["./padlok", "run"]
          args:
            - --secure-port=6443
            - --tls-private-key-file=/var/run/secrets/serving-cert/tls.key
            - --tls-cert-file=/var/run/secrets/serving-cert/tls.crt
            - --config=/var/run/secrets/padlok-config/config.yaml
          resources:
            requests:
              cpu: 100m
              memory: 100Mi
          ports:
            - containerPort: 6443
          volumeMounts:
            - name: padlok-config
              mountPath: /var/run/secrets/padlok-config
            - name: padlok-serving-cert
              mountPath: /var/run/secrets/serving-cert
      volumes:
        - name: padlok-config
          secret:
            secretName: padlok-config
            items:
              - key: config.yaml
                path: config.yaml
        - name: padlok-serving-cert
          secret:
            secretName: serving-cert
      restartPolicy: Always
```

## Configuring OpenShift to use `padlok`

This section breaks down the different pieces you'll need to configure an OpenShift cluster to use `padlok` as a webhook authenticator.

It provides examples of the commands to run and YAML templates to apply for each step.

### 1: Get the OpenShift `service-ca-operator` signing CA bundle

Because the `padlok` server runs locally on the cluster, we leverage OpenShift's `service-ca-operator` to provision the serving certificates for `padlok`.

In order for the kube-apiserver to trust the `padlok` server's certificate, we have to provide the kube-apiserver with the CA bundle to verify it with.

You can fetch the `service-ca-operator`'s signing CA bundle with the following command:

```sh
kubectl -n openshift-service-ca get configmaps/signing-cabundle -o json | jq -r '.data."ca-bundle.crt" | @base64'
```

Keep the output of this command handy, you'll need it later when we are writing the kubeconfig for communicating with the `padlok` server.

### 2: Get the cluster IP for the `padlok` Service

Use the following command to get the internal IP address for the `padlok` Service that we will use to tell the kube-apiserver where to make authentication requests to.

```sh
kubectl -n padlok get svc
```

The output should look something like:

```sh
NAME            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
padlok-server   ClusterIP   172.30.35.220   <none>        443/TCP   37s
```

### 3: Build the kubeconfig

Now that we have all the information we need, we can create the kubeconfig file that the kube-apiserver will use to interact with the webhook authenticator.

In a new YAML file, populate the following:

```yaml
apiVersion: v1
clusters:
  - cluster:
      certificate-authority-data: ${SERVICE_CA_OPERATOR_SIGNING_BUNDLE}
      server: https://${CLUSTER_IP}/authenticate
      tls-server-name: padlok-server.padlok.svc
    name: padlok
contexts:
  - context:
      cluster: padlok
      user: apiserver
    name: local-cluster
current-context: local-cluster
kind: Config
users:
  - name: apiserver
    user:
      token: blah
```

Replace the `${SERVICE_CA_OPERATOR_SIGNING_BUNDLE}` placeholder with the output of the command run in step 1.

Replace the `${CLUSTER_IP}` placeholder with the output of the command run in step 2.

> [!NOTE]
> OpenShift expects any webhook authenticator configuration to specify exactly one user.
> Because `padlok` does not yet have any authentication in place for accessing the server,
> we intentionally use a placeholder user.

### 4: Create the webhook Secret

OpenShift requires that the kubeconfig for interacting with the webhook is placed in a Secret that it copies over and mounts to the kube-apiserver pods.

First, we need to base64 encode the kubeconfig contents with a command like:

```sh
cat kubeconfig.yaml | base64
```

Once you've got the base64 encoded contents, create a secret like:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: padlok-webhook-config
  namespace: openshift-config
type: Opaque
data:
  kubeConfig: ${BASE64_ENCODED_KUBECONFIG}
```

### 5: Update the `Authentication` resource

Now that we've got the `padlok` server deployed and we have the configuration ready for the kube-apiserver to delegate authentication decisions to it,
we need to tell OpenShift to configure the kube-apiserver with our configuration.

To do that, we need to apply the following `Authentication` resource changes:

```yaml
apiVersion: config.openshift.io/v1
kind: Authentication
metadata:
  name: cluster
spec:
  type: None
  webhookTokenAuthenticator:
    kubeConfig:
      name: padlok-webhook-config
```

> [!WARNING]
> Setting `.spec.type: None` will disable OpenShift's built-in OAuth server.
> By enabling this configuration, any users that logged in via the built-in OAuth server will no longer be authenticated.
> This may also result in a degraded OpenShift Console experience as the OpenShift Console will not know how it needs to log a user in to the cluster.
> Use this configuration at your own risk.
> If `padlok` has been misconfigured, you will only be able to access the cluster using certificate-based authentication.
> To revert the configuration change, set `.spec.type: ""`.

### Verify configuration

Once the OpenShift cluster has been configured, you'll need to use an OAuth2 client to fetch an ID token or an access token and configure your local kubeconfig to use the token to authenticate with the cluster.

Once you have that, you can use `kubectl auth whoami` to see the user information that `padlok` has resolved for the token.
