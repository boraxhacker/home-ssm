package ssm

import (
	"errors"
	"fmt"
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
	request *DeleteParameterRequest) (*DeleteParameterResponse, error) {

	err := request.Name.CheckValidity()
	if err != nil {
		return nil, ErrInvalidName
	}

	err = service.dataStore.delete(string(request.Name))
	if err != nil {
		return nil, err
	}

	return &DeleteParameterResponse{}, nil
}

func (service *ParameterService) DeleteParameters(
	request *DeleteParametersRequest) (*DeleteParametersResponse, error) {

	var response DeleteParametersResponse
	for _, name := range request.Names {

		err := name.CheckValidity()
		if err != nil {

			response.InvalidParameters = append(response.InvalidParameters, name)

		} else {

			err := service.dataStore.delete(string(name))
			if err == nil {

				response.DeleteParameters = append(response.DeleteParameters, name)

			} else {

				response.InvalidParameters = append(response.InvalidParameters, name)
			}
		}
	}

	return &response, nil
}

func (service *ParameterService) DescribeParameters(
	request *DescribeParametersRequest) (*DescribeParametersResponse, error) {

	// TODO incomplete implementation

	var parameters []Parameter

	var keyFilters []KeyFilter

	for _, filter := range request.ParameterFilters {

		err := filter.CheckValidity()
		if err != nil {
			return nil, err
		}

		if (filter.Key == NameKeyFilter && filter.Option == EqualsOptionFilter) ||
			(filter.Key == NameKeyFilter && filter.Option == BeginsWithOptionFilter) {

			for _, value := range filter.Values {

				err := ParamName(value).CheckValidity()
				if err != nil {
					return nil, ErrInvalidName
				}

				keyFilter := KeyFilter{
					Path:       value,
					StartsWith: filter.Option == BeginsWithOptionFilter,
				}

				keyFilters = append(keyFilters, keyFilter)
			}
		}

		if (filter.Key == PathKeyFilter && filter.Option == RecursiveOptionFilter) ||
			(filter.Key == PathKeyFilter && filter.Option == OneLevelOptionFilter) {

			// TODO not right

			for _, value := range filter.Values {

				err := ParamName(value).CheckValidity()
				if err != nil {
					return nil, ErrInvalidName
				}

				keyFilter := KeyFilter{
					Path:       value,
					StartsWith: false,
				}

				keyFilters = append(keyFilters, keyFilter)

				if filter.Option == RecursiveOptionFilter {

					keyFilter := KeyFilter{
						Path:       value + "/",
						StartsWith: true,
					}

					keyFilters = append(keyFilters, keyFilter)
				}
			}
		}
	}

	parameters, err := service.dataStore.findParametersByKey(keyFilters)
	if err != nil {

		return nil, err
	}

	var response DescribeParametersResponse
	for _, param := range parameters {

		response.Parameters = append(response.Parameters, *param.toDescribeParameterItem(service.createParameterArn))
	}

	return &response, nil
}

func (service *ParameterService) GetParameter(
	request *GetParameterRequest) (*GetParameterResponse, error) {

	result, err := service.getParameterByName(request.Name, request.WithDecryption)
	if err != nil {
		return nil, err
	}

	response := GetParameterResponse{
		Parameter: *result.toGetParameterItem(service.createParameterArn),
	}

	return &response, nil
}

func (service *ParameterService) GetParameters(
	request *GetParametersRequest) (*GetParametersResponse, error) {

	var response GetParametersResponse
	for _, name := range request.Names {

		param, err := service.getParameterByName(name, request.WithDecryption)
		if err == nil {
			item := param.toGetParameterItem(service.createParameterArn)
			response.Parameters = append(response.Parameters, *item)
		} else {
			response.InvalidParameters = append(response.InvalidParameters, name)
		}
	}

	return &response, nil
}

func (service *ParameterService) GetParametersByPath(
	request *GetParametersByPathRequest) (*GetParametersByPathResponse, error) {

	// TODO incomplete implementation

	err := ParamName(request.Path).CheckValidity()
	if err != nil {
		return nil, ErrInvalidName
	}

	for _, filter := range request.ParameterFilters {
		err := filter.CheckValidity()
		if err != nil {
			return nil, err
		}
	}

	filters := []KeyFilter{{Path: request.Path, StartsWith: false}}
	if request.Recursive {
		filters = append(filters, KeyFilter{Path: request.Path + "/", StartsWith: true})
	}

	parameters, err := service.dataStore.findParametersByKey(filters)
	if err != nil {

		return nil, err
	}

	var response GetParametersByPathResponse
	for _, param := range parameters {

		if request.WithDecryption && param.Type == SecureStringType {

			decryptedValue, err := service.dataStore.decrypt(param.Value, param.KeyId)
			if err != nil {
				return nil, ErrInvalidKeyId
			}

			param.Value = decryptedValue
		}

		response.Parameters = append(response.Parameters, *param.toGetParameterItem(service.createParameterArn))
	}

	return &response, nil
}

func (service *ParameterService) PutParameter(
	creds *aws.Credentials, request *PutParameterRequest) (*PutParameterResponse, error) {

	err := request.CheckValidity()
	if err != nil {
		return nil, err
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

	if param.Tier == "" {
		param.Tier = StandardTier
	}

	if param.DataType == "" {
		param.DataType = TextDataType
	}

	if request.Type == SecureStringType {
		encryptedValue, err := service.dataStore.encrypt(param.Value, DefaultKeyId)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil, ErrInvalidKeyId
			}
			return nil, ErrInternalError
		}
		param.Value = encryptedValue
		param.KeyId = "alias/" + DefaultKeyId
	}

	newVersion, err := service.dataStore.putParameter(string(param.Name), &param, request.Overwrite)
	if err != nil {

		return nil, err
	}

	return &PutParameterResponse{Tier: param.Tier, Version: newVersion}, nil
}

func (service *ParameterService) createUserArn(creds *aws.Credentials) string {

	return fmt.Sprintf("arn:aws:iam::%s:user/%s", service.accountId, creds.Source)
}

func (service *ParameterService) getParameterByName(name ParamName, withDecryption bool) (*Parameter, error) {

	err := name.CheckValidity()
	if err != nil {
		return nil, ErrInvalidName
	}

	result, err := service.dataStore.getParameter(string(name))
	if err != nil {
		return nil, err
	}

	if result.Type == "SecureString" && withDecryption {

		decryptedValue, err := service.dataStore.decrypt(result.Value, result.KeyId)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil, ErrInvalidKeyId
			}
			return nil, ErrInternalError
		}

		result.Value = decryptedValue
	}

	return result, nil
}

func (service *ParameterService) createParameterArn(name ParamName) string {

	return fmt.Sprintf("arn:aws:ssm:%s:%s:parameter/%s",
		service.region, service.accountId, strings.TrimPrefix(string(name), "/"))
}
