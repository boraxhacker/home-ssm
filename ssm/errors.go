package ssm

import (
	"errors"
	"home-ssm/awslib"
	"net/http"
)

var (
	ErrParameterNotFound        = errors.New("The ParameterData Name provided does not exist.")
	ErrParameterAlreadyExists   = errors.New("The parameter already exists. You can't create duplicate parameters.")
	ErrInternalError            = errors.New("We encountered an internal error, please try again.")
	ErrInvalidKeyId             = errors.New("The ParameterData KeyId is not valid.")
	ErrInvalidName              = errors.New("The ParameterData Name is not valid.")
	ErrInvalidTier              = errors.New("The ParameterData Tier is not valid.")
	ErrInvalidDataType          = errors.New("The ParameterData DataType is not valid.")
	ErrInvalidFilterKey         = errors.New("The specified key isn't valid.")
	ErrInvalidFilterOption      = errors.New("The specified filter option isn't valid. Valid options are Equals and BeginsWith. For Path filter, valid options are Recursive and OneLevel.")
	ErrInvalidFilterValue       = errors.New("The filter value isn't valid. Verify the value and try again.")
	ErrUnsupportedParameterType = errors.New("The parameter type isn't supported.")
	ErrInvalidPath              = errors.New("The parameter doesn't meet the parameter name requirements. The parameter name must begin with a forward slash '/'.")
)

type errorCodeMap map[error]awslib.APIError

var SsmErrorCodes = errorCodeMap{
	ErrInternalError: {
		Code:           "InternalError",
		Description:    ErrInternalError.Error(),
		HTTPStatusCode: http.StatusInternalServerError,
	},
	ErrParameterNotFound: {
		Code:           "ParameterNotFound",
		Description:    ErrParameterNotFound.Error(),
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrParameterAlreadyExists: {
		Code:           "ParameterAlreadyExists",
		Description:    ErrParameterAlreadyExists.Error(),
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrUnsupportedParameterType: {
		Code:           "UnsupportedParameterType",
		Description:    ErrUnsupportedParameterType.Error(),
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrInvalidKeyId: {
		Code:           "InvalidKeyId",
		Description:    ErrInternalError.Error(),
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrInvalidName: {
		Code:           "InvalidParameterName",
		Description:    ErrInvalidName.Error(),
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrInvalidDataType: {
		Code:           "InvalidDataType",
		Description:    ErrInvalidDataType.Error(),
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrInvalidTier: {
		Code:           "InvalidParameterTier",
		Description:    ErrInvalidTier.Error(),
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrInvalidFilterKey: {
		Code:           "InvalidParameterTier",
		Description:    ErrInvalidFilterKey.Error(),
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrInvalidFilterOption: {
		Code:           "InvalidFilterOption",
		Description:    ErrInvalidFilterOption.Error(),
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrInvalidFilterValue: {
		Code:           "InvalidFilterValue",
		Description:    ErrInvalidFilterValue.Error(),
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrInvalidPath: {
		Code:           "ValidationException",
		Description:    ErrInvalidPath.Error(),
		HTTPStatusCode: http.StatusBadRequest,
	},
}

func translateToApiError(err error) awslib.APIError {

	value, ok := SsmErrorCodes[err]
	if ok {

		return value
	}

	return awslib.ErrorCodes[awslib.ErrInternalError]
}
