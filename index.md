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
### Option1: Using go:
You need have golang (>1.16) and git installed to be able to use this method.
1. Follow [https://go.dev/doc/install](https://go.dev/doc/install) to install the latest go.
2. Install git using you favorite package manager.
4. Install Kromium:
   If you are using go >= 1.17 use:
   ```
   go install github.com/sharvanath/kromium@latest
   ```
   or
   If you are using go < 1.17 use:
   ```
   go get github.com/sharvanath/kromium
   ```
   Note that go < 1.16 is not supported.
5. Make sure to include $GOBIN to the $PATH.


### Option2: Using [Homebrew](https://brew.sh/) (on Mac OS X/Linux):
```
brew install sharvanath/core/kromium
```

To confirm the installation
```
brew test kromium
```

### Option3: Downloading Binaries:
Assuming the go(>=1.16) is already installed.
Download the latest release binary for your platform from [here](https://github.com/sharvanath/kromium/releases).
For example, for linux amd64:
```
curl -L -o kromium.tgz https://github.com/sharvanath/kromium/releases/download/v0.1.7/kromium-v0.1.7-linux-amd64.tar.gz
tar xvf kromium.tgz
sudo mv kromium /usr/local/bin/
```
To confirm the installation
`kromium -version`

### Build and run from source
```
git clone https://github.com/sharvanath/kromium 
cd kromium
go run main.go --run examples/identity_local.cue
```

## Running the first pipeline
1. Download the identity local pipeline config: 
```
curl -o /tmp/identity_local.cue https://raw.githubusercontent.com/sharvanath/kromium/main/examples/identity_local.cue
```
2. Create test src, dst and state directories:
```
rm -rf /tmp/kr; mkdir -p /tmp/kr/src; mkdir /tmp/kr/dst; mkdir /tmp/kr/state; for i in {1..16}; do echo "hello" > /tmp/kr/src/hello_$i; done
```
3. Confirm the difference between src and dst before the run (should show all 16 source files).
```
diff /tmp/kr/src /tmp/kr/dst
```
5. Run the pipeline
```
kromium -run /tmp/identity_local.cue
```
7. Confirm there is no difference between src and dst after the run
```
diff /tmp/kr/src /tmp/kr/dst` # should be empty now.
```

## Troubleshooting
* If you see Mac blocking the binary since it's untrusted. Follow [this](https://github.molgen.mpg.de/pages/bs/macOSnotes/mac/mac_procs_unsigned.html)
* If you see errors of the form "too many open files" or "could not resolve address", check `ulimit -n` and if it is less than 1024 set it to a higher value `ulimit -n 1024`.
