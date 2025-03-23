package ssm

import (
	"home-ssm/awslib"
)

type Parameter struct {
	AllowedPattern   string               `json:"AllowedPattern"`
	DataType         ParamDataType        `json:"DataType"`
	Description      string               `json:"Description"`
	KeyId            string               `json:"KeyId"`
	LastModifiedDate float64              `json:"LastModifiedDate"`
	LastModifiedUser string               `json:"LastModifiedUser"`
	Name             string               `json:"Name"`
	Policies         string               `json:"Policies"`
	Tags             []awslib.ResourceTag `json:"Tags"`
	Tier             ParamTier            `json:"Tier"`
	Type             ParamType            `json:"Type"`
	Value            string               `json:"Value"`
	Version          int64                `json:"Version"`
}

type GetParameterRequest struct {
	Name           string `json:"Name"`
	WithDecryption bool   `json:"WithDecryption"`
}

type ParamDataType string

const (
	TextDataType           ParamDataType = "text"
	Ec2ImageDataType       ParamDataType = "aws:ec2:image"
	SsmIntegrationDataType ParamDataType = "aws:ssm:int64egration"
)

type ParamTier string

const (
	StandardTier    ParamTier = "Standard"
	AdvancedTier    ParamTier = "Advanced"
	IntelligentTier ParamTier = "Intelligent-Tiering"
)

type ParamType string

const (
	StringType       ParamType = "String"
	StringListType   ParamType = "StringList"
	SecureStringType ParamType = "SecureString"
)

type PutParameterRequest struct {
	AllowedPattern string               `json:"AllowedPattern"`
	DataType       ParamDataType        `json:"DataType"`
	Description    string               `json:"Description"`
	KeyId          string               `json:"KeyId"`
	Name           string               `json:"Name"`
	Overwrite      bool                 `json:"Overwrite"`
	Policies       string               `json:"Policies"`
	Tags           []awslib.ResourceTag `json:"Tags"`
	Tier           ParamTier            `json:"Tier"`
	Type           ParamType            `json:"Type"`
	Value          string               `json:"Value"`
}

type DeleteParameterRequest struct {
	Name string `json:"Name"`
}

type DescribeParametersRequest struct {
	MaxResults int64  `json:"MaxResults"`
	NextToken  string `json:"NextToken"`
}

type GetParameterItem struct {
	ARN              string        `json:"ARN"`
	DataType         ParamDataType `json:"DataType"`
	LastModifiedDate float64       `json:"LastModifiedDate"`
	Name             string        `json:"Name"`
	Selector         string        `json:"Selector"`
	SourceResult     string        `json:"SourceResult"`
	Type             ParamType     `json:"Type"`
	Value            string        `json:"Value"`
	Version          int64         `json:"Version"`
}

type GetParameterResponse struct {
	Parameter GetParameterItem `json:"Parameter"`
}

type PutParameterResponse struct {
	Tier    ParamTier `json:"Tier"`
	Version int64     `json:"Version"`
}

type DeleteParameterResponse struct{}

type DescribeParameterItem struct {
	AllowedPattern   string        `json:"AllowedPattern"`
	ARN              string        `json:"ARN"`
	DataType         ParamDataType `json:"DataType"`
	Description      string        `json:"Description"`
	KeyId            string        `json:"KeyId"`
	LastModifiedDate float64       `json:"LastModifiedDate"`
	LastModifiedUser string        `json:"LastModifiedUser"`
	Name             string        `json:"Name"`
	Policies         string        `json:"Policies"`
	Tier             ParamTier     `json:"Tier"`
	Type             ParamType     `json:"Type"`
	Version          int64         `json:"Version"`
}

type DescribeParametersResponse struct {
	Parameters []DescribeParameterItem `json:"Parameters"`
	NextToken  string                  `json:"NextToken"`
}
