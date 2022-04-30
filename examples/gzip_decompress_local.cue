{
 SourceBucket: "file:///tmp/dst",
 DestinationBucket: "file:///tmp/src",
 StateBucket: "file:///tmp/state",
 StripSuffix: ".gz",
 Transforms: [
   {
     Type: "GzipDecompress",
   }
 ]
}
