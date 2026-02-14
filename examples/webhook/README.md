# XStatefulSet Webhook Examples

This directory contains examples demonstrating the XStatefulSet mutating webhook functionality.

## Overview

The XStatefulSet controller includes a mutating admission webhook that automatically applies default values to XStatefulSet resources. This ensures consistent configuration and reduces boilerplate in your manifests.

## Quick Start

### 1. Install the Controller with Webhook

```bash
# Install using Helm
helm install xstatefulset ../../charts/xstatefulset

# Verify the webhook is running
kubectl get pods -l app.kubernetes.io/component=controller-manager
kubectl get svc xstatefulset-controller-manager-webhook
kubectl get mutatingwebhookconfiguration xstatefulset-mutating-webhook
```

### 2. Create a Minimal XStatefulSet

The webhook allows you to create XStatefulSets with minimal configuration:

```yaml
apiVersion: apps.x-k8s.io/v1
kind: XStatefulSet
metadata:
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  serviceName: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
        ports:
        - containerPort: 80
```

Apply it:

```bash
kubectl apply -f test-xstatefulset.yaml
```

### 3. Verify Defaults Were Applied

Check the created XStatefulSet:

```bash
kubectl get xstatefulset nginx -o yaml
```

You should see the following defaults automatically applied:

```yaml
spec:
  replicas: 1  # Default
  revisionHistoryLimit: 10  # Default
  podManagementPolicy: OrderedReady  # Default
  updateStrategy:
    type: RollingUpdate  # Default
    rollingUpdate:
      partition: 0  # Default
      maxUnavailable: 1  # Default
  persistentVolumeClaimRetentionPolicy:
    whenDeleted: Retain  # Default
    whenScaled: Retain  # Default
```

## Examples

### Example 1: Minimal Configuration

File: `test-xstatefulset.yaml` (first resource)

This example shows the absolute minimum required to create an XStatefulSet. The webhook will apply all defaults.

**What gets defaulted:**
- `replicas: 1`
- `podManagementPolicy: OrderedReady`
- `updateStrategy.type: RollingUpdate`
- `updateStrategy.rollingUpdate.partition: 0`
- `updateStrategy.rollingUpdate.maxUnavailable: 1`
- `revisionHistoryLimit: 10`
- `persistentVolumeClaimRetentionPolicy.whenDeleted: Retain`
- `persistentVolumeClaimRetentionPolicy.whenScaled: Retain`

### Example 2: With Persistent Volume Claims

File: `test-xstatefulset.yaml` (second resource)

This example includes volume claim templates. The webhook still applies defaults for fields not specified.

```bash
kubectl apply -f test-xstatefulset.yaml
kubectl get xstatefulset test-xsts-with-pvc -o yaml
```

### Example 3: Custom Configuration

File: `test-xstatefulset.yaml` (third resource)

This example explicitly sets some values. The webhook will:
- Keep your explicit values (e.g., `replicas: 3`, `podManagementPolicy: Parallel`)
- Apply defaults only for missing fields

```bash
kubectl apply -f test-xstatefulset.yaml
kubectl get xstatefulset test-xsts-custom -o yaml
```

## Testing the Webhook

### Test 1: Verify Webhook is Active

```bash
# Create a test XStatefulSet without defaults
cat <<EOF | kubectl apply -f -
apiVersion: apps.x-k8s.io/v1
kind: XStatefulSet
metadata:
  name: webhook-test
spec:
  selector:
    matchLabels:
      app: test
  serviceName: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: busybox
        image: busybox:latest
        command: ["sleep", "3600"]
EOF

# Check if defaults were applied
kubectl get xstatefulset webhook-test -o jsonpath='{.spec.replicas}'
# Should output: 1

kubectl get xstatefulset webhook-test -o jsonpath='{.spec.podManagementPolicy}'
# Should output: OrderedReady

# Clean up
kubectl delete xstatefulset webhook-test
```

### Test 2: Verify Explicit Values Are Preserved

```bash
# Create with explicit replicas
cat <<EOF | kubectl apply -f -
apiVersion: apps.x-k8s.io/v1
kind: XStatefulSet
metadata:
  name: explicit-test
spec:
  replicas: 5
  selector:
    matchLabels:
      app: test
  serviceName: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: busybox
        image: busybox:latest
        command: ["sleep", "3600"]
EOF

# Verify replicas is 5 (not overridden to 1)
kubectl get xstatefulset explicit-test -o jsonpath='{.spec.replicas}'
# Should output: 5

# Clean up
kubectl delete xstatefulset explicit-test
```

### Test 3: Check Webhook Logs

```bash
# View webhook logs
kubectl logs -l app.kubernetes.io/component=controller-manager --tail=50

# You should see log entries like:
# "Applying defaults to XStatefulSet: default/webhook-test"
```

## Troubleshooting

### Webhook Not Applying Defaults

1. **Check webhook pod status:**
   ```bash
   kubectl get pods -l app.kubernetes.io/component=controller-manager
   ```

2. **Check webhook logs:**
   ```bash
   kubectl logs -l app.kubernetes.io/component=controller-manager
   ```

3. **Verify webhook configuration:**
   ```bash
   kubectl get mutatingwebhookconfiguration xstatefulset-mutating-webhook -o yaml
   ```

4. **Check if CA bundle is set:**
   ```bash
   kubectl get mutatingwebhookconfiguration xstatefulset-mutating-webhook \
     -o jsonpath='{.webhooks[0].clientConfig.caBundle}' | base64 -d
   ```

### XStatefulSet Creation Fails

1. **Check webhook service:**
   ```bash
   kubectl get svc xstatefulset-controller-manager-webhook
   ```

2. **Test webhook connectivity:**
   ```bash
   kubectl run test-curl --image=curlimages/curl --rm -it -- \
     curl -k https://xstatefulset-controller-manager-webhook.default.svc:8443/healthz
   ```

3. **Check certificate secret:**
   ```bash
   kubectl get secret xstatefulset-controller-manager-webhook-certs
   ```

### Disable Webhook Temporarily

If you need to disable the webhook:

```bash
# Delete the webhook configuration
kubectl delete mutatingwebhookconfiguration xstatefulset-mutating-webhook

# Or set failurePolicy to Ignore
kubectl patch mutatingwebhookconfiguration xstatefulset-mutating-webhook \
  --type='json' -p='[{"op": "replace", "path": "/webhooks/0/failurePolicy", "value": "Ignore"}]'
```

## Advanced Usage

### Custom Namespace Selector

To only apply the webhook to specific namespaces:

```yaml
# In values.yaml
webhook:
  namespaceSelector:
    matchLabels:
      webhook: enabled
```

Then label your namespace:

```bash
kubectl label namespace default webhook=enabled
```

### Custom Object Selector

To only apply the webhook to XStatefulSets with specific labels:

```yaml
# In values.yaml
webhook:
  objectSelector:
    matchLabels:
      webhook: enabled
```

Then add the label to your XStatefulSet:

```yaml
metadata:
  labels:
    webhook: enabled
```

## Cleanup

```bash
# Delete all test XStatefulSets
kubectl delete xstatefulset --all

# Uninstall the controller
helm uninstall xstatefulset
```

## References

- [Main Webhook Documentation](../../docs/webhook-setup.md)
- [XStatefulSet API Reference](../../docs/xstatefulset/docs/reference/crd/apps.x-k8s.io.md)
