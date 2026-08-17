package apperr

import "fmt"

type Error struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(status int, code, message string) *Error {
	return &Error{HTTPStatus: status, Code: code, Message: message}
}

var (
	Validation       = New(400, "VALIDATION_ERROR", "Invalid request")
	Unauthorized     = New(401, "UNAUTHORIZED", "Unauthorized")
	Forbidden        = New(403, "FORBIDDEN", "Forbidden")
	RegisterDisabled = New(403, "REGISTER_DISABLED", "Registration is disabled")
	EmailNotVerified = New(403, "EMAIL_NOT_VERIFIED", "Email not verified")
	NotFound         = New(404, "NOT_FOUND", "Not found")
	Conflict         = New(409, "CONFLICT", "Conflict")
	InvalidState     = New(409, "INVALID_STATE", "Invalid state")
	QuotaExceeded    = New(409, "QUOTA_EXCEEDED", "Storage quota exceeded")
	FileTooLarge     = New(413, "FILE_TOO_LARGE", "File too large")
	UploadExpired    = New(409, "UPLOAD_EXPIRED", "Upload session expired")
)
