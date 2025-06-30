# Node Label Controller

A Kubernetes controller that automatically applies labels to nodes based on configurable selection strategies.

## Description

The Node Label Controller is a custom Kubernetes controller that manages node labeling through custom resources. It allows you to define policies that automatically select nodes and apply labels to them based on various strategies such as oldest, newest, or random selection.

## Features

- **Flexible Node Selection**: Choose nodes based on creation time (oldest/newest) or random selection
- **Automatic Label Management**: Apply and remove labels automatically based on policies
- **Policy-based Configuration**: Define labeling policies using Custom Resources
- **Status Tracking**: Monitor which nodes are currently labeled by each policy

## Usage

### Node Selection Strategies

The controller supports three node selection strategies:

- **oldest**: Selects the oldest nodes (earliest creation time)
- **newest**: Selects the newest nodes (latest creation time)
- **random**: Selects nodes randomly

### Example Policy

```yaml
apiVersion: nlp.lento.dev/v1alpha1
kind: NodeLabelPolicy
metadata:
  name: nodelabelpolicy-sample
spec:
  strategy:
    type: oldest
    count: 3
  labels:
    environment: production
    workload: critical
    team: devops
```

This policy will:
- Select the 3 oldest nodes in the cluster
- Apply the labels `environment=production`, `workload=critical`, and `team=devops`
- Automatically remove these labels from nodes that are no longer selected

## Getting Started

### Prerequisites
- Go version v1.24.0+
- Docker version 17.03+
- kubectl version v1.11.3+
- Access to a Kubernetes v1.11.3+ cluster

### Installation

1. **Build and push the image:**
```sh
make docker-build docker-push IMG=<your-registry>/node-label-controller:tag
```

2. **Install the CRDs:**
```sh
make install
```

3. **Deploy the controller:**
```sh
make deploy IMG=<your-registry>/node-label-controller:tag
```

4. **Apply sample policies:**
```sh
kubectl apply -k config/samples/
```

### Uninstallation

```sh
kubectl delete -k config/samples/
make uninstall
make undeploy
```

## Use Cases

### Production Environment Management
- Label production nodes with `environment=production`
- Select nodes based on creation time for workload placement

### Workload Distribution
- Distribute workloads across different node groups
- Use random selection for load balancing across nodes

### Team Resource Allocation
- Assign nodes to specific teams or projects
- Track resource ownership through labels

### Infrastructure Automation
- Integrate with CI/CD pipelines for automated node labeling
- Support infrastructure-as-code practices

### Cost Optimization with DaemonSets

A practical use case is combining DaemonSets with NodeSelectors and NodeLabelPolicy for cost optimization in dynamic environments.

**Scenario**: Limit monitoring agent installation to only the 5 oldest nodes to reduce licensing costs.

```yaml
# NodeLabelPolicy to select oldest nodes
apiVersion: nlp.lento.dev/v1alpha1
kind: NodeLabelPolicy
metadata:
  name: monitoring-nodes
spec:
  strategy:
    type: oldest
    count: 5
  labels:
    monitoring: enabled
    agent: datadog
---
# DaemonSet with NodeSelector
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: datadog-agent
spec:
  selector:
    matchLabels:
      name: datadog-agent
  template:
    metadata:
      labels:
        name: datadog-agent
    spec:
      nodeSelector:
        monitoring: enabled
        agent: datadog
      containers:
      - name: datadog-agent
        image: datadog/agent:latest
        # ... other configuration
```

**Considerations for Dynamic Environments**:

When using this strategy with node autoscalers like Karpenter, consider the following configurations for stability:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: your-workload
spec:
  replicas: 3
  selector:
    matchLabels:
      app: your-workload
  template:
    metadata:
      labels:
        app: your-workload
      annotations:
        karpenter.sh/do-not-disrupt: "true"
    spec:
      containers:
      - name: your-app
        image: your-app:latest
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
```

**Key Requirements**:
- **`karpenter.sh/do-not-disrupt: "true"`**: Prevents node autoscalers from terminating pods during scale-in operations
- **Horizontal Scaling**: Ensure your workloads can scale horizontally

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
