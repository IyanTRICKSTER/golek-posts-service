package status

type PostOperationStatus int

var OperationAllowed PostOperationStatus = 223
var OperationForbidden PostOperationStatus = 224
var OperationUnauthorized PostOperationStatus = 225

var PostCreatedStatusSuccess PostOperationStatus = 111
var PostCreatedStatusFailed PostOperationStatus = 112

var PostUpdatedStatusSuccess PostOperationStatus = 221
var PostUpdatedStatusFailed PostOperationStatus = 222

var PostDeletedStatusSuccess PostOperationStatus = 331
var PostDeletedStatusFailed PostOperationStatus = 332

var PostValidateOwnerFailed PostOperationStatus = 441
var PostValidateOwnerSuccess PostOperationStatus = 442
var PostAlreadyReturned PostOperationStatus = 443
var PostRequestValidationFailed PostOperationStatus = 444
var PostRequestValidationSuccess PostOperationStatus = 445
