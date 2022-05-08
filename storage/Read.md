# Storage providers

Different storage providers can be used as source, destination, and state. The state bucket is used for storing the state of the run.

## GCS

The format for GCS buckets is `gs://bucket_name`.
[Application default credentials](https://cloud.google.com/docs/authentication/production#automatically) are used for authentication. Using ADC means while running in cloud it would automatically pick the auth. When running locally you must [login as the user](https://cloud.google.com/sdk/gcloud/reference/auth/application-default/login) or set the env var [GOOGLE_APPLICATION_CREDENTIALS](https://cloud.google.com/docs/authentication/production#automatically) to the path of the service account key file.

## S3

The format for GCS buckets is `s3://bucket_name`.
For auth the [~/.s3/config file](https://cloud.google.com/docs/authentication/production#automatically) and the [~/.s3/credentials](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#creating-the-credentials-file) must be set.

## Local
The format for local filesystem buckets (folders) is `file://folder`.
