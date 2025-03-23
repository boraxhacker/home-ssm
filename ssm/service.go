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

	err := result.dataStore.initializeDataStore()

	if err != nil {
		log.Panicln("Unable to generate default key.", err)
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

func (service *ParameterService) GetParameter(request *GetParameterRequest) (*GetParameterResponse, awslib.APIError) {

	result, apiError := service.dataStore.getParameter(request.Name)
	if apiError.Code != "" {
		return nil, apiError
	}

	if result.Type == "SecureString" && request.WithDecryption {

		decryptedValue, err := service.dataStore.decrypt(result.Value, result.KeyId)
		if err != nil {
			return nil, SsmErrorCodes[ErrInvalidKeyId]
		}

		result.Value = decryptedValue
	}

	response := GetParameterResponse{
		Parameter: GetParameterItem{
			ARN:              service.createParameterArn(result.Name),
			DataType:         result.DataType,
			LastModifiedDate: result.LastModifiedDate,
			Name:             result.Name,
			Selector:         "",
			SourceResult:     "",
			Type:             result.Type,
			Value:            result.Value,
			Version:          result.Version,
		},
	}

	return &response, SsmErrorCodes[ErrNone]
}

func (service *ParameterService) PutParameter(
	creds *aws.Credentials, request *PutParameterRequest) (*PutParameterResponse, awslib.APIError) {

	if !service.isValidParamName(request.Name) {
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
		param.KeyId = defaultKeyId
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

func (service *ParameterService) isValidParamName(name string) bool {

	name = strings.ToLower(name)
	if strings.HasPrefix(name, "aws") || strings.HasPrefix(name, "ssm") {
		return false
	}

	return true
}

func (service *ParameterService) createParameterArn(name string) string {

	return fmt.Sprintf("arn:aws:ssm:%s:%s:parameter/%s",
		service.region, service.accountId, strings.TrimPrefix(name, "/"))
}
