package ssm

import (
	"errors"
	"home-ssm/awslib"
	"net/http"
)

var (
	ErrParameterNotFound        = errors.New("The Parameter Name provided does not exist.")
	ErrParameterAlreadyExists   = errors.New("The parameter already exists. You can't create duplicate parameters.")
	ErrInternalError            = errors.New("We encountered an internal error, please try again.")
	ErrInvalidKeyId             = errors.New("The Parameter KeyId is not valid.")
	ErrInvalidName              = errors.New("The Parameter Name is not valid.")
	ErrInvalidTier              = errors.New("The Parameter Tier is not valid.")
	ErrInvalidDataType          = errors.New("The Parameter DataType is not valid.")
	ErrInvalidFilterKey         = errors.New("The specified key isn't valid.")
	ErrInvalidFilterOption      = errors.New("The specified filter option isn't valid. Valid options are Equals and BeginsWith. For Path filter, valid options are Recursive and OneLevel.")
	ErrInvalidFilterValue       = errors.New("The filter value isn't valid. Verify the value and try again.")
	ErrUnsupportedParameterType = errors.New("The Parameter Type is not supported.")
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
}

func translateToApiError(err error) awslib.APIError {

	value, ok := SsmErrorCodes[err]
	if ok {

		return value
	}

	return awslib.ErrorCodes[awslib.ErrInternalError]
}
