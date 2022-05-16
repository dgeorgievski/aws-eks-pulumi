
EKS provisioning and management in AWS
-------------------------------------------

# Reference
1. Pulumi CLI: https://www.pulumi.com/docs/reference/cli/pulumi_new/
2. Pulimi State and Backends: https://www.pulumi.com/docs/intro/concepts/state/

# Creating aa new Project
1. set AWS env

2. pulumi login s3://my-state-bucket/pulumi/eks

3. pulumi new aws-go \
 --secrets-provider="awskms://alias/pulumi-eks-dev?region=us-east-1"
 ```
 project name: (eks) perf-test
project description: (A minimal AWS Go Pulumi program) Mozart's Distributed Perf Tests Platform
Created project 'perf-test'

stack name: (dev) nonprod
Created stack 'nonprod'

aws:region: The AWS region to deploy into: (us-east-1)
Saved config

Installing dependencies...

go: downloading github.com/pulumi/pulumi/sdk/v3 v3.28.0
go: downloading github.com/pulumi/pulumi-aws/sdk/v5 v5.1.0
go: downloading pgregory.net/rapid v0.4.7
go: downloading github.com/spf13/cobra v1.4.0
go: downloading golang.org/x/sys v0.0.0-20210817190340-bfb29a6856f2
go: downloading github.com/rogpeppe/go-internal v1.8.1
Finished installing dependencies

Your new project is ready to go! ✨
```
