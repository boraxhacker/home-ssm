package ssm

import (
	"fmt"
	"home-ssm/awslib"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/dgraph-io/badger/v4"
)

type ParameterService struct {
	dataStore *DataStore
	accountId string
	region    string
}

func NewParameterService(region string, accountId string, db *badger.DB) *ParameterService {

	result := ParameterService{
		region:    region,
		accountId: accountId,
		dataStore: NewDataStore(db),
	}

	return &result
}

func (service *ParameterService) Close() {

	if service.dataStore != nil && service.dataStore.db != nil {
		err := service.dataStore.db.Close()
		if err != nil {

			log.Println("Failed to close database.", err)
		}
	}
}

func (service *ParameterService) DeleteParameter(
	request *DeleteParameterRequest) (*DeleteParameterResponse, awslib.APIError) {

	if service.isInvalidParamName(request.Name) {
		return nil, SsmErrorCodes[ErrInvalidName]
	}

	apiError := service.dataStore.delete(request.Name)
	if apiError.Code != "" {
		return nil, apiError
	}

	return &DeleteParameterResponse{}, SsmErrorCodes[ErrNone]
}

func (service *ParameterService) DeleteParameters(
	request *DeleteParametersRequest) (*DeleteParametersResponse, awslib.APIError) {

	var response DeleteParametersResponse
	for _, name := range request.Names {

		if service.isInvalidParamName(name) {

			response.InvalidParameters = append(response.InvalidParameters, name)

		} else {

			apiError := service.dataStore.delete(name)
			if apiError.Code == "" {

				response.DeleteParameters = append(response.DeleteParameters, name)

			} else {

				response.InvalidParameters = append(response.InvalidParameters, name)
			}
		}
	}

	return &response, SsmErrorCodes[ErrNone]
}

func (service *ParameterService) DescribeParameters(
	request *DescribeParametersRequest) (*DescribeParametersResponse, awslib.APIError) {

	var parameters []Parameter

	var keyFilters []KeyFilter

	for _, filter := range request.ParameterFilters {

		if (filter.Key == NameKeyFilter && filter.Option == EqualsOptionFilter) ||
			(filter.Key == NameKeyFilter && filter.Option == BeginsWithOptionFilter) {

			for _, value := range filter.Values {

				keyFilter := KeyFilter{
					Path:       value,
					StartsWith: filter.Option == BeginsWithOptionFilter,
				}

				keyFilters = append(keyFilters, keyFilter)
			}
		}

		if (filter.Key == PathKeyFilter && filter.Option == EqualsOptionFilter) ||
			(filter.Key == PathKeyFilter && filter.Option == BeginsWithOptionFilter) {

			for _, value := range filter.Values {

				keyFilter := KeyFilter{
					Path:       value,
					StartsWith: false,
				}

				keyFilters = append(keyFilters, keyFilter)

				if filter.Option == BeginsWithOptionFilter {

					keyFilter := KeyFilter{
						Path:       value + "/",
						StartsWith: true,
					}

					keyFilters = append(keyFilters, keyFilter)
				}
			}
		}
	}
	parameters, apiError := service.dataStore.findParametersByKey(keyFilters)

	if apiError.Code != "" {

		return nil, apiError
	}

	var response DescribeParametersResponse
	for _, param := range parameters {

		response.Parameters = append(response.Parameters, *param.toDescribeParameterItem(service.createParameterArn))
	}

	return &response, SsmErrorCodes[ErrNone]
}

func (service *ParameterService) GetParameter(
	request *GetParameterRequest) (*GetParameterResponse, awslib.APIError) {

	result, apiError := service.getParameterByName(request.Name, request.WithDecryption)
	if apiError.Code != "" {
		return nil, apiError
	}

	response := GetParameterResponse{
		Parameter: *result.toGetParameterItem(service.createParameterArn),
	}

	return &response, SsmErrorCodes[ErrNone]
}

func (service *ParameterService) GetParameters(
	request *GetParametersRequest) (*GetParametersResponse, awslib.APIError) {

	var response GetParametersResponse
	for _, name := range request.Names {

		param, apiError := service.getParameterByName(name, request.WithDecryption)
		if apiError.Code == "" {
			item := param.toGetParameterItem(service.createParameterArn)
			response.Parameters = append(response.Parameters, *item)
		} else {
			response.InvalidParameters = append(response.InvalidParameters, name)
		}
	}

	return &response, SsmErrorCodes[ErrNone]
}

