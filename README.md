
# ecr-auth-refresh

Allow Kubernetes to pull images from private ECR registry and periodically refresh ECR credentials before standard 12 hour timeout.

## How it works

Once `ecr-auth-refresh-credentials` have been been provisioned with scoped IAM user permissions to access ECR. Credentials are then pulled into the kubernetes secret type of `kubernetes.io/dockerconfigjson`. This is rotated every 3 hours.

## Example Kubernetes deployment

Create an IAM user with a policy giving ECR access and attach to the user.

Example policy allowing access to all ECR repositories in the account. This can be scoped down in `Resource`.

```
{
	"Version": "2012-10-17",
	"Statement": [{
		"Effect": "Allow",
		"Action": [
		    "ecr:GetDownloadUrlForLayer",
                    "ecr:BatchGetImage",
                    "ecr:BatchCheckLayerAvailability"
		],
		"Resource": "*"
	}]
}
```

Below environment variables outputs from the ECR user need to be populated in ./deployment/secrets-aws.yaml

| Variable                 | Explanation                                                |
|--------------------------|------------------------------------------------------------|
| AWS_ACCESS_KEY_ID        | ECR IAM users key ID                                       |
| AWS_SECRET_ACCESS_KEY    | ECR IAM users secret access key                            |
| ACCOUNT_ID               | AWS account ID where the ECR repository is provisioned     |
| AWS_DEFAULT_REGION       | AWS region where the ECR repository is provisioned         |

Example dployment files are in ./deployment/*

Deploying directly to your cluster.

``` kubectl apply -f ./deployment/ ```




