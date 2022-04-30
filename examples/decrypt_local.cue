{
 SourceBucket: "file:///tmp/dst",
 DestinationBucket: "file:///tmp/src",
 StateBucket: "file:///tmp/state",
 StripSuffix: "_1",
 Transforms: [
   {
     Type: "Decrypt",
     Args: {
        HexKey: "6368616e676520746869732070617373"
     }
   }
 ]
}
