package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/route53"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	pcfg "github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"

	mcfg "dgeorgievski/eks/pkg/config"
	mec2 "dgeorgievski/eks/pkg/ec2"
	miam "dgeorgievski/eks/pkg/iam"
)

func main() {

	pulumi.Run(func(ctx *pulumi.Context) error {

		// Read stack config
		var config mcfg.Config
		cfg := pcfg.New(ctx, "")
		cfg.RequireObject("config", &config)

		var globalTags map[string]string
		data, _ := json.Marshal(config.Tags.Global)
		json.Unmarshal(data, &globalTags)
		// globalTags = lowerCaseTags(globalTags)

		//===== AWS Lookups======
		currentPartition, err := aws.GetPartition(ctx, nil, nil)
		if err != nil {
			return err
		}
		stsPrincipal := fmt.Sprintf("sts.%s", currentPartition.DnsSuffix)

		current, err := aws.GetCallerIdentity(ctx, nil, nil)
		if err != nil {
			return err
		}
		ctx.Export("accountId", pulumi.String(current.AccountId))

		// Get VPC details
		vpc, err := ec2.LookupVpc(ctx, &ec2.LookupVpcArgs{
			Tags: map[string]string{"Name": "eks-shared-nonprod"},
		})
		if err != nil {
			return err
		}

		subnetIds, err := ec2.GetSubnets(ctx, &ec2.GetSubnetsArgs{
			Filters: []ec2.GetSubnetsFilter{
				{
					Name:   "vpc-id",
					Values: []string{vpc.Id},
				},
				{
					Name:   "mapPublicIpOnLaunch",
					Values: []string{"false"},
				},
			},
		})
		if err != nil {
			return err
		}

		// Create a KeyPair
		nodeKeyPair, err := ec2.NewKeyPair(ctx, "asgOneKeyPair", &ec2.KeyPairArgs{
			KeyName:   pulumi.String(config.ASG.One.SSHKeys.Name),
			PublicKey: pulumi.String(config.ASG.One.SSHKeys.Pub),
		})
		if err != nil {
			return err
		}

		// Route 53 zones
		// TODO: add private zone if required in future
		var pubZoneArns []string
		for _, zone := range config.HostedZonesPublic {
			selected, err := route53.LookupZone(ctx, &route53.LookupZoneArgs{
				Name:        pulumi.StringRef(zone),
				PrivateZone: pulumi.BoolRef(false),
			}, nil)
			if err != nil {
				return err
			}

			pubZoneArns = append(pubZoneArns, selected.Arn)
		}

		//===== Provisioning =====
		eksRole, err := iam.NewRole(ctx, "eks-iam-eksRole", &iam.RoleArgs{

			Tags: pulumi.ToStringMap(globalTags),
			AssumeRolePolicy: pulumi.String(`{
		    "Version": "2008-10-17",
		    "Statement": [{
		        "Effect": "Allow",
		        "Principal": {
		            "Service": "eks.amazonaws.com"
		        },
		        "Action": "sts:AssumeRole"
		    }]
		}`),
		})
		if err != nil {
			return err
		}
		eksPolicies := []string{
			"arn:aws:iam::aws:policy/AmazonEKSServicePolicy",
			"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
		}
		for i, eksPolicy := range eksPolicies {
			_, err := iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("rpa-%d", i), &iam.RolePolicyAttachmentArgs{
				PolicyArn: pulumi.String(eksPolicy),
				Role:      eksRole.Name,
			})
			if err != nil {
				return err
			}
		}

		// Create the EC2 NodeGroup Role
		nodeGroupRole, err := iam.NewRole(ctx, "eks-iam-node-role", &iam.RoleArgs{
			Tags: pulumi.ToStringMap(globalTags),
			AssumeRolePolicy: pulumi.String(`{
		    "Version": "2012-10-17",
		    "Statement": [{
		        "Effect": "Allow",
		        "Principal": {
		            "Service": "ec2.amazonaws.com"
		        },
		        "Action": "sts:AssumeRole"
		    }]
		}`),
		})
		if err != nil {
			return err
		}

		nodeGroupPolicies := []string{
			"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
			"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
			"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
			"arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore",
		}

		for i, nodeGroupPolicy := range nodeGroupPolicies {
			_, err := iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("ngpa-%d", i), &iam.RolePolicyAttachmentArgs{
				Role:      nodeGroupRole.Name,
				PolicyArn: pulumi.String(nodeGroupPolicy),
			})
			if err != nil {
				return err
			}
		}

		nodeTags := make(map[string]string)
		for k, v := range globalTags {
			nodeTags[k] = v
		}

		nodeTags["karpenter.sh/discovery"] = config.ClusterName

		clusterSg, err := ec2.NewSecurityGroup(ctx, "cluster-sg", &ec2.SecurityGroupArgs{
			VpcId: pulumi.String(vpc.Id),
			Egress: ec2.SecurityGroupEgressArray{
				ec2.SecurityGroupEgressArgs{
					Protocol:   pulumi.String("-1"),
					FromPort:   pulumi.Int(0),
					ToPort:     pulumi.Int(0),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
			Ingress: ec2.SecurityGroupIngressArray{
				ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("-1"),
					FromPort:   pulumi.Int(0),
					ToPort:     pulumi.Int(0),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
				ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(22),
					ToPort:     pulumi.Int(22),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
			Tags: pulumi.ToStringMap(nodeTags),
		})
		if err != nil {
			return err
		}
		// Create EKS Cluster
		eksCluster, err := eks.NewCluster(ctx, config.ClusterName, &eks.ClusterArgs{
			Name:    pulumi.String(config.ClusterName),
			Version: pulumi.String(config.Version),
			RoleArn: pulumi.StringInput(eksRole.Arn),
			VpcConfig: &eks.ClusterVpcConfigArgs{
				EndpointPublicAccess:  pulumi.Bool(false),
				EndpointPrivateAccess: pulumi.Bool(true),
				PublicAccessCidrs: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
				SecurityGroupIds: pulumi.StringArray{
					clusterSg.ID().ToStringOutput(),
				},
				SubnetIds: toPulumiStringArray(subnetIds.Ids),
			},
			KubernetesNetworkConfig: eks.ClusterKubernetesNetworkConfigArgs{
				IpFamily:        pulumi.String("ipv4"),
				ServiceIpv4Cidr: pulumi.String("10.100.0.0/16"),
			},
			EnabledClusterLogTypes: pulumi.ToStringArray([]string{"api",
				"audit",
				"authenticator",
				"controllerManager",
				"scheduler"}),
			Tags: pulumi.ToStringMap(globalTags),
		})
		if err != nil {
			return err
		}

		nodeTags = make(map[string]string)
		for k, v := range globalTags {
			nodeTags[k] = v
		}
		// used by escalator node group auto-scaler.
		nodeTags["customer"] = "shared"
		nodeTags["Name"] = "eks-ng"

		launchTemplate, err := mec2.AsgLaunchTemplate(ctx,
			config.ClusterName,
			&config.ASG.One,
			nodeKeyPair.KeyName,
			clusterSg,
			globalTags)
		if err != nil {
			return err
		}

		ngOne, err := eks.NewNodeGroup(ctx, config.ASG.One.Name, &eks.NodeGroupArgs{
			ClusterName:        eksCluster.Name,
			NodeGroupName:      pulumi.String(config.ASG.One.Name),
			NodeRoleArn:        nodeGroupRole.Arn,
			SubnetIds:          toPulumiStringArray(subnetIds.Ids),
			ForceUpdateVersion: pulumi.Bool(true),
			CapacityType:       pulumi.String(config.ASG.One.CapacityType),
			LaunchTemplate: eks.NodeGroupLaunchTemplateArgs{
				Id:      launchTemplate.ID(),
				Version: pulumi.Sprintf("%d", launchTemplate.LatestVersion),
				// Version: launchTemplate pulumi.String(fmt.Sprintf("%v%v", "$", "Latest")),
			},
			ScalingConfig: &eks.NodeGroupScalingConfigArgs{
				DesiredSize: pulumi.Int(config.ASG.One.DesiredSize),
				MaxSize:     pulumi.Int(config.ASG.One.MaxSize),
				MinSize:     pulumi.Int(config.ASG.One.MinSize),
			},
			Tags: pulumi.ToStringMap(nodeTags),
		})
		if err != nil {
			return err
		}

		ctx.Export("nodeGroupOne", ngOne.Arn)

		// Export the cluster's kubeconfig.
		ctx.Export("kubeconfig", generateKubeconfig(eksCluster.Endpoint,
			eksCluster.CertificateAuthority.Data().Elem().ToStringOutput(),
			eksCluster.Name))

		oidcIdentity := eksCluster.Identities.ApplyT(func(idsOidc []eks.ClusterIdentity) string {
			return *idsOidc[0].Oidcs[0].Issuer
		}).(pulumi.StringOutput)

		oidcProvider, err := iam.NewOpenIdConnectProvider(ctx, "eks-nonprod", &iam.OpenIdConnectProviderArgs{
			ClientIdLists: pulumi.StringArray{
				pulumi.String(stsPrincipal),
			},
			ThumbprintLists: pulumi.StringArray{
				pulumi.String(config.EksOidcRootCAThumbprint),
			},
			Url:  oidcIdentity,
			Tags: pulumi.ToStringMap(globalTags),
		})
		if err != nil {
			return err
		}

		// create IAM external DNS roles and policies
		miam.IamPoliciesExternalDNS(ctx,
			pubZoneArns,
			oidcProvider,
			globalTags)

		miam.IamPoliciesCertManager(ctx,
			pubZoneArns,
			oidcProvider,
			globalTags)

		miam.IamPoliciesEscalator(ctx,
			oidcProvider,
			globalTags)

		miam.IamPoliciesKarpenterAS(ctx,
			config.ClusterName,
			current.AccountId,
			eksCluster,
			ngOne,
			oidcProvider,
			globalTags)

		mec2.KarpenterLaunchTemplate(ctx,
			config.ClusterName,
			&config.ASG.One,
			nodeKeyPair.KeyName,
			clusterSg,
			nodeGroupRole,
			globalTags)

		return nil
	})
}

//Create the KubeConfig Structure as per https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
func generateKubeconfig(clusterEndpoint pulumi.StringOutput,
	certData pulumi.StringOutput,
	clusterName pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.Sprintf(`{
        "apiVersion": "v1",
        "clusters": [{
            "cluster": {
                "server": "%s",
                "certificate-authority-data": "%v"
            },
            "name": "kubernetes",
        }],
        "contexts": [{
            "context": {
                "cluster": "kubernetes",
                "user": "aws",
            },
            "name": "aws",
        }],
        "current-context": "aws",
        "kind": "Config",
        "users": [{
            "name": "aws",
            "user": {
                "exec": {
                    "apiVersion": "client.authentication.k8s.io/v1alpha1",
                    "command": "aws-iam-authenticator",
                    "args": [
                        "token",
                        "-i",
                        "%s",
                    ],
                },
            },
        }],
    }`, clusterEndpoint, certData, clusterName)
}

func toPulumiStringArray(a []string) pulumi.StringArrayInput {
	var res []pulumi.StringInput
	for _, s := range a {
		res = append(res, pulumi.String(s))
	}
	return pulumi.StringArray(res)
}
