{
 SourceBucket: "file:///tmp/src",
 DestinationBucket: "file:///tmp/dst",
 StateBucket: "file:///tmp/state",
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
