
# Install

## References

### AWS IAM authorization for Pods
https://docs.aws.amazon.com/eks/latest/userguide/specify-service-account-role.html

## pulumi up - provisioning
aws eks describe-cluster --name perf-test-nonprod --query "cluster.identity.oidc.issuer" --output text
https://oidc.eks.us-east-1.amazonaws.com/id/85466A2067925D6F4D52C11120375E10

https://oidc.eks.{region}.amazonaws.com/id/{identity}

aws iam list-open-id-connect-providers | grep 85466A2067925D6F4D52C11120375E10

### Nginx Ingress Conrollers
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.2/deploy/static/provider/cloud/deploy.yaml

External Nginx
```
helm upgrade --install \
ingress-nginx ingress-nginx \
--namespace ingress-nginx --create-namespace \
-f ingress-external-values.yaml \
--debug --dry-run
```

Verify
```
$ kubectl get pods,svc -n ingress-nginx
```

Internal Nginx
```
helm upgrade --install \
internal-ingress-nginx ingress-nginx \
--namespace internal-ingress-nginx --create-namespace \
-f ingress-internal-values.yaml \
--debug --dry-run
```
Verify
```
$ k get pods,svc -n internal-ingress-nginx
```

Logs
```
stern -n ingress-nginx -l app.kubernetes.io/instance=ingress-nginx
stern -n internal-ingress-nginx -l app.kubernetes.io/instance=internal-ingress-nginx
```

## cert-manager

stern -n cert-manager -l app=cert-manager

$ kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml

$ kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
stern -n cert-manager -l app=cert-manager

k -n cert-manager annotate sa/cert-manager \
"eks.amazonaws.com/role-arn"="arn:aws:iam::864590937242:role/perf-test-iam-cm"
 serviceaccount/cert-manager annotated

# default in v1.22 - not necessary
k -n cert-manager annotate sa/cert-manager \
eks.amazonaws.com/sts-regional-endpoints=true

k get sa cert-manager -o yaml
apiVersion: v1
automountServiceAccountToken: true
kind: ServiceAccount
metadata:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::864590937242:role/perf-test-iam-cm

k apply -f Certificate-staging.yaml
k apply -f Certificate-dex-prod.yaml
k apply -f Certificate-login-prod.yaml

# Kubed to replicate secrets
Reference:
- https://appscode.com/products/kubed/0.8.0/setup/install/
- https://github.com/kubeops/config-syncer/tree/v0.12.0/charts/kubed

1. create secret
helm install kubed appscode/kubed \
  --version v0.12.0 \
  --set nameOverride=perf-test \
  --namespace kube-system \
  --debug --dry-run

$ kubectl get deployment --namespace kube-system -l "app.kubernetes.io/name=perf-test,app.kubernetes.io/instance=kubed"
NAME              READY   UP-TO-DATE   AVAILABLE   AGE
kubed-perf-test   1/1     1            1           15s


## annotate secrets

Only if needed. Annotations are added by Certificate deployment.
```
k -n cert-manager annotate secret dex-staging-secret \
kubed.appscode.com/sync="kubed=dex"

k -n cert-manager annotate secret login-prod-secret \
kubed.appscode.com/sync="kubed=dex"
``

sync with ingress-nginx namespace
```
k get secrets -n ingress-nginx
k label ns/ingress-nginx kubed=dex
k label ns/dex kubed=dex
k label ns/dex kubed=login
```

Verify secrets are copied to dex namespace
```
k get secrets -n dex
```

## Dex
``
helm upgrade dex dex/ \
--install --namespace dex \
--create-namespace \
--values dex.yaml \
--debug --dry-run
```

Verify Dex certificate
```
$ curl -vL https://dex.nonprod.mozart.wiley.host
*   Trying 3.211.164.53:443...
* Connected to dex.nonprod.mozart.wiley.host (3.211.164.53) port 443 (#0)
* ALPN, offering h2
...

* Server certificate:
*  subject: CN=dex.nonprod.mozart.wiley.host
*  start date: Apr  8 20:01:12 2022 GMT
*  expire date: Jul  7 20:01:11 2022 GMT
*  subjectAltName: host "dex.nonprod.mozart.wiley.host" matched cert's "dex.nonprod.mozart.wiley.host"
*  issuer: C=US; O=Let's Encrypt; CN=R3
*  SSL certificate verify ok.
```

## dex-k8s-authenticator
Get k8s API cert
```
aws eks describe-cluster \
--name perf-test-nonprod \
--query 'cluster.certificateAuthority' \
--region us-east-1 --output text | base64 -d
```


Deploy
```
helm upgrade --install \
dex-k8s-authenticator dex-k8s-authenticator \
--namespace dex \
--values dex-k8s-authenticator.yaml \
--debug --dry-run
```
