{
 SourceBucket: "s3://kromium-src",
 DestinationBucket: "s3://kromium-dst",
 StateBucket: "s3://kromium-state",
 Transforms: [
   {
     Type: "Identity"
   }
 ],
 StorageConfig: {
    S3Config: {
      Region: "us-east-1"
    }
 }
}
