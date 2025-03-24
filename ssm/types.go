package ssm

import (
	"home-ssm/awslib"
	"strings"
)

type ParamName string

func (p ParamName) CheckValidity() error {

	name := strings.TrimPrefix(strings.ToLower(string(p)), "/")
	if strings.HasPrefix(name, "aws") || strings.HasPrefix(name, "ssm") {
		return ErrInvalidName
	}

	return nil
}

type Parameter struct {
	AllowedPattern   string               `json:"AllowedPattern"`
	DataType         ParamDataType        `json:"DataType"`
	Description      string               `json:"Description"`
	KeyId            string               `json:"KeyId"`
	LastModifiedDate float64              `json:"LastModifiedDate"`
	LastModifiedUser string               `json:"LastModifiedUser"`
	Name             ParamName            `json:"Name"`
	Policies         string               `json:"Policies"`
	Tags             []awslib.ResourceTag `json:"Tags"`
	Tier             ParamTier            `json:"Tier"`
	Type             ParamType            `json:"Type"`
	Value            string               `json:"Value"`
	Version          int64                `json:"Version"`
}

type GetParameterRequest struct {
	Name           ParamName `json:"Name"`
	WithDecryption bool      `json:"WithDecryption"`
}

type GetParametersRequest struct {
	Names          []ParamName `json:"Names"`
	WithDecryption bool        `json:"WithDecryption"`
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

func (dtype ParamDataType) isValid(allowBlank bool) bool {

	if allowBlank && dtype == "" {
		return true
	}

	return dtype == "" || dtype == TextDataType || dtype == Ec2ImageDataType || dtype == SsmIntegrationDataType
}

type ParamTier string

const (
	StandardTier    ParamTier = "Standard"
	AdvancedTier    ParamTier = "Advanced"
	IntelligentTier ParamTier = "Intelligent-Tiering"
)

func (tier ParamTier) isValid(allowBlank bool) bool {

	if allowBlank && tier == "" {
		return true
	}

	return tier == StandardTier || tier == AdvancedTier || tier == IntelligentTier
}

type ParamType string

const (
	StringType       ParamType = "String"
	StringListType   ParamType = "StringList"
	SecureStringType ParamType = "SecureString"
)

func (ptype ParamType) isValid(allowBlank bool) bool {

	if allowBlank && ptype == "" {
		return true
	}

	return ptype == StringType || ptype == StringListType || ptype == SecureStringType
}

type PutParameterRequest struct {
	AllowedPattern string               `json:"AllowedPattern"`
	DataType       ParamDataType        `json:"DataType"`
	Description    string               `json:"Description"`
	KeyId          string               `json:"KeyId"`
	Name           ParamName            `json:"Name"`
	Overwrite      bool                 `json:"Overwrite"`
	Policies       string               `json:"Policies"`
	Tags           []awslib.ResourceTag `json:"Tags"`
	Tier           ParamTier            `json:"Tier"`
	Type           ParamType            `json:"Type"`
	Value          string               `json:"Value"`
}

func (request *PutParameterRequest) CheckValidity() error {

	if !request.Tier.isValid(true) {
		return ErrInvalidTier
	}

	if !request.DataType.isValid(true) {
		return ErrInvalidDataType
	}

	if !request.Type.isValid(true) {
		return ErrInvalidTier
	}

	return request.Name.CheckValidity()
}

type DeleteParameterRequest struct {
	Name ParamName `json:"Name"`
}

type DeleteParametersRequest struct {
	Names []ParamName `json:"Names"`
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

func (ktype KeyFilterType) isValid() bool {

	return ktype == NameKeyFilter || ktype == TypeKeyFilter || ktype == KeyIdFilter ||
		ktype == PathKeyFilter || ktype == LabelKeyFilter || ktype == TierKeyFilter ||
		ktype == DataTypeKeyFilter
}

type OptionFilterType string

const (
	EqualsOptionFilter     = "Equals"
	BeginsWithOptionFilter = "BeginsWith"
	RecursiveOptionFilter  = "Recursive"
	OneLevelOptionFilter   = "OneLevel"
)

func (ftype OptionFilterType) isValid() bool {

	return ftype == EqualsOptionFilter || ftype == BeginsWithOptionFilter ||
		ftype == RecursiveOptionFilter || ftype == OneLevelOptionFilter
}

type ParameterFilter struct {
	Key    KeyFilterType    `json:"Key"`
	Option OptionFilterType `json:"Option"`
	Values []string         `json:"Values"`
}

func (filter *ParameterFilter) CheckValidity() error {

	if !filter.Key.isValid() {
		return ErrInvalidFilterKey
	}

	if !filter.Option.isValid() {
		return ErrInvalidFilterOption
	}

	if filter.Key == PathKeyFilter {
		if filter.Option != OneLevelOptionFilter && filter.Option != RecursiveOptionFilter {
			return ErrInvalidFilterOption
		}
	}

	if filter.Key == NameKeyFilter {
		if filter.Option != EqualsOptionFilter && filter.Option != BeginsWithOptionFilter {
			return ErrInvalidFilterOption
		}
	}

	return nil
}

type DeleteParametesResponse struct {
}

type DeleteParametersResponse struct {
	DeleteParameters  []ParamName `json:"DeletedParameters"`
	InvalidParameters []ParamName `json:"InvalidParameters"`
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
	Name             ParamName     `json:"Name"`
	Selector         string        `json:"Selector"`
	SourceResult     string        `json:"SourceResult"`
	Type             ParamType     `json:"Type"`
	Value            string        `json:"Value"`
	Version          int64         `json:"Version"`
}

type GetParametersResponse struct {
	InvalidParameters []ParamName        `json:"InvalidParameters"`
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
	Name             ParamName     `json:"Name"`
	Policies         string        `json:"Policies"`
	Tier             ParamTier     `json:"Tier"`
	Type             ParamType     `json:"Type"`
	Version          int64         `json:"Version"`
}

type DescribeParametersResponse struct {
	NextToken  string                  `json:"NextToken"`
	Parameters []DescribeParameterItem `json:"Parameters"`
}

type ParameterArnGenerator func(ParamName) string

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
