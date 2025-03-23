package ssm

import (
	"home-ssm/awslib"
	"net/http"
)

type SsmErrorCode int

type errorCodeMap map[SsmErrorCode]awslib.APIError

const (
	ErrNone SsmErrorCode = iota
	ErrParameterNotFound
	ErrParameterAlreadyExists
	ErrInternalError
	ErrInvalidKeyId
	ErrInvalidName
	ErrInvalidTier
	ErrUnsupportedParameterType
)

var SsmErrorCodes = errorCodeMap{
	ErrNone: {
		Code:           "",
		Description:    "No Error.",
		HTTPStatusCode: http.StatusOK,
	},
	ErrInternalError: {
		Code:           "InternalError",
		Description:    "We encountered an internal error, please try again.",
		HTTPStatusCode: http.StatusInternalServerError,
	},
	ErrParameterNotFound: {
		Code:           "ParameterNotFound",
		Description:    "The Parameter Name provided does not exist.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrParameterAlreadyExists: {
		Code:           "ParameterAlreadyExists",
		Description:    "The Parameter Name provided already exists.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrUnsupportedParameterType: {
		Code:           "UnsupportedParameterType",
		Description:    "The Parameter Type is not supported.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrInvalidKeyId: {
		Code:           "InvalidKeyId",
		Description:    "The Parameter KeyId is not valid.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrInvalidName: {
		Code:           "InvalidParameterName",
		Description:    "The Parameter Name is not valid.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrInvalidTier: {
		Code:           "InvalidParameterTier",
		Description:    "The Parameter Tier is not valid.",
		HTTPStatusCode: http.StatusBadRequest,
	},
}
