{
 SourceBucket: "file:///tmp/src",
 DestinationBucket: "file:///tmp/dst",
 StateBucket: "file:///tmp/state",
 NameSuffix: "_1",
 Transforms: [
   {
     "Type": "Sed",
     "Args":  "s/hello/hi/g"
   }
 ]
}
