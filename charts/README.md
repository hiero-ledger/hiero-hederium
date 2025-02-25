# Hederium Helm Chart

This directory contains the Helm chart for deploying Hederium, a Hedera JSON-RPC relay service, to Kubernetes.

## Chart Overview

The Hederium chart deploys the following Kubernetes resources:

- **Deployment**: Runs the Hederium application with the specified number of replicas
- **Service**: Exposes the application via a LoadBalancer
- **ConfigMap**: Stores the application configuration
- **ServiceAccount**: Identity for the application pods

Optional components (disabled by default):
- **HorizontalPodAutoscaler**: Automatically scales the application based on CPU usage
- **Ingress**: Provides advanced routing capabilities

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- Minikube (for local deployment)
- Docker

## Deploying to Minikube

### 1. Start Minikube

```bash
minikube start --cpus=2 --memory=4096 --disk-size=20g
```

### 2. Build the Docker Image in Minikube

```bash
# Configure your terminal to use Minikube's Docker daemon
eval $(minikube docker-env)

# Build the image
docker build -t hederium:latest ../
```

### 3. Install the Helm Chart

```bash
# From the root of the repository
helm install hederium ./charts
```

### 4. Access the Application

Since we're using a LoadBalancer service in Minikube, you need to run:

```bash
# In a separate terminal window
minikube tunnel
```

Then get the service URL:

```bash
kubectl get svc hederium
```

Or use the Minikube service command:

```bash
minikube service hederium --url
```

## Configuration

### Key Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `2` |
| `image.repository` | Image repository | `hederium` |
| `image.tag` | Image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `service.type` | Service type | `LoadBalancer` |
| `service.port` | Service port | `7546` |
| `autoscaling.enabled` | Enable autoscaling | `false` |
| `ingress.enabled` | Enable ingress | `false` |
| `config.environment` | Application environment | `development` |
| `config.hedera.network` | Hedera network | `testnet` |

### Custom Values

To override default values, create a custom values file:

```yaml
# custom-values.yaml
replicaCount: 3
config:
  environment: "production"
  hedera:
    network: "mainnet"
```

Then install/upgrade with:

```bash
helm install hederium ./charts -f custom-values.yaml
# or
helm upgrade hederium ./charts -f custom-values.yaml
```

## Monitoring and Debugging

### View Pods

```bash
kubectl get pods -l app.kubernetes.io/instance=hederium
```

### Check Logs

```bash
kubectl logs -l app.kubernetes.io/instance=hederium
```

### Using K9s

For a better terminal UI experience, install and use K9s:

```bash
# Install K9s
brew install k9s  # macOS
# or
sudo snap install k9s  # Linux

# Run K9s
k9s
```

Note: To see CPU and memory metrics in K9s, enable the metrics-server in Minikube:

```bash
minikube addons enable metrics-server
```

## Uninstalling the Chart

```bash
helm uninstall hederium
```

## Upgrading the Chart

```bash
# After making changes to the application
eval $(minikube docker-env)
docker build -t hederium:latest ../

# Upgrade the Helm release
helm upgrade hederium ./charts
```

## Troubleshooting

### Common Issues

1. **Image pull errors**: Ensure you've built the image in Minikube's Docker environment.
2. **Service shows `<pending>`**: Run `minikube tunnel` in a separate terminal.
3. **Configuration issues**: Check the logs to see if the application is reading the config correctly.
4. **N/A metrics in K9s**: Enable the metrics-server addon in Minikube.

For more help, check the application logs or describe the pods:

```bash
kubectl describe pod -l app.kubernetes.io/instance=hederium
``` 