Karpenter Auto Scaler 
-----------------------

Ref: https://karpenter.sh/v0.9.0/getting-started/
https://aws.amazon.com/premiumsupport/knowledge-center/eks-troubleshoot-oidc-and-irsa/?nc1=h_ls
      

# Helm deployment

Ref: https://github.com/aws/karpenter/tree/main/charts/karpenter

```shell
$ cd manifests/karpenter

$ helm upgrade --install \
karpenter karpenter/karpenter \
--namespace karpenter --create-namespace \
--version 0.9.1 \
--wait \
-f values.yaml \
--debug --dry-run

```
KarpenterNodeInstanceProfile-perf-test-nonprod

helm upgrade --install \
karpenter karpenter/karpenter \
--namespace karpenter --create-namespace \
--version 0.1.1 \
--wait \
--debug --dry-run