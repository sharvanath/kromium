{
  SourceBucket: "file:///tmp/src",
  DestinationBucket: "file:///tmp/dst",
  StateBucket: "file:///tmp/state",
  NameSuffix: "_1",
  Transforms: [
    {
      Type: "Decrypt",
      Args: {
        HexKey: "6368616e676520746869732070617373"
      }
    }
  ]
}