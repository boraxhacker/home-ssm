package ssm

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/dgraph-io/badger/v4"
)

type ParameterService struct {
	dataStore *DataStore
	accountId string
	region    string
}

func NewParameterService(region string, accountId string, dataStore *DataStore) *ParameterService {

	result := ParameterService{
		region:    region,
		accountId: accountId,
		dataStore: dataStore,
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
	request *awsssm.DeleteParameterInput) (*awsssm.DeleteParameterOutput, error) {

	paramName, err := NewParamName(request.Name)
	if err != nil {
		return nil, ErrInvalidName
	}

	err = service.dataStore.delete(string(paramName.asPathName()))
	if err != nil {
		return nil, err
	}

	return &awsssm.DeleteParameterOutput{}, nil
}

func (service *ParameterService) DeleteParameters(
	request *awsssm.DeleteParametersInput) (*awsssm.DeleteParametersOutput, error) {

	var response awsssm.DeleteParametersOutput
	for _, name := range request.Names {

		paramName, err := NewParamName(&name)
		if err != nil {

			response.InvalidParameters = append(response.InvalidParameters, name)

		} else {

			err := service.dataStore.delete(string(paramName.asPathName()))
			if err == nil {

				response.DeletedParameters = append(response.DeletedParameters, name)

			} else {

				response.InvalidParameters = append(response.InvalidParameters, name)
			}
		}
	}

	return &response, nil
}

func (service *ParameterService) DescribeParameters(
	request *awsssm.DescribeParametersInput) (*DescribeParametersResponse, error) {

	// TODO incomplete implementation

	var parameters []ParameterData

	var filters []string

	for _, awsfilter := range request.ParameterFilters {

		filter, err := NewParameterFilter(&awsfilter)
		if err != nil {
			return nil, err
		}

		if (filter.Key == NameKeyFilter && filter.Option == EqualsOptionFilter) ||
			(filter.Key == NameKeyFilter && filter.Option == BeginsWithOptionFilter) {

			for _, value := range filter.Values {

				paramName, err := NewParamName(&value)
				if err != nil {
					return nil, err
				}

				if filter.Option == BeginsWithOptionFilter {

					filters = append(filters, paramName.asBeginsWithRegex())
				} else {

					filters = append(filters, paramName.asEqualsRegex())
				}
			}
		}

		if (filter.Key == PathKeyFilter && filter.Option == RecursiveOptionFilter) ||
			(filter.Key == PathKeyFilter && filter.Option == OneLevelOptionFilter) {

			for _, value := range filter.Values {

				paramPath, err := NewParamPath(&value)
				if err != nil {
					return nil, err
				}

				if filter.Option == RecursiveOptionFilter {
					filters = append(filters, paramPath.asRecursiveRegex())
				} else {
					filters = append(filters, paramPath.asOneLevelRegex())
				}
			}
		}
	}

	parameters, err := service.dataStore.findParametersByKey(filters)
	if err != nil {

		return nil, err
	}

	var response DescribeParametersResponse
	for _, param := range parameters {

		response.Parameters = append(response.Parameters,
			*param.toDescribeParameterItem(service.createParameterArn))
	}

	return &response, nil
}

func (service *ParameterService) GetParameter(
	request *awsssm.GetParameterInput) (*GetParameterResponse, error) {

	result, err := service.getParameterByName(
		aws.ToString(request.Name), aws.ToBool(request.WithDecryption))
	if err != nil {
		return nil, err
	}

	response := GetParameterResponse{
		Parameter: result.toGetParameterItem(service.createParameterArn),
	}

	return &response, nil
}

func (service *ParameterService) GetParameters(
	request *awsssm.GetParametersInput) (*GetParametersResponse, error) {

	var response GetParametersResponse
	for _, name := range request.Names {

		param, err := service.getParameterByName(name, aws.ToBool(request.WithDecryption))
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
	request *awsssm.GetParametersByPathInput) (*GetParametersByPathResponse, error) {

	// TODO incomplete implementation

	paramPath, err := NewParamPath(request.Path)
	if err != nil {
		return nil, err
	}

	for _, filter := range request.ParameterFilters {
		_, err := NewParameterFilter(&filter)
		if err != nil {
			return nil, err
		}
	}

	var filters []string
	if aws.ToBool(request.Recursive) {
		filters = append(filters, paramPath.asRecursiveRegex())
	} else {
		filters = append(filters, paramPath.asOneLevelRegex())
	}

	parameters, err := service.dataStore.findParametersByKey(filters)
	if err != nil {

		return nil, err
	}

	var response GetParametersByPathResponse
	for _, param := range parameters {

		if aws.ToBool(request.WithDecryption) && param.Type == awstypes.ParameterTypeSecureString {

			decryptedValue, err := service.dataStore.decrypt(param.Value, param.KeyId)
			if err != nil {
				return nil, ErrInvalidKeyId
			}

			param.Value = decryptedValue
		}

		response.Parameters = append(response.Parameters,
			*param.toGetParameterItem(service.createParameterArn))
	}

	return &response, nil
}

func (service *ParameterService) PutParameter(
	creds *aws.Credentials, request *awsssm.PutParameterInput) (*awsssm.PutParameterOutput, error) {

	param, err := NewParameterData(request)
	if err != nil {
		return nil, err
	}

	param.LastModifiedUser = service.createUserArn(creds)

	if param.Type == awstypes.ParameterTypeSecureString {

		if param.KeyId == "" {
			param.KeyId = "alias/" + service.dataStore.keys[0].Alias
		}

		encryptedValue, err := service.dataStore.encrypt(param.Value, param.KeyId)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil, ErrInvalidKeyId
			}
			return nil, ErrInternalError
		}

		param.Value = encryptedValue
	}

	newVersion, err := service.dataStore.putParameter(string(param.Name), param, aws.ToBool(request.Overwrite))
	if err != nil {

		return nil, err
	}

	return &awsssm.PutParameterOutput{Tier: awstypes.ParameterTierStandard, Version: newVersion}, nil
}

