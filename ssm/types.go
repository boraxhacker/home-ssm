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

type GetParametersRequest struct {
	Names          []string `json:"Names"`
	WithDecryption bool     `json:"WithDecryption"`
}

type GetParametersByPathRequest struct {
	MaxResults       int64             `json:"MaxResults"`
	NextToken        string            `json:"NextToken"`
	ParameterFilters []ParameterFilter `json:"ParameterFilters"`
	Path             string            `json:"Path"`
	Recursive        bool              `json:"Recursive"`
	WithDecryption   bool              `json:"WithDecryption"`
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

type DeleteParametersRequest struct {
	Names []string `json:"Names"`
}

type KeyFilterType string

const (
	NameKeyFilter     = "Name"
	TypeKeyFilter     = "Type"
	KeyIdFilter       = "KeyId"
	PathKeyFilter     = "Path"
	LabelKeyFilter    = "Label"
	TierKeyFilter     = "Tier"
	DataTypeKeyFilter = "DataType"
)

type OptionFilterType string

const (
	EqualsOptionFilter     = "Equals"
	BeginsWithOptionFilter = "BeginsWith"
)

type ParameterFilter struct {
	Key    KeyFilterType    `json:"Key"`
	Option OptionFilterType `json:"Option"`
	Values []string         `json:"Values"`
}

type DeleteParametesResponse struct {
}

type DeleteParametersResponse struct {
	DeleteParameters  []string `json:"DeletedParameters"`
	InvalidParameters []string `json:"InvalidParameters"`
}

type DescribeParametersRequest struct {
	MaxResults       int64             `json:"MaxResults"`
	NextToken        string            `json:"NextToken"`
	ParameterFilters []ParameterFilter `json:"ParameterFilters"`
	Shared           bool              `json:"Shared"`
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

type GetParametersResponse struct {
	InvalidParameters []string           `json:"InvalidParameters"`
	Parameters        []GetParameterItem `json:"Parameters"`
}

type GetParametersByPathResponse struct {
	NextToken  string             `json:"NextToken"`
	Parameters []GetParameterItem `json:"Parameters"`
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
	NextToken  string                  `json:"NextToken"`
	Parameters []DescribeParameterItem `json:"Parameters"`
}

type ParameterArnGenerator func(string) string

func (param *Parameter) toGetParameterItem(arnGenerator ParameterArnGenerator) *GetParameterItem {

	return &GetParameterItem{
		ARN:              arnGenerator(param.Name),
		DataType:         param.DataType,
		LastModifiedDate: param.LastModifiedDate,
		Name:             param.Name,
		Selector:         "",
		SourceResult:     "",
		Type:             param.Type,
		Value:            param.Value,
		Version:          param.Version,
	}
}

func (param *Parameter) toDescribeParameterItem(arnGenerator ParameterArnGenerator) *DescribeParameterItem {

	return &DescribeParameterItem{
		AllowedPattern:   param.AllowedPattern,
		ARN:              arnGenerator(param.Name),
		DataType:         param.DataType,
		Description:      param.Description,
		KeyId:            param.KeyId,
		LastModifiedDate: param.LastModifiedDate,
		LastModifiedUser: param.LastModifiedUser,
		Name:             param.Name,
		Policies:         param.Policies,
		Tier:             param.Tier,
		Type:             param.Type,
		Version:          param.Version,
	}
}
