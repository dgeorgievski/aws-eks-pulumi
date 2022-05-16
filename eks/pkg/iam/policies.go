package iam

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	mutils "dgeorgievski/eks/pkg/utils"
)

// IamPoliciesExternalDNS - Set IAM policy for external-dns
func IamPoliciesExternalDNS(ctx *pulumi.Context,
	pubZoneArns []string,
	oidcProvider *iam.OpenIdConnectProvider,
	tagsGlobal map[string]string) error {

	// fmt.Printf("IamPoliciesExternalDNS pubZoneIDs %v\n", pubZoneArns)

	cmRole, err := iam.NewRole(ctx, "eks-external-dns", &iam.RoleArgs{
		Name:             pulumi.String("eks-external-dns"),
		Tags:             pulumi.ToStringMap(tagsGlobal),
		AssumeRolePolicy: genAssumePolicy(oidcProvider, "external-dns", "external-dns"),
	})

	if err != nil {
		return err
	}

	ctx.Export("IAMRoleExternalDNS", cmRole.Arn)

	tmpJSON0, err := json.Marshal(map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Action": []string{
					"route53:ChangeResourceRecordSets",
				},
				"Effect":   "Allow",
				"Resource": pubZoneArns,
			},
			{
				"Effect": "Allow",
				"Action": []string{
					"route53:ListHostedZones",
					"route53:ListResourceRecordSets",
				},
				"Resource": []string{"*"},
			},
		},
	})
	if err != nil {
		return err
	}

	json0 := string(tmpJSON0)
	cmPolicy, err := iam.NewPolicy(ctx, "eks-external-dns", &iam.PolicyArgs{
		Description: pulumi.String("External-dns policy"),
		Policy:      pulumi.String(json0),
	})
	if err != nil {
		return err
	}

	_, err = iam.NewRolePolicyAttachment(ctx, "eks-external-dns", &iam.RolePolicyAttachmentArgs{
		PolicyArn: cmPolicy.Arn,
		Role:      cmRole.Name,
	})
	if err != nil {
		return err
	}

	return nil
}

// IamPoliciesCertManager - Set IAM policy for cert-manager
func IamPoliciesCertManager(ctx *pulumi.Context,
	pubZoneArns []string,
	oidcProvider *iam.OpenIdConnectProvider,
	tagsGlobal map[string]string) error {

	cmRole, err := iam.NewRole(ctx, "eks-iam-cm", &iam.RoleArgs{
		Name:             pulumi.String("eks-iam-cm"),
		Tags:             pulumi.ToStringMap(tagsGlobal),
		AssumeRolePolicy: genAssumePolicy(oidcProvider, "cert-manager", "cert-manager"),
	})

	if err != nil {
		return err
	}

	ctx.Export("IAMRoleCertManager", cmRole.Arn)

	tmpJSON0, err := json.Marshal(map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Action": []string{
					"route53:GetChange",
				},
				"Effect":   "Allow",
				"Resource": []string{"arn:aws:route53:::change/*"},
			},
			{
				"Effect": "Allow",
				"Action": []string{
					"route53:ListHostedZonesByName",
				},
				"Resource": []string{"*"},
			},
			{
				"Effect": "Allow",
				"Action": []string{
					"route53:ChangeResourceRecordSets",
					"route53:ListResourceRecordSets",
				},
				"Resource": pubZoneArns,
			},
		},
	})
	if err != nil {
		return err
	}

	json0 := string(tmpJSON0)
	cmPolicy, err := iam.NewPolicy(ctx, "eks-iam-cm", &iam.PolicyArgs{
		Description: pulumi.String("Cert Manager DNS policy"),
		Policy:      pulumi.String(json0),
	})
	if err != nil {
		return err
	}

	_, err = iam.NewRolePolicyAttachment(ctx, "eks-iam-cm", &iam.RolePolicyAttachmentArgs{
		PolicyArn: cmPolicy.Arn,
		Role:      cmRole.Name,
	})
	if err != nil {
		return err
	}

	return nil
}

