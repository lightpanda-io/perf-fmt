# perf-fmt

perf-fmt read, format and save performance results.
The program is written in [Go](https://go.dev/).

## Docker

Each version of perf-fmt is bundled in a docker image available on GH registry.

## AWS S3

The perf-fmt formats and stores json result on AWS S3 bucket.
Currently we use it with the `lpd-perf` S3 bucket.

A dedicated user has been created `perf-fmt`
[1](https://us-east-1.console.aws.amazon.com/iam/home?region=eu-central-1#/users/details/perf-fmt?section=permissions)
and a policy `s3_lpd-perf`
[2](https://us-east-1.console.aws.amazon.com/iam/home?region=eu-central-1#/policies/details/arn%3Aaws%3Aiam%3A%3A095723607651%3Apolicy%2Fs3_lpd-perf?section=permissions)

### AWS Policy

The policy used with perf-fmt is `s3_lpd-perf`
[2](https://us-east-1.console.aws.amazon.com/iam/home?region=eu-central-1#/policies/details/arn%3Aaws%3Aiam%3A%3A095723607651%3Apolicy%2Fs3_lpd-perf?section=permissions)

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:PutObjectAcl"
            ],
            "Resource": "arn:aws:s3:::lpd-perf/*"
        },
        {
            "Sid": "VisualEditor1",
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket"
            ],
            "Resource": "arn:aws:s3:::lpd-perf"
        }
    ]
}
```
