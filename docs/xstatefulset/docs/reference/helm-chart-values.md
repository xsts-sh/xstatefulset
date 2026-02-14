# xstatefulset

A Helm subchart for XStatefulSet.

![Version: 1.0.0](https://img.shields.io/badge/Version-1.0.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.0.0](https://img.shields.io/badge/AppVersion-1.0.0-informational?style=flat-square)

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| controllerManager.image.args[0] | string | `"--v=2"` |  |
| controllerManager.image.pullPolicy | string | `"IfNotPresent"` |  |
| controllerManager.image.repository | string | `"ghcr.io/volcano-sh/xstatefulset-controller-manager"` |  |
| controllerManager.image.tag | string | `"latest"` |  |
| controllerManager.kubeAPIBurst | int | `0` |  |
| controllerManager.kubeAPIQPS | int | `0` |  |
| controllerManager.replicas | int | `1` |  |
| controllerManager.resource.limits.cpu | string | `"500m"` |  |
| controllerManager.resource.limits.memory | string | `"512Mi"` |  |
| controllerManager.resource.requests.cpu | string | `"100m"` |  |
| controllerManager.resource.requests.memory | string | `"128Mi"` |  |
| global.certManagementMode | string | `"auto"` | Certificate Management Mode.<br/>  Three mutually exclusive options for managing TLS certificates:<br/>  - `auto`: Webhook servers generate self-signed certificates automatically.<br/>  - `cert-manager`: Use cert-manager to generate and manage certificates (requires cert-manager installation).<br/>  - `manual`: Provide your own certificates via caBundle. |
| global.webhook.caBundle | string | `""` | CA bundle for webhook server certificates (base64-encoded).<br/> This is ONLY required when `certManagementMode` is set to "manual".<br/> You can generate it with: `cat /path/to/your/ca.crt | base64 | tr -d '\n'`<br/> |
| webhook.enabled | bool | `true` |  |
| webhook.failurePolicy | string | `"Fail"` |  |
| webhook.namespaceSelector | object | `{}` |  |
| webhook.objectSelector | object | `{}` |  |
| webhook.port | int | `8443` |  |
| webhook.serviceName | string | `"xstatefulset-controller-manager-webhook"` |  |
| webhook.timeoutSeconds | int | `30` |  |
| webhook.tls.certSecretName | string | `"webhook-server-cert"` | Secret name for storing webhook certificates. |

## Notes

- Values marked as “usually set by CI” are automatically updated during the release process; manual changes are not required.
- For detailed information about each component, refer to the corresponding architecture and user guide documents.
- Always review the [values.yaml](https://github.com/xsts-sh/xstatefulset/blob/main/charts/xstatefulset/values.yaml) file in the repository for the latest defaults and available options.