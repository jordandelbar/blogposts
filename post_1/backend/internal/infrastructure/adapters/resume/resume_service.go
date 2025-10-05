package resume

import (
	"context"
	"io"
	"personal_website/config"
	"personal_website/internal/app/core/domain"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type resumeService struct {
	client *minio.Client
	config *config.MinioConfig
}

func NewService(cfg *config.MinioConfig) (*resumeService, error) {
	endpoint := cfg.Endpoint.String()
	accessKey := cfg.AccessKey.String()
	secretKey := cfg.SecretKey.String()
	useSSL := cfg.UseSSL

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, domain.DomainError{
			Code:       "resume_storage_unavailable",
			Message:    "resume storage service is unavailable",
			Type:       domain.ErrorTypeInternal,
			Underlying: err,
		}
	}

	return &resumeService{
		client: client,
		config: cfg,
	}, nil
}

func (s *resumeService) GetResume(ctx context.Context, objectName string) ([]byte, error) {
	bucket := s.config.Bucket.String()

	object, err := s.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, s.mapMinioError(err, "get object")
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, domain.DomainError{
			Code:       "resume_read_failed",
			Message:    "failed to read resume file",
			Type:       domain.ErrorTypeInternal,
			Underlying: err,
		}
	}

	return data, nil
}

func (s *resumeService) CheckConnection(ctx context.Context) error {
	bucket := s.config.Bucket.String()

	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return s.mapMinioError(err, "check bucket existence")
	}

	if !exists {
		return domain.DomainError{
			Code:    "resume_storage_unavailable",
			Message: "resume storage service is unavailable",
			Type:    domain.ErrorTypeInternal,
		}
	}

	return nil
}

func (s *resumeService) mapMinioError(err error, operation string) error {
	errorStr := strings.ToLower(err.Error())

	if strings.Contains(errorStr, "no such key") ||
	   strings.Contains(errorStr, "not found") ||
	   strings.Contains(errorStr, "nosuchkey") {
		return domain.DomainError{
			Code:       "resume_not_found",
			Message:    "resume file not found",
			Type:       domain.ErrorTypeNotFound,
			Underlying: err,
		}
	}

	if strings.Contains(errorStr, "connection") ||
	   strings.Contains(errorStr, "timeout") ||
	   strings.Contains(errorStr, "network") ||
	   strings.Contains(errorStr, "dns") {
		return domain.DomainError{
			Code:       "resume_storage_unavailable",
			Message:    "resume storage service is unavailable",
			Type:       domain.ErrorTypeInternal,
			Underlying: err,
		}
	}

	return domain.DomainError{
		Code:       "resume_read_failed",
		Message:    "failed to read resume file",
		Type:       domain.ErrorTypeInternal,
		Underlying: err,
	}
}
