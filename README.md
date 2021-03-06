# Kromium
**Kromium([https://kromium.io](https://kromium.io)) is under https://www.apache.org/licenses/LICENSE-2.0**

## What is Kromium?

Kromium is an efficient no-code bulk file copy/transformation tool. The pipeline is a linear chain of transformations and is expressed using simple CUE[https://cuelang.org/] based configs. Each transform is a stateless function and every run of the pipeline is idempotent. Kromium is designed for simplicity and ease of use. A simple configuration example is the following:

```
{
 SourceBucket: "gs://kromium-src",
 DestinationBucket: "gs://kromium-dst",
 StateBucket: "gs://kromium-state",
 NameSuffix: ".gz",
 Transforms: [
   {
     Type: "GzipCompress",
     Args: {
       level: 4
     }
   }
 ]
}
```

This configuration will simply read all objects from the `kromium-src` bucket, apply the gzip compression transform and write the output to the `kromium-dst` bucket. The checkpointing state will be written to `kromium-state`. The optional `NameSuffix` argument specifies if a suffix should be applied to the object names when writing to the destination bucket, this can be used for adding filename extensions. The state bucket is used for checkpointing and tracking other types of state information. More examples can be found in https://github.com/sharvanath/kromium/tree/main/examples.

## Features
- Resumeable. Kromium checkpoints progress in the state bucket. So in case of any crashes it can be simply restarted.
- Efficient. Kromium uses efficient go concurrency constructs to run fast and in parallel. It can easily process up to 100 Google cloud storage objects/second on a simple macbook pro (8-Core Intel i9). Local files processing can be much faster.
- Parallelizable without synchronization. Multiple parallel runs of the Kromium pipeline can be executed independantly to achieve large parallelism. It only relies on the checkpoint state to avoid duplicate work. 
- Transformations. Comes with a few common transformations, and it is very easy to a add new one.
- High level details on checkpointing/state manegment can be found in https://github.com/sharvanath/kromium/blob/main/core/README.md.

## Use cases
- Copying large amounts of data, e.g. copying large amounts data from one bucket/SQL table to another destination.
- ETL workloads, Reading file data in bulk, Transforming it and writing back to some other location.

## Storage providers
Different storage providers can be used as source, destination, and state. The state bucket is used for storing the state of the run.
As of now, Kromium only supports GCS and Local filesystem for storage. The support for SQL and Azure will be added soon. The source bucket is a uri which should be fully qualified. Following are the prefixes for supported storage solution:
```
Local filesystem: file://folderpath
GCS: gs://bucket
s3: s3://bucket
```

More details on how to configure auth for storage provider https://github.com/sharvanath/kromium/tree/main/storage.

## Supported transforms
**Generic file transforms**
```
- Identity: does not change the file content.
- GzipCompress/GzipDecompress: The arguments for compression level.
- Encrypt: The arguments are the compression algorithms and key. Default is simple openssl encryption to the file.
- Decrypt: Same as encryption but decrypts the file instead.
- Sed: Use sed commands for modifying text.
```

## Execute from source
go run main.go --run examples/identity_local.cue 

## Release binary
After downloading the latest release binary for your platform from [https://github.com/sharvanath/kromium/releases](https://github.com/sharvanath/kromium/releases).
Simply run

./kromium --run pipeline.cue

## Troubleshooting
* If you see Mac blocking the binary since it's untrusted. Follow [this](https://github.molgen.mpg.de/pages/bs/macOSnotes/mac/mac_procs_unsigned.html)
* If you see errors of the form "too many open files" or "could not resolve address", check `ulimit -n` and if it is less than 1024 set it to a higher value `ulimit -n 1024`.

## Development
### Testing
* `go test ./... -test.short` for unit tests
* `go test ./core -run RunIdentityPipelineLargeSequential` for large sequential copy test.
* `go test ./core -run RunIdentityPipelineLargeParallel` for large parallel copy test.
* `go test ./storage  -tags=integration` for storage integration tests. Make sure to update the auth config and to update the test to use the right bucket names.

### Profiling
* go run main.go --run examples/identity_local.cue 
* http://localhost:6060/debug/pprof/trace?seconds=120
* go tool trace <file_name>

### Docker
docker build -t kromium .

Example run:
docker run -v /tmp/src:/tmp/src -v /tmp/dst:/tmp/dst -v /tmp/state:/tmp/state -v /Users/sharva/Workspace/kromium_sync/examples/identity_local.cue:/tmp/identity_local.cue kromium --run /tmp/identity_local.cue

## Future work
- Add SQL/CSV transforms to support simple ETL pipelines, e.g. Load CSVs from a bucket to SQL.
- Storage optimized for very large processing rates. Kromium should employ storage source/sink optimizations to optimize the overall resource usage for the job. GCS (https://cloud.google.com/storage/docs/request-rate)
- By default the transformation runs on the local machine. Support for Kubernetes will be added soon.
