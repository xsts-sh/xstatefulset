# XStatefulSet Webhook Quick Reference

## Quick Start

### Deploy with Webhook
```bash
helm install xstatefulset ./charts/xstatefulset
```

### Verify Webhook is Running
```bash
kubectl get pods -l app.kubernetes.io/component=controller-manager
kubectl get svc xstatefulset-controller-manager-webhook
kubectl get mutatingwebhookconfiguration xstatefulset-mutating-webhook
```

### Test Webhook
```bash
kubectl apply -f examples/webhook/test-xstatefulset.yaml
kubectl get xstatefulset -o yaml
```

## Default Values Applied

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

## Configuration Options

### Helm Values
```yaml
webhook:
  enabled: true                    # Enable/disable webhook
  serviceName: xstatefulset-controller-manager-webhook
  port: 8443                       # Webhook server port
  certSecretName: xstatefulset-controller-manager-webhook-certs
  timeoutSeconds: 30               # Request timeout
  failurePolicy: Fail              # Fail or Ignore
  namespaceSelector: {}            # Target specific namespaces
  objectSelector: {}               # Target specific objects
```

### Command-Line Flags
```bash
--enable-webhook=true
--port=8443
--tls-cert-file=/etc/tls/tls.crt
--tls-key-file=/etc/tls/tls.key
--cert-secret-name=xstatefulset-controller-manager-webhook-certs
--service-name=xstatefulset-controller-manager-webhook
--webhook-timeout=30
```

## Common Tasks

### Disable Webhook
```bash
helm upgrade xstatefulset ./charts/xstatefulset --set webhook.enabled=false
```

### Change Failure Policy to Ignore
```bash
helm upgrade xstatefulset ./charts/xstatefulset --set webhook.failurePolicy=Ignore
```

### Target Specific Namespaces
```yaml
webhook:
  namespaceSelector:
    matchLabels:
      webhook: enabled
```

Then label namespace:
```bash
kubectl label namespace my-namespace webhook=enabled
```

### Regenerate Certificates
```bash
kubectl delete secret xstatefulset-controller-manager-webhook-certs
kubectl delete pod -l app.kubernetes.io/component=controller-manager
```

## Troubleshooting

### Check Webhook Logs
```bash
kubectl logs -l app.kubernetes.io/component=controller-manager --tail=100
```

### Verify CA Bundle
```bash
kubectl get mutatingwebhookconfiguration xstatefulset-mutating-webhook \
  -o jsonpath='{.webhooks[0].clientConfig.caBundle}' | base64 -d
```

### Test Webhook Connectivity
```bash
kubectl run test-curl --image=curlimages/curl --rm -it -- \
  curl -k https://xstatefulset-controller-manager-webhook.default.svc:8443/healthz
```

### Check Certificate Secret
```bash
kubectl get secret xstatefulset-controller-manager-webhook-certs -o yaml
```

### Temporarily Disable Webhook
```bash
kubectl delete mutatingwebhookconfiguration xstatefulset-mutating-webhook
```

## Minimal XStatefulSet Example

```yaml
apiVersion: apps.x-k8s.io/v1
kind: XStatefulSet
metadata:
  name: my-app
spec:
  selector:
    matchLabels:
      app: my-app
  serviceName: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: app
        image: my-app:latest
```

All defaults will be automatically applied!

## Testing

### Run Webhook Tests
```bash
go test ./pkg/webhook/... -v
```

### Build Controller
```bash
go build -o bin/xstatefulset-controller-manager ./cmd/main.go
```

## Endpoints

- `/mutate-xstatefulset` - Webhook endpoint
- `/healthz` - Health check
- `/readyz` - Readiness check

## Security

- TLS 1.2+ required
- Self-signed certificates (10-year validity)
- Automatic CA bundle injection
- Minimal RBAC permissions

## Documentation

- [Complete Setup Guide](webhook-setup.md)
- [Usage Examples](../examples/webhook/README.md)
- [Implementation Details](../WEBHOOK_IMPLEMENTATION.md)

## Support

For issues or questions:
1. Check webhook logs
2. Verify webhook configuration
3. Test connectivity
4. Review documentation
5. Check certificate validity
