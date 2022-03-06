package config

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/sharvanath/kromium/core"
	"io/ioutil"
)

var schema = `#BaseTransform: {
    Type: string
    Args?: _
}

#Identity: #BaseTransform& {
   Type: "Identity"
}

#GzipCompress: #BaseTransform& {
    Type: "GzipCompress"
    Args?: {
        level: int
    }
}

#GzipDecompress: #BaseTransform& {
   Type: "GzipDecompress"
}

#Encrypt: #BaseTransform& {
   Type: "Encrypt"
   Args?: {
    HexKey: string
   }
}

#Decrypt: #BaseTransform& {
   Type: "Decrypt"
   Args?: {
    HexKey: string
   }
}

#Sed: #BaseTransform& {
   Type: "Sed"
   Args?: string
}

#Transform: (#GzipCompress | #GzipDecompress | #Encrypt | #Decrypt | #Sed | #Identity)

#Bucket: string & (=~"file:///" | =~"gs://" | =~"s3://")

#Pipeline: {
 SourceBucket: #Bucket,
 DestinationBucket: #Bucket,
 StateBucket: #Bucket,
 NameSuffix: string,
 Transforms: [...#Transform]
}`

func validatePipelineConfigString(config string) error {
	ctx := cuecontext.New()
	schemaVal:= ctx.CompileString(schema)
	configVal := ctx.CompileString(config)
	return schemaVal.LookupPath(cue.ParsePath("#Pipeline")).Subsume(configVal)
}


func ValidatePipelineConfig(configPath string) error {
	buf, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	return validatePipelineConfigString(string(buf))
}

func ConvertToPipelineConfig(configPath string) (*core.PipelineConfig, error) {
	buf, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := string(buf)
	if err = validatePipelineConfigString(config); err != nil {
		return nil, err
	}

	ctx := cuecontext.New()
	configVal := ctx.CompileString(config)
	var c core.PipelineConfig
	if err = configVal.Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}