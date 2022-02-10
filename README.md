# Kromium

## What is Kromium?

Kromium is a bulk file copy/transformation tool. The pipeline is linear chain of transformations and is expressed using simple JSON config. Kromium is designed for simplicity and ease of use. A simple configuration example is the following:

```
{
 "SourceBucket": "gs://kromium-src",
 "DestinationBucket": "gs://kromium-dst",
 "NameSuffix": ".gz",
 "Transforms": [
   {
     "name": "GzipCompress",
     "args": {
      "level": 10
     }
   }
 ]
}
```

This configuration will simply read all objects from the `kromium-src` bucket, apply the gzip compression transform and write the output to the `kromium-dst` bucket. The optional `NameSuffix` argument specifies if a suffix should be applied to the object names when writing to the destination bucket, this can be used for adding filename extensions.

## Features
- Resumeable. Kromium checkpoints progress. So in case of any crashes it can be simply restarted.
- Stateless. Multiple parallel runs of the Kromium pipeline can be executed independantly to achieve large parallelism. It only relies on the checkpoint state to avoid duplicate work.
- Resource optimized. Kromium employs storage source/sink optimizations to optimize the overall resource usage for the job.
- Transformation. Comes with a few common transformations, and is very easy to extend. 

## Use cases
- Simple ETL pipelines, e.g. Load CSVs from a bucket to SQL.
- Copying large amounts of data, e.g. copying large amounts data from one bucket/SQL table to another destination.

## Storage providers
As of now, Kromium only supports GCS and Local filesystem for storage. The support for S3 (and Azure) will be added soon. The source bucket is a uri which should be fully qualified. Following are the prefixes for supported storage solution:
```
GCS: gs://bucket
Local filesystem: file://folderpath
S3: s3://bucket
```

## Supported transforms
**Generic file transforms**
```
- Identity: does not change the file content.
- GzipCompress/GzipDecompress: The arguments for compression level.
- Encrypt: The arguments are the compression algorithms and key. Default is simple openssl encryption to the file.
- Decrypt: Same as encryption but decrypts the file instead.
```

**CSV transforms**
```
- CSVFilter: selects particular columns from CSV. The first row is considered the name of columns, and the arg for this transform is simply the name of columns.
```

**Image transforms**
```
- Resize: takes the new resolution as arguements.
```
## Installation
### From source
go test
go install
$GO_BIN/kromium --run /tmp/identity_local.json (e.g. ~/go/bin/kromium --run /tmp/identity_local.json)

### Docker
docker build -t kromium .
Example run:
docker run -v /tmp/src:/tmp/src -v /tmp/dst:/tmp/dst -v /tmp/state:/tmp/state -v /Users/sharva/Workspace/kromium_sync/examples/identity_local.json:/tmp/identity_local.json kromium --run /tmp/identity_local.json

## Future work
- By default the transformation runs on the local machine. Support for Kubernetes will be added soon.
