package nodes

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	mcfg "dgeorgievski/eks/pkg/config"
)

// KarpenterLaunchTemplate - a custom launch template for Karpenter provisioned nodes.
func KarpenterLaunchTemplate(ctx *pulumi.Context,
	clusterName string,
	asgOneData *mcfg.AsgData,
	SSHKeyName pulumi.StringOutput,
	secGroup *ec2.SecurityGroup,
	nodeGroupRole *iam.Role,
	globalTags map[string]string) (*ec2.LaunchTemplate, error) {

	nodeTags := make(map[string]string)
	for k, v := range globalTags {
		nodeTags[k] = v
	}
	// used by escalator node group auto-scaler.
	nodeTags["customer"] = "shared"
	nodeTags["Name"] = fmt.Sprintf("eks-%s", clusterName)
	nodeTags["karpenter.sh/discovery"] = clusterName

	fileAsset := pulumi.NewFileAsset("./files/bootstrap.sh")

	// Needed by Karpented to set the IAM profile on the EC2 nodes it provisions
	insProfName := fmt.Sprintf("KarpenterNodeInstanceProfile-%s", clusterName)
	instanceProfile, err := iam.NewInstanceProfile(ctx, insProfName, &iam.InstanceProfileArgs{
		Name: pulumi.String(insProfName),
		Role: nodeGroupRole.Name,
		Tags: pulumi.ToStringMap(globalTags),
	})

	if err != nil {
		return nil, err
	}

	ctx.Export("KarpenterNodeInstanceProfileArn", instanceProfile.Arn)

	kpLaunchTemplate, err := ec2.NewLaunchTemplate(ctx, "karpenter-eks-lt", &ec2.LaunchTemplateArgs{
		NamePrefix:   pulumi.String("karpenter-eks-lt"),
		KeyName:      SSHKeyName,
		ImageId:      pulumi.String(asgOneData.ImageId),
		InstanceType: pulumi.String(asgOneData.InstanceType),
		EbsOptimized: pulumi.String("true"),
		BlockDeviceMappings: ec2.LaunchTemplateBlockDeviceMappingArray{
			&ec2.LaunchTemplateBlockDeviceMappingArgs{
				DeviceName: pulumi.String("/dev/xvda"),
				Ebs: &ec2.LaunchTemplateBlockDeviceMappingEbsArgs{
					DeleteOnTermination: pulumi.String("true"),
					VolumeSize:          pulumi.Int(asgOneData.DiskSizeGB),
					VolumeType:          pulumi.String("gp3"),
				},
			},
		},
		// required for Targetted autoscaling policy
		Monitoring: ec2.LaunchTemplateMonitoringArgs{
			Enabled: pulumi.Bool(true),
		},

		IamInstanceProfile: &ec2.LaunchTemplateIamInstanceProfileArgs{
			Arn: instanceProfile.Arn,
		},

		UserData: filebase64OrPanic(fileAsset.Path()),

		UpdateDefaultVersion: pulumi.Bool(true),

		VpcSecurityGroupIds: pulumi.StringArray{
			secGroup.ID(),
		},

		TagSpecifications: ec2.LaunchTemplateTagSpecificationArray{
			&ec2.LaunchTemplateTagSpecificationArgs{
				ResourceType: pulumi.String("instance"),
				Tags:         pulumi.ToStringMap(nodeTags),
			},
		},
		Tags: pulumi.ToStringMap(nodeTags),
	})
	if err != nil {
		return nil, err
	}

	ctx.Export("KarpenterLaunchTemplate", kpLaunchTemplate.Name)

	return kpLaunchTemplate, nil
}
