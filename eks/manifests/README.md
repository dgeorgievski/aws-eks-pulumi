# Customization of the cluster

## Metrics server
Ref: https://github.com/kubernetes-sigs/metricsdoc-server

Deployment
```shell
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/high-availability.yaml

```

## RBAC roles

```shell
k apply -f teams/namespaces.yaml
k apply -f teams/team-role-bindings.yaml
k apply -f teams/rbac-cluster-readonly.yaml
```

## Kubed
Replicate sercrets
See for more https://appscode.com/products/kubed/v0.12.0/setup/install/
```
helm install kubed appscode/kubed \
  --version v0.12.0 \
  --namespace kube-system \
  --debug --dry-run
```

## Cert-manager
```shell
helm upgrade -i \
cert-manager cert-manager \
--namespace cert-manager \
--create-namespace \
--set installCRDs=true \
--debug --dry-run
```

Install cert issuers
```shell
k apply -f certificates/ClusterIssuer-staging.yaml
k apply -f certificates/ClusterIssuer.yaml

k get ClusterIssuers
NAME                  READY   AGE
letsencrypt           True    24s
letsencrypt-staging   True    43s

```

Install Certificates
```shell
# Staging
k apply -f certificates/Certificate-staging.yaml

# test
k get certificates,orders

k get certificates
NAME          READY   SECRET               AGE
dex-staging   True    dex-staging-secret   77s

# Prod dex and login
k apply -f certificates/Certificate-dex-prod.yaml
k apply -f certificates/Certificate-login-prod.yaml
```

## Nginx Ingress

Reference: https://kubernetes.github.io/ingress-nginx/deploy/#quick-start

Public
```shell
helm upgrade -i \
ingress-nginx ingress-nginx \
--repo https://kubernetes.github.io/ingress-nginx \
--namespace ingress-nginx --create-namespace \
-f ingress-external-values.yaml \
--debug --dry-run
```

Private
```shell
helm upgrade -i \
internal-nginx ingress-nginx \
--repo https://kubernetes.github.io/ingress-nginx \
--namespace internal-nginx  --create-namespace \
-f ingress-internal-values.yaml \
--debug --dry-run

```

## External DNS
Ref: https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/aws.md

Install
```shell
k apply -f external-dns/deployment.yaml

```

## Dex

Install
```shell
helm upgrade dex dex/ \
--install \
--namespace dex \
--create-namespace \
--values dex.yaml \
--debug --dry-run
```

## k8s dex authenticator

Install
```shell
helm upgrade -i \
dex-k8s-authenticator dex-k8s-authenticator \
--namespace dex-k8s-authenticator \
--create-namespace \
--values dex-k8s-authenticator.yaml \
--debug --dry-run

```

## Tekton Pipelines

Ref: https://tekton.dev/docs/getting-started/tasks/

Install Pipelines
```shell
kubectl apply --filename \
https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
```

Install CLI:
- Instructions: https://tekton.dev/docs/cli/
- Releases: https://github.com/tektoncd/cli/releases
```shell
# Debian/Ubuntu package
sudo apt update;sudo apt install -y gnupg
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 3EFE0E0A2F2F60AA
echo "deb http://ppa.launchpad.net/tektoncd/cli/ubuntu eoan main"|sudo tee /etc/apt/sources.list.d/tektoncd-ubuntu-cli.list
sudo apt update && sudo apt install -y tektoncd-cli
```

Enable tkn as a kubectl plugin
```shell
ln -s /usr/bin/tkn /usr/local/bin/kubectl-tkn
kubectl plugin list

kubectl tkn
```

## Escalator

An interesting clsuter autoscaler in alpha stage, best suited for jobs workloads.
Ref: https://github.com/atlassian/escalator/tree/master/docs/deployment

```shell
# Create RBAC configuration
kubectl create -f docs/deployment/escalator-rbac.yaml

# Create config map - modify to suit your needs
kubectl create -f docs/deployment/escalator-cm.yaml

# Create deployment
kubectl create -f docs/deployment/escalator-deployment.yaml
```
