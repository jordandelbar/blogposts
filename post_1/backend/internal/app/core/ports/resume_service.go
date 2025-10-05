package ports

import "context"

type ResumeService interface {
	GetResume(ctx context.Context, objectName string) ([]byte, error)
	CheckConnection(ctx context.Context) error
}
