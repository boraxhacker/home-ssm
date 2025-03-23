package ssm

import (
	"encoding/json"
	"errors"
	"home-ssm/awslib"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type ParameterApi struct {
	service     *ParameterService
	credentials *awslib.CredentialsProvider
}

func NewParameterApi(service *ParameterService, credentials *awslib.CredentialsProvider) *ParameterApi {

	return &ParameterApi{service: service, credentials: credentials}
}

/*
o delete-parameter
o delete-parameters
o describe-parameters
o get-parameter
o get-parameters
o get-parameters-by-path
o put-parameter
*/

func (api *ParameterApi) Handle(w http.ResponseWriter, r *http.Request) {

	creds, err := api.parseCredentials(r)
	if err != nil {
		awslib.WriteErrorResponseJSON(w, awslib.ErrorCodes[awslib.ErrInternalError], r.URL, api.credentials.Region)
		return
	}

	amztarget := r.Header.Get("X-Amz-Target")
	log.Printf("Amazon-Target: %s\n", amztarget)
	if amztarget == "AmazonSSM.DeleteParameter" {

		api.deleteParameter(w, r)

	} else if amztarget == "AmazonSSM.DeleteParameters" {

		api.deleteParameters(w, r)

	} else if amztarget == "AmazonSSM.DescribeParameters" {

		api.describeParameters(w, r)

	} else if amztarget == "AmazonSSM.GetParameter" {

		api.getParameter(w, r)

	} else if amztarget == "AmazonSSM.GetParameters" {

		api.getParameters(w, r)

	} else if amztarget == "AmazonSSM.GetParametersByPath" {

		api.getParametersByPath(w, r)

	} else if amztarget == "AmazonSSM.PutParameter" {

		api.putParameter(creds, w, r)

	} else {

		awslib.WriteErrorResponseJSON(w, awslib.ErrorCodes[awslib.ErrValidationError], r.URL, api.credentials.Region)
	}
}

func (api *ParameterApi) getParameter(w http.ResponseWriter, r *http.Request) {

	var request GetParameterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, apiErr := api.service.GetParameter(&request)
	if response == nil {

		awslib.WriteErrorResponseJSON(w, apiErr, r.URL, api.credentials.Region)
		return
	}

	awslib.WriteSuccessResponseJSON(w, response)
}

func (api *ParameterApi) getParameters(w http.ResponseWriter, r *http.Request) {

	var request GetParametersRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, apiErr := api.service.GetParameters(&request)
	if response == nil {

		awslib.WriteErrorResponseJSON(w, apiErr, r.URL, api.credentials.Region)
		return
	}

	awslib.WriteSuccessResponseJSON(w, response)
}

func (api *ParameterApi) getParametersByPath(w http.ResponseWriter, r *http.Request) {

	var request GetParametersByPathRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, apiErr := api.service.GetParametersByPath(&request)
	if response == nil {

		awslib.WriteErrorResponseJSON(w, apiErr, r.URL, api.credentials.Region)
		return
	}

	awslib.WriteSuccessResponseJSON(w, response)
}

func (api *ParameterApi) describeParameters(w http.ResponseWriter, r *http.Request) {

	var request DescribeParametersRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, apiErr := api.service.DescribeParameters(&request)
	if response == nil {

		awslib.WriteErrorResponseJSON(w, apiErr, r.URL, api.credentials.Region)
		return
	}

	awslib.WriteSuccessResponseJSON(w, response)
}

func (api *ParameterApi) deleteParameter(w http.ResponseWriter, r *http.Request) {

	var request DeleteParameterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, apiErr := api.service.DeleteParameter(&request)
	if response == nil {

		awslib.WriteErrorResponseJSON(w, apiErr, r.URL, api.credentials.Region)
		return
	}

	awslib.WriteSuccessResponseJSON(w, response)
}

func (api *ParameterApi) deleteParameters(w http.ResponseWriter, r *http.Request) {

	var request DeleteParametersRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, apiErr := api.service.DeleteParameters(&request)
	if response == nil {

		awslib.WriteErrorResponseJSON(w, apiErr, r.URL, api.credentials.Region)
		return
	}

	awslib.WriteSuccessResponseJSON(w, response)

}

func (api *ParameterApi) putParameter(creds *aws.Credentials, w http.ResponseWriter, r *http.Request) {

	var request PutParameterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, apiErr := api.service.PutParameter(creds, &request)
	if response == nil {

		awslib.WriteErrorResponseJSON(w, apiErr, r.URL, api.credentials.Region)
		return
	}

	awslib.WriteSuccessResponseJSON(w, response)
}

func (api *ParameterApi) parseCredentials(r *http.Request) (*aws.Credentials, error) {

	var result aws.Credentials
	accessKey := r.Header.Get("x-home-ssm-access-key")

	for _, cred := range api.credentials.Credentials {

		if accessKey == cred.AccessKeyID {
			result = cred
			break
		}
	}

	if result.AccessKeyID == "" {

		return nil, errors.New("unable to find access key in headers")
	}

	return &result, nil
}
