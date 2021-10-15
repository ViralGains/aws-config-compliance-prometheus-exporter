# aws-config-compliance-prometheus-exporter
Prometheus Exporter for AWS Config Compliance
(forked from chaspy - thanks!)
## How to run

### Local

```
$ go run main.go
```


### Environment Variables needed
AWS_CONFIG_SCRAPE_PREFIX should be set and is used to determine which rules to alert on.  

This allows us to alert based on environment.

AWS_API_INTERVAL specifies how often to scrape.  Defaults to 60 seconds if you do not set it.


AWS_REGION you will need to set this if you are running in an environment that uses IAM role instead of an aws config.  
### Binary

Get the binary file from [Releases](https://github.com/ViralGains/aws-config-compliance-prometheus-exporter/releases) and run it.

## Metrics

```
$ curl -s localhost:8083/metrics | grep aws_custom_config_compliance
# HELP aws_custom_config_compliance Number of compliance
# TYPE aws_custom_config_compliance gauge
aws_custom_config_compliance{cap_exceeded="false",compliance="COMPLIANT",config_rule_name="securityhub-efs-encrypted-check-bd414301"} 0
aws_custom_config_compliance{cap_exceeded="false",compliance="INSUFFICIENT_DATA",config_rule_name="securityhub-dms-replication-not-public-1f6729b8"} 0
aws_custom_config_compliance{cap_exceeded="false",compliance="INSUFFICIENT_DATA",config_rule_name="securityhub-ec2-managedinstance-patch-compliance-440fg71a"} 0
aws_custom_config_compliance{cap_exceeded="false",compliance="NON_COMPLIANT",config_rule_name="eip-attached"} 2
aws_custom_config_compliance{cap_exceeded="false",compliance="NON_COMPLIANT",config_rule_name="s3-bukcet-logging-enabled"} 23
```

## IAM Role

The following policy must be attached to the AWS role to be executed.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "config:DescribeComplianceByConfigRule",
            ],
            "Resource": "*"
        }
    ]
}
```

