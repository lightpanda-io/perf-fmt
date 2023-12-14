# perf-fmt

perf-fmt read, format and save performance results.
The program is written in [Go](https://go.dev/).

## Docker

Each version of perf-fmt is bundled in a docker image available on GH registry.

## AWS S3

The perf-fmt formats and stores json result on AWS S3 bucket.
Currently we use it with the `lpd-perf` S3 bucket.

A dedicated user has been created `perf-fmt`

and a policy `s3_lpd-perf`


### AWS Policy

The policy used with perf-fmt is `s3_lpd-perf`


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
