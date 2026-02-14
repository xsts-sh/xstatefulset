# XStatefulSet Webhook with Controller-Runtime

This document describes the controller-runtime based mutating admission webhook for XStatefulSet resources.

## Overview

The XStatefulSet controller uses [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) to implement a mutating admission webhook that automatically applies default values to XStatefulSet resources when they are created or updated.

## Architecture

### Components

1. **XStatefulSetDefaulter** (`pkg/webhook/xstatefulset_webhook.go`)
   - Implements `webhook.CustomDefaulter` interface
   - Applies defaults using `legacyscheme.Scheme.Default()`
   - Registered with controller-runtime manager

2. **Controller-Runtime Manager** (`cmd/main.go`)
   - Manages webhook server lifecycle
   - Handles TLS certificate management
   - Provides health check endpoints

3. **Webhook Server**
   - Runs on port 9443 (configurable)
   - Serves at path `/mutate-apps-x-k8s-io-v1-xstatefulset`
   - Handles admission review requests

## Certificate Management

Controller-runtime supports multiple certificate sources:

### Option 1: cert-manager (Recommended)

Install cert-manager and create a Certificate resource:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: xstatefulset-webhook-cert
  namespace: default
spec:
  secretName: webhook-server-cert
  dnsNames:
    - xstatefulset-controller-manager-webhook.default.svc
    - xstatefulset-controller-manager-webhook.default.svc.cluster.local
  issuerRef:
    kind: Issuer
    name: selfsigned-issuer
```

### Option 2: Manual Certificates

Generate and create a secret manually:

```bash
# Generate certificates
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout tls.key -out tls.crt -days 365 \
  -subj "/CN=xstatefulset-controller-manager-webhook.default.svc"

# Create secret
kubectl create secret tls webhook-server-cert \
  --cert=tls.crt --key=tls.key
```

### Option 3: Controller-Runtime Auto-Generation

Controller-runtime can generate certificates automatically (development only):

```bash
# Certificates will be generated in /tmp/k8s-webhook-server/serving-certs
# Not recommended for production
```

## Configuration

### Helm Values

```yaml
webhook:
  enabled: true
  serviceName: xstatefulset-controller-manager-webhook
  port: 9443  # Internal pod port
  certSecretName: webhook-server-cert
  timeoutSeconds: 30
  failurePolicy: Fail
  namespaceSelector: {}
  objectSelector: {}
```

### Command-Line Flags

```bash
--enable-webhook=true                    # Enable webhook
--webhook-port=9443                      # Webhook server port
--webhook-cert-dir=/tmp/k8s-webhook-server/serving-certs  # Certificate directory
```

## Deployment

### Prerequisites

1. Kubernetes cluster with admission webhooks enabled
2. Certificate management solution (cert-manager recommended)
3. Network connectivity between API server and webhook service

### Installation with cert-manager

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create self-signed issuer
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: selfsigned-issuer
  namespace: default
spec:
  selfSigned: {}
EOF

# Create certificate
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: xstatefulset-webhook-cert
  namespace: default
spec:
  secretName: webhook-server-cert
  dnsNames:
    - xstatefulset-controller-manager-webhook.default.svc
    - xstatefulset-controller-manager-webhook.default.svc.cluster.local
  issuerRef:
    kind: Issuer
    name: selfsigned-issuer
EOF

# Install XStatefulSet controller
helm install xstatefulset ./charts/xstatefulset
```

### Installation without cert-manager

```bash
# Generate certificates
./scripts/generate-webhook-certs.sh

# Install XStatefulSet controller
helm install xstatefulset ./charts/xstatefulset
```

## Default Values Applied

The webhook automatically applies these defaults:

| Field | Default Value |
|-------|---------------|
| `spec.replicas` | `1` |
| `spec.podManagementPolicy` | `OrderedReady` |
| `spec.updateStrategy.type` | `RollingUpdate` |
| `spec.updateStrategy.rollingUpdate.partition` | `0` |
| `spec.updateStrategy.rollingUpdate.maxUnavailable` | `1` |
| `spec.revisionHistoryLimit` | `10` |
| `spec.persistentVolumeClaimRetentionPolicy.whenDeleted` | `Retain` |
| `spec.persistentVolumeClaimRetentionPolicy.whenScaled` | `Retain` |

## Testing

### Unit Tests

```bash
go test ./pkg/webhook/... -v
```

### Integration Testing

1. Deploy the controller:
```bash
helm install xstatefulset ./charts/xstatefulset
```

2. Create a minimal XStatefulSet:
```bash
kubectl apply -f examples/webhook/test-xstatefulset.yaml
```

3. Verify defaults were applied:
```bash
kubectl get xstatefulset test-xsts-minimal -o yaml
```

## Troubleshooting

### Webhook Not Working

1. **Check webhook pod logs:**
```bash
kubectl logs -l app.kubernetes.io/component=xstatefulset-controller-manager
```

2. **Verify webhook configuration:**
```bash
kubectl get mutatingwebhookconfiguration xstatefulset-mutating-webhook -o yaml
```

3. **Check certificate secret:**
```bash
kubectl get secret webhook-server-cert
kubectl describe secret webhook-server-cert
```

4. **Test webhook connectivity:**
```bash
kubectl run test-curl --image=curlimages/curl --rm -it -- \
  curl -k https://xstatefulset-controller-manager-webhook.default.svc:443/healthz
```

### Certificate Issues

1. **Check cert-manager logs (if using cert-manager):**
```bash
kubectl logs -n cert-manager -l app=cert-manager
```

2. **Verify certificate is ready:**
```bash
kubectl get certificate xstatefulset-webhook-cert
```

3. **Check certificate details:**
```bash
kubectl get secret webhook-server-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -text -noout
```

### Webhook Timeout

If webhook requests are timing out:

1. **Increase timeout in values.yaml:**
```yaml
webhook:
  timeoutSeconds: 60
```

2. **Check network policies:**
```bash
kubectl get networkpolicies
```

3. **Verify service endpoints:**
```bash
kubectl get endpoints xstatefulset-controller-manager-webhook
```

## Advantages of Controller-Runtime

1. **Standard Pattern**: Follows Kubernetes operator best practices
2. **Less Code**: ~50 lines vs. 300+ lines for custom implementation
3. **Built-in Features**:
   - Automatic certificate rotation
   - Health check endpoints
   - Metrics integration
   - Leader election support
4. **Better Testing**: Integration with envtest
5. **Community Support**: Well-documented and widely used

## Comparison with Custom Implementation

| Feature | Controller-Runtime | Custom Implementation |
|---------|-------------------|----------------------|
| Lines of Code | ~50 | ~300+ |
| Certificate Management | Built-in | Manual |
| Health Checks | Built-in | Manual |
| Metrics | Built-in | Manual |
| Testing | envtest support | Custom |
| Maintenance | Low | High |
| Community Support | High | Low |

## References

- [Controller-Runtime Documentation](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Kubernetes Admission Webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
- [cert-manager Documentation](https://cert-manager.io/docs/)
- [Kubebuilder Book](https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation.html)