// IamPoliciesEscalator - Set IAM policy for AWS Escalator
func IamPoliciesEscalator(ctx *pulumi.Context,
	oidcProvider *iam.OpenIdConnectProvider,
	tagsGlobal map[string]string) error {

	cmRole, err := iam.NewRole(ctx, "eks-escalator", &iam.RoleArgs{
		Name:             pulumi.String("eks-escalator"),
		Tags:             pulumi.ToStringMap(tagsGlobal),
		AssumeRolePolicy: genAssumePolicy(oidcProvider, "kube-system", "escalator"),
	})

	if err != nil {
		return err
	}

	ctx.Export("IAMRoleEscalator", cmRole.Arn)

	tmpJSON0, err := json.Marshal(map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Action": []string{
					"autoscaling:AttachInstances",
					"autoscaling:CreateOrUpdateTags",
					"autoscaling:DescribeAutoScalingGroups",
					"autoscaling:SetDesiredCapacity",
					"autoscaling:TerminateInstanceInAutoScalingGroup",
					"ec2:CreateFleet",
					"ec2:CreateTags",
					"ec2:DescribeInstances",
					"ec2:DescribeInstanceStatus",
					"ec2:RunInstances",
					"ec2:TerminateInstances",
					"iam:PassRole",
				},
				"Effect":   "Allow",
				"Resource": []string{"*"},
			},
		},
	})
	if err != nil {
		return err
	}

	json0 := string(tmpJSON0)
	cmPolicy, err := iam.NewPolicy(ctx, "eks-escalator", &iam.PolicyArgs{
		Description: pulumi.String("AWS Escalator policy"),
		Policy:      pulumi.String(json0),
	})
	if err != nil {
		return err
	}

	_, err = iam.NewRolePolicyAttachment(ctx, "eks-escalator", &iam.RolePolicyAttachmentArgs{
		PolicyArn: cmPolicy.Arn,
		Role:      cmRole.Name,
	})
	if err != nil {
		return err
	}

	return nil
}

// IamPoliciesKarpenterAS - Set IAM policy for Karpenter AutoScaling service.
// https://github.com/terraform-aws-modules/terraform-aws-iam/blob/master/modules/iam-role-for-service-accounts-eks/policies.tf#L510
func IamPoliciesKarpenterAS(ctx *pulumi.Context,
	eksClusterName string,
	awsAccoountId string,
	eksCluster *eks.Cluster,
	nodeGroup *eks.NodeGroup,
	oidcProvider *iam.OpenIdConnectProvider,
	tagsGlobal map[string]string) error {

	resourceName := fmt.Sprintf("karpenter-controller-%s", eksClusterName)

	cmRole, err := iam.NewRole(ctx, resourceName, &iam.RoleArgs{
		Name:             pulumi.String(resourceName),
		Tags:             pulumi.ToStringMap(tagsGlobal),
		AssumeRolePolicy: genAssumePolicy(oidcProvider, "karpenter", "karpenter"),
	})

	if err != nil {
		return err
	}

	ctx.Export("IAMRoleKarpenterController", cmRole.Arn)

	nodeIamPolicy := iamNodePoliciesKarpenter(awsAccoountId, eksCluster, nodeGroup)

	cmPolicy, err := iam.NewPolicy(ctx, resourceName, &iam.PolicyArgs{
		Name:        pulumi.String(resourceName),
		Description: pulumi.String("Karpenter Controller IAM policy"),
		Policy:      nodeIamPolicy,
	})
	if err != nil {
		return err
	}

	_, err = iam.NewRolePolicyAttachment(ctx, resourceName, &iam.RolePolicyAttachmentArgs{
		PolicyArn: cmPolicy.Arn,
		Role:      cmRole.Name,
	})
	if err != nil {
		return err
	}

	return nil
}

// IamPoliciesKarpenterAS - Set IAM policy for Karpenter AutoScaling service.
// https://github.com/terraform-aws-modules/terraform-aws-iam/blob/master/modules/iam-role-for-service-accounts-eks/policies.tf#L510
// Read the policy from a file, use pulumi Apply All future to collect the the required info of requird resources.
func iamNodePoliciesKarpenter(awsAccoountId string,
	eksCluster *eks.Cluster,
	nodeGroup *eks.NodeGroup) pulumi.Output {

	karpenterIamPolicy := pulumi.All(eksCluster.ID().ToStringOutput(), awsAccoountId, nodeGroup.NodeRoleArn).ApplyT(
		func(args []interface{}) (string, error) {
			clusterId := args[0].(string)
			accountId := args[1].(string)
			nodeGroupRoleArn := args[2].(string)

			karpenterIamPolicyAsset := pulumi.NewFileAsset("./files/karpenter-iam.json")
			karpenterIamPolicyTmp := mutils.ReadFileAsset(karpenterIamPolicyAsset.Path())

			re := strings.NewReplacer("{EksClusterId}", clusterId,
				"{AwsAccountId}", accountId,
				"{NodeGroupRoleArn}", nodeGroupRoleArn)

			retIamPolicy := re.Replace(karpenterIamPolicyTmp)

			return retIamPolicy, nil
		},
	)

	return karpenterIamPolicy
}

// genAssumePolicy - Generate trust policy for a given k8s service account
func genAssumePolicy(oidcProvider *iam.OpenIdConnectProvider,
	k8sNamespace string,
	k8sServiceAccount string) pulumi.StringOutput {

	return pulumi.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {
						"Federated": "%s"
					},
					"Action": "sts:AssumeRoleWithWebIdentity",
					"Condition": {
						"StringEquals": {
							"%s:sub": "system:serviceaccount:%s:%s"
						}
					}
				}
			]
	}`, oidcProvider.Arn, oidcProvider.Url, k8sNamespace, k8sServiceAccount)
}
