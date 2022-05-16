package nodes

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	mcfg "dgeorgievski/eks/pkg/config"
)

func filebase64OrPanic(path string) pulumi.StringPtrInput {
	if fileData, err := ioutil.ReadFile(path); err == nil {
		return pulumi.String(base64.StdEncoding.EncodeToString(fileData[:]))
	} else {
		panic(err.Error())
	}
}

func AsgLaunchTemplate(ctx *pulumi.Context,
	clusterName string,
	asgOneData *mcfg.AsgData,
	SSHKeyName pulumi.StringOutput,
	secGroup *ec2.SecurityGroup,
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

	oneLaunchTemplate, err := ec2.NewLaunchTemplate(ctx, asgOneData.Name, &ec2.LaunchTemplateArgs{
		NamePrefix:   pulumi.String(asgOneData.Name),
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

	ctx.Export("AsgOneLaunchTemplate", oneLaunchTemplate.Name)

	return oneLaunchTemplate, nil
}
