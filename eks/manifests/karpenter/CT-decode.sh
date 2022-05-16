






ref: https://aws.amazon.com/premiumsupport/knowledge-center/ec2-not-auth-launch/

alias dimi='aws sts decode-authorization-message --encoded-message'

{
    "allowed":false,
    "explicitDeny":false,
    "matchedStatements":
        {"items":[]},
    "failures":{"items":[]},
    "context":{
        "principal":{
            "id":"AROA4STNCXSNDFTED5B2H:1651608034126661953",
            "arn":"arn:aws:sts::864590937242:assumed-role/karpenter-controller-perf-test-nonprod/1651608034126661953"},
            "action":"ec2:RunInstances",
            "resource":"arn:aws:ec2:us-east-1:864590937242:key-pair/perf-test-ng-ssh",
            "conditions":{
                "items":[
                    {
                        "key":"aws:Region",
                        "values":{"items":[{"value":"us-east-1"}]}
                    },
                    {
                        "key":"aws:Service",
                        "values":{"items":[{"value":"ec2"}]}
                    },
                    {
                        "key":"aws:Resource",
                        "values":{"items":[{"value":"key-pair/perf-test-ng-ssh"}]}
                    },
                    {
                        "key":"aws:Type",
                        "values":{"items":[{"value":"key-pair"}]}
                    },
                    {
                        "key":"aws:Account",
                        "values":{"items":[{"value":"864590937242"}]}
                    },
                    {
                        "key":"ec2:KeyPairType",
                        "values":{"items":[{"value":"rsa"}]}
                    },
                    {
                        "key":"ec2:Region",
                        "values":{"items":[{"value":"us-east-1"}]}
                    },
                    {
                        "key":"aws:ARN",
                        "values":{"items":[{"value":"arn:aws:ec2:us-east-1:864590937242:key-pair/perf-test-ng-ssh"}]}
                    },
                    {
                        "key":"ec2:LaunchTemplate",
                        "values":{"items":[{"value":"arn:aws:ec2:us-east-1:864590937242:launch-template/lt-0ef453aacac257506"}]}
                    },
                    {
                        "key":"ec2:IsLaunchTemplateResource",
                        "values":{"items":[{"value":"true"}]}
                    },
                    {
                        "key":"ec2:KeyPairName",
                        "values":{"items":[{"value":"perf-test-ng-ssh"}]}}]}}}
