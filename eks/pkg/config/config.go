package config

type Config struct {
	ClusterName             string
	Version                 string
	InstanceType            string
	HostedZonesPublic       []string
	HostedZonesPrivate      []string
	EksOidcRootCAThumbprint string
	ASG                     ASG
	Tags                    Tags
}

type ASG struct {
	One AsgData
}

type SSHKeys struct {
	Name string
	Priv string
	Pub  string
}

type AsgData struct {
	Name         string
	DiskSizeGB   int16
	InstanceType string
	ImageId      string
	CapacityType string
	MinSize      int8
	MaxSize      int8
	DesiredSize  int8
	SSHKeys      SSHKeys
}

type Tags struct {
	Global GlobalTags
}

type GlobalTags struct {
	SysCode        string
	SysID          string
	InfraOwner     string
	FunctionalArea string
	EnvName        string
	ManagedBy      string
}