func (service *ParameterService) AddTagsToResource(
	request *awsssm.AddTagsToResourceInput) (*awsssm.AddTagsToResourceOutput, error) {

	var response awsssm.AddTagsToResourceOutput
	if request.ResourceType == awstypes.ResourceTypeForTaggingParameter {

		param, err := service.getParameterByName(aws.ToString(request.ResourceId), false)
		if err != nil {
			return nil, err
		}

		for _, tag := range request.Tags {

			tagName := aws.ToString(tag.Key)
			found := false
			for _, paramTag := range param.Tags {

				if paramTag.Key == tagName {
					paramTag.Value = aws.ToString(tag.Value)
					found = true
					break
				}
			}

			if !found {

				param.Tags = append(param.Tags,
					ResourceTag{Key: tagName, Value: aws.ToString(tag.Value)})
			}
		}

		_, err = service.dataStore.putParameter(string(param.Name), param, true)
		if err != nil {
			return nil, err
		}
	}

	return &response, nil
}

func (service *ParameterService) RemoveTagsFromResource(
	request *awsssm.RemoveTagsFromResourceInput) (*awsssm.RemoveTagsFromResourceOutput, error) {

	var response awsssm.RemoveTagsFromResourceOutput
	if request.ResourceType == awstypes.ResourceTypeForTaggingParameter {

		param, err := service.getParameterByName(aws.ToString(request.ResourceId), false)
		if err != nil {
			return nil, err
		}

		for _, tag := range request.TagKeys {

			for i, paramTag := range param.Tags {

				if paramTag.Key == tag {
					param.Tags = slices.Delete(param.Tags, i, i+1)
					break
				}
			}
		}

		_, err = service.dataStore.putParameter(string(param.Name), param, true)
		if err != nil {
			return nil, err
		}
	}

	return &response, nil
}

func (service *ParameterService) ListTagsForResource(
	request *awsssm.ListTagsForResourceInput) (*awsssm.ListTagsForResourceOutput, error) {

	var response awsssm.ListTagsForResourceOutput
	if request.ResourceType == awstypes.ResourceTypeForTaggingParameter {

		param, err := service.getParameterByName(aws.ToString(request.ResourceId), false)
		if err != nil {
			return nil, err
		}

		for _, paramTag := range param.Tags {

			response.TagList = append(response.TagList,
				awstypes.Tag{Key: aws.String(paramTag.Key), Value: aws.String(paramTag.Value)})
		}
	}

	return &response, nil
}

func (service *ParameterService) createUserArn(creds *aws.Credentials) string {

	return fmt.Sprintf("arn:aws:iam::%s:user/%s", service.accountId, creds.Source)
}

func (service *ParameterService) getParameterByName(name string, withDecryption bool) (*ParameterData, error) {

	paramName, err := NewParamName(&name)
	if err != nil {
		return nil, ErrInvalidName
	}

	result, err := service.dataStore.getParameter(string(paramName.asPathName()))
	if err != nil {
		return nil, err
	}

	// always stored as path but if requested by name then return the name
	result.Name = paramName

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
