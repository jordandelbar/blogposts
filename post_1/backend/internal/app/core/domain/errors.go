package domain

type DomainError struct {
	Code       string
	Message    string
	Type       ErrorType
	Underlying error
}

type ErrorType string

func (e DomainError) Error() string {
	return e.Message
}

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeConflict   ErrorType = "conflict"
	ErrorTypeAuth       ErrorType = "authentication"
	ErrorTypeInternal   ErrorType = "internal"
)

const (
	InvalidCredentialsErrorMsg string = "invalid credentials"
)

var (
	ErrSessionNotFound = DomainError{
		Code:    "session_not_found",
		Message: "session not found",
		Type:    ErrorTypeNotFound,
	}
	ErrSessionExpired = DomainError{
		Code:    "session_expired",
		Message: "session has expired",
		Type:    ErrorTypeAuth,
	}
	ErrUserAlreadyExists = DomainError{
		Code:    "user_already_exists",
		Message: "user already exists",
		Type:    ErrorTypeConflict,
	}
	ErrUserNotFound = DomainError{
		Code:    "user_not_found",
		Message: "user not found",
		Type:    ErrorTypeNotFound,
	}
	ErrArticleNotFound = DomainError{
		Code:    "article_not_found",
		Message: "article not found",
		Type:    ErrorTypeNotFound,
	}
	ErrArticleAlreadyExists = DomainError{
		Code:    "article_already_exists",
		Message: "article already exists",
		Type:    ErrorTypeConflict,
	}
	ErrArticleNotSoftDeleted = DomainError{
		Code:    "article_not_soft_deleted",
		Message: "article must be soft deleted before permanent deletion",
		Type:    ErrorTypeConflict,
	}
	ErrInvalidCredentials = DomainError{
		Code:    "invalid_credentials",
		Message: InvalidCredentialsErrorMsg,
		Type:    ErrorTypeAuth,
	}
	ErrEmailConfigurationMissing = DomainError{
		Code:    "email_configuration_missing",
		Message: "email service configuration is incomplete",
		Type:    ErrorTypeInternal,
	}
	ErrEmailTemplateFailed = DomainError{
		Code:    "email_template_failed",
		Message: "failed to process email template",
		Type:    ErrorTypeInternal,
	}
	ErrEmailSendFailed = DomainError{
		Code:    "email_send_failed",
		Message: "failed to send email",
		Type:    ErrorTypeInternal,
	}
	ErrResumeNotFound = DomainError{
		Code:    "resume_not_found",
		Message: "resume file not found",
		Type:    ErrorTypeNotFound,
	}
	ErrResumeStorageUnavailable = DomainError{
		Code:    "resume_storage_unavailable",
		Message: "resume storage service is unavailable",
		Type:    ErrorTypeInternal,
	}
	ErrResumeReadFailed = DomainError{
		Code:    "resume_read_failed",
		Message: "failed to read resume file",
		Type:    ErrorTypeInternal,
	}
	ErrInternal = DomainError{
		Code:    "internal_error",
		Message: "an internal error occurred",
		Type:    ErrorTypeInternal,
	}
	ErrInvalidAuthToken = DomainError{
		Code:    "invalid_auth_token",
		Message: "invalid or missing authentication token",
		Type:    ErrorTypeAuth,
	}
	ErrAuthenticationRequired = DomainError{
		Code:    "authentication_required",
		Message: "you must be authenticated to access this resource",
		Type:    ErrorTypeAuth,
	}
	ErrInactiveAccount = DomainError{
		Code:    "inactive_account",
		Message: "your user account must be activated to access this resource",
		Type:    ErrorTypeAuth,
	}
	ErrNotPermitted = DomainError{
		Code:    "not_permitted",
		Message: "your user account doesn't have the necessary permissions to access this resource",
		Type:    ErrorTypeAuth,
	}
)

func NewInternalError(underlying error) DomainError {
	return DomainError{
		Code:       "internal_error",
		Message:    "an internal error occurred",
		Type:       ErrorTypeInternal,
		Underlying: underlying,
	}
}
