package core

type TransformConfig struct {
	Name string
	Args interface{}
}

type PipelineConfig struct {
	SourceBucket string
	DestinationBucket string
	NameSuffix string
	Transforms []TransformConfig
}
