package core

type SyncConfig struct {
	SourceBucket string
	DestinationBucket string
	NameSuffix string
	Operations []string
}
