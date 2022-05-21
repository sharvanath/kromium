# Kromium
**Kromium is currently in development and is in pre-release. The licensing is also not final but current version is under [Apache license](https://www.apache.org/licenses/LICENSE-2.0)**

## What is Kromium?

Kromium is an efficient no-code bulk file copy/transformation tool. The pipeline is a linear chain of transformations and is expressed using simple [CUE](https://cuelang.org/) based configs. Each transform is a stateless function and every run of the pipeline is idempotent. Kromium is designed for simplicity and ease of use. A simple configuration example is the following:

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

This configuration will simply read all objects from the `kromium-src` bucket, apply the gzip compression transform and write the output to the `kromium-dst` bucket. The checkpointing state will be written to `kromium-state`. The optional `NameSuffix` argument specifies if a suffix should be applied to the object names when writing to the destination bucket, this can be used for adding filename extensions. The state bucket is used for checkpointing and tracking other types of state information. More examples can be found in [here](https://github.com/sharvanath/kromium/tree/main/examples).

## Features
- **Resumeable**: Kromium checkpoints progress in the state bucket. So in case of any crashes it can be simply restarted.
- **Efficient**: Kromium uses efficient go concurrency constructs to run fast and in parallel. It can easily process up to 100 Google cloud storage objects/second on a simple macbook pro (8-Core Intel i9). Local files processing can be much faster.
- **Parallelizable**: Multiple parallel runs of the Kromium pipeline can be executed independantly to achieve large parallelism. It only relies on the checkpoint state to avoid duplicate work. 
- **Transformations**: Comes with a few common transformations, and it is very easy to a add new one.
High level details on checkpointing/state manegment can be found [here](https://github.com/sharvanath/kromium/blob/main/core/README.md).

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

More details on how to configure auth for storage provider can be found [here](https://github.com/sharvanath/kromium/tree/main/storage).

## Supported transforms
**Generic file transforms**
```
- Identity: does not change the file content.
- GzipCompress/GzipDecompress: The arguments for compression level.
- Encrypt: The arguments are the compression algorithms and key. Default is simple openssl encryption to the file.
- Decrypt: Same as encryption but decrypts the file instead.
- Sed: Use sed commands for modifying text.
```

## Installation
### Using go (>=1.16):
`go get github.com/sharvanath/kromium`

### Using [Homebrew](https://brew.sh/) (on Mac OS X/Linux):
`brew install sharvanath/core/kromium`
to confirm the installation
`brew test kromium`

### Downloading Binaries:
After downloading the latest release binary for your platform from [here](https://github.com/sharvanath/kromium/releases).
To confirm the installation
`kromium -version`

## Build and run from source
```
git clone https://github.com/sharvanath/kromium 
cd kromium
go run main.go --run examples/identity_local.cue
```

## Running the first pipeline
Download the identity local pipeline config: 
1. `curl -o /tmp/identity_local.cue https://raw.githubusercontent.com/sharvanath/kromium/main/examples/identity_local.cue`
2. `rm -rf /tmp/kr; mkdir -p /tmp/kr/src; mkdir /tmp/kr/dst; mkdir /tmp/kr/state; for i in {1..16}; do echo "hello" > /tmp/kr/src/hello_$i; done`
3. `diff /tmp/kr/src /tmp/kr/dst` should show all 16 source files.
4. `kromium -run /tmp/identity_local.cue`
5. `diff /tmp/kr/src /tmp/kr/dst` should be empty now.

## Troubleshooting
* If you see Mac blocking the binary since it's untrusted. Follow [this](https://github.molgen.mpg.de/pages/bs/macOSnotes/mac/mac_procs_unsigned.html)
* If you see errors of the form "too many open files" or "could not resolve address", check `ulimit -n` and if it is less than 1024 set it to a higher value `ulimit -n 1024`.