func (service *ParameterService) GetParametersByPath(
	request *GetParametersByPathRequest) (*GetParametersByPathResponse, awslib.APIError) {

	filters := []KeyFilter{KeyFilter{Path: request.Path, StartsWith: false}}
	if request.Recursive {
		filters = append(filters, KeyFilter{Path: request.Path + "/", StartsWith: true})
	}

	parameters, apiError := service.dataStore.findParametersByKey(filters)

	if apiError.Code != "" {

		return nil, apiError
	}

	var response GetParametersByPathResponse
	for _, param := range parameters {

		if request.WithDecryption && param.Type == SecureStringType {

			decryptedValue, err := service.dataStore.decrypt(param.Value, param.KeyId)
			if err != nil {
				return nil, SsmErrorCodes[ErrInvalidKeyId]
			}

			param.Value = decryptedValue
		}

		response.Parameters = append(response.Parameters, *param.toGetParameterItem(service.createParameterArn))
	}

	return &response, SsmErrorCodes[ErrNone]
}

func (service *ParameterService) PutParameter(
	creds *aws.Credentials, request *PutParameterRequest) (*PutParameterResponse, awslib.APIError) {

	if service.isInvalidParamName(request.Name) {
		return nil, SsmErrorCodes[ErrInvalidName]
	}

	param := Parameter{
		AllowedPattern:   request.AllowedPattern,
		DataType:         request.DataType,
		Description:      request.Description,
		LastModifiedDate: float64(time.Now().UnixNano()) / float64(time.Second),
		LastModifiedUser: service.createUserArn(creds),
		Name:             request.Name,
		Policies:         request.Policies,
		Tags:             request.Tags,
		Tier:             request.Tier,
		Type:             request.Type,
		Value:            request.Value,
	}

	if param.Tier != StandardTier && param.Tier != AdvancedTier && param.Tier != IntelligentTier {
		if param.Tier == "" {

			param.Tier = StandardTier

		} else {

			return nil, SsmErrorCodes[ErrInvalidTier]
		}
	}

	if request.Type == SecureStringType {
		encryptedValue, err := service.dataStore.encrypt(param.Value, request.KeyId)
		if err != nil {
			return nil, awslib.ErrorCodes[awslib.ErrInternalError]
		}
		param.Value = encryptedValue
		param.KeyId = DefaultKeyId
	}

	newVersion, apiError := service.dataStore.putParameter(param.Name, &param, request.Overwrite)
	if apiError.Code != "" {

		return nil, apiError
	}

	return &PutParameterResponse{Tier: param.Tier, Version: newVersion}, SsmErrorCodes[ErrNone]
}

func (service *ParameterService) createUserArn(creds *aws.Credentials) string {

	return fmt.Sprintf("arn:aws:iam::%s:user/%s", service.accountId, creds.Source)
}

func (service *ParameterService) getParameterByName(name string, withDecryption bool) (*Parameter, awslib.APIError) {

	if service.isInvalidParamName(name) {
		return nil, SsmErrorCodes[ErrInvalidName]
	}

	result, apiError := service.dataStore.getParameter(name)
	if apiError.Code != "" {
		return nil, apiError
	}

	if result.Type == "SecureString" && withDecryption {

		decryptedValue, err := service.dataStore.decrypt(result.Value, result.KeyId)
		if err != nil {
			return nil, SsmErrorCodes[ErrInvalidKeyId]
		}

		result.Value = decryptedValue
	}

	return result, SsmErrorCodes[ErrNone]
}

func (service *ParameterService) isInvalidParamName(name string) bool {

	name = strings.ToLower(name)
	if strings.HasPrefix(name, "aws") || strings.HasPrefix(name, "ssm") {
		return true
	}

	return false
}

func (service *ParameterService) createParameterArn(name string) string {

	return fmt.Sprintf("arn:aws:ssm:%s:%s:parameter/%s",
		service.region, service.accountId, strings.TrimPrefix(name, "/"))
}
