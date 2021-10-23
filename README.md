# Kromium

## What is Kromium?

Kromium is a low-code data transformation pipeline. The pipeline is represented as a JSON config. Kromium is designed for simplicity and ease of use.
A simple configuration example is the following:

```
{                                                                                  
 "SourceBucket": "gs://kromium-src",                                                     
 "DestinationBucket": "gs://kromium-dst",                                                
 "NameSuffix": "_all",                                                             
 "Operations": ["Identity": {}]                                                        
}  
```

This configuration will simply read all objects from the `kromium-src` bucket, apply the Identity Transform (which does not change the file) and write the output to the `kromium-dst` bucket. The optional `NameSuffix` argument specifies if a suffix should be applied to the object names when writing to the destination bucket, this can be used for adding filename extensions.

## Storage providers
Currently Kromium only supports GCS and Local filesystem for storage. The support for S3 (and Azure) will be added soon. The source bucket is a uri which should be fully qualified. Following are the prefixes for supported storage solution:
```
GCS: gs://bucket
Local filesystem: file://folderpath
S3: s3://bucket
```

## Supported transforms
**Generic file transforms**
```
- Identity: does not change the file content.
- Compress: The arguments are the compression algorithm.
- Encrypt: The arguments are the compression algorithms and key. Default is simple openssl encryption to the file.
- Decrypt: Same as encryption but decrypts the file instead.
```

**CSV transforms**
```
- CSVFilter: selects particular columns from CSV.
```

**Image transforms**
```
- Resize: takes the new resolution as arguements.
```

## Supported runners
By default the transformation runs on the local machine. Support for remotely running will be added soon.
