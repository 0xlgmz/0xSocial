package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

var ErrObjectNotFound = errors.New("media object not found")

type ObjectInfo struct {
	ContentType string
	SizeBytes   int64
}

type UploadGrant struct {
	URL       string
	ExpiresAt time.Time
}

type Store interface {
	PresignUpload(context.Context, string, string, time.Duration) (UploadGrant, error)
	HeadObject(context.Context, string) (ObjectInfo, error)
	DeleteObject(context.Context, string) error
	PublicURL(string) string
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	PublicBaseURL   string
	Endpoint        string
}

type R2Store struct {
	client        *s3.Client
	presigner     *s3.PresignClient
	bucket        string
	publicBaseURL string
}

func R2ConfigFromEnv() (R2Config, bool, error) {
	configuration := R2Config{
		AccountID:       strings.TrimSpace(os.Getenv("R2_ACCOUNT_ID")),
		AccessKeyID:     strings.TrimSpace(os.Getenv("R2_ACCESS_KEY_ID")),
		SecretAccessKey: strings.TrimSpace(os.Getenv("R2_SECRET_ACCESS_KEY")),
		Bucket:          strings.TrimSpace(os.Getenv("R2_BUCKET")),
		PublicBaseURL:   strings.TrimRight(strings.TrimSpace(os.Getenv("R2_PUBLIC_BASE_URL")), "/"),
		Endpoint:        strings.TrimRight(strings.TrimSpace(os.Getenv("R2_ENDPOINT")), "/"),
	}

	if configuration.AccountID == "" &&
		configuration.AccessKeyID == "" &&
		configuration.SecretAccessKey == "" &&
		configuration.Bucket == "" &&
		configuration.PublicBaseURL == "" &&
		configuration.Endpoint == "" {
		return R2Config{}, false, nil
	}

	missing := make([]string, 0)
	for name, value := range map[string]string{
		"R2_ACCESS_KEY_ID":     configuration.AccessKeyID,
		"R2_SECRET_ACCESS_KEY": configuration.SecretAccessKey,
		"R2_BUCKET":            configuration.Bucket,
		"R2_PUBLIC_BASE_URL":   configuration.PublicBaseURL,
	} {
		if value == "" {
			missing = append(missing, name)
		}
	}

	if configuration.Endpoint == "" {
		if configuration.AccountID == "" {
			missing = append(missing, "R2_ACCOUNT_ID")
		} else {
			configuration.Endpoint = fmt.Sprintf(
				"https://%s.r2.cloudflarestorage.com",
				configuration.AccountID,
			)
		}
	}

	if len(missing) > 0 {
		return R2Config{}, false, fmt.Errorf("incomplete R2 configuration: missing %s", strings.Join(missing, ", "))
	}

	publicURL, err := url.Parse(configuration.PublicBaseURL)
	if err != nil || publicURL.Scheme == "" || publicURL.Host == "" {
		return R2Config{}, false, errors.New("R2_PUBLIC_BASE_URL must be an absolute URL")
	}

	endpoint, err := url.Parse(configuration.Endpoint)
	if err != nil || endpoint.Scheme == "" || endpoint.Host == "" {
		return R2Config{}, false, errors.New("R2_ENDPOINT must be an absolute URL")
	}

	return configuration, true, nil
}

func NewR2Store(ctx context.Context, configuration R2Config) (*R2Store, error) {
	awsConfiguration, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			configuration.AccessKeyID,
			configuration.SecretAccessKey,
			"",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("load R2 configuration: %w", err)
	}

	client := s3.NewFromConfig(awsConfiguration, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(configuration.Endpoint)
	})

	return &R2Store{
		client:        client,
		presigner:     s3.NewPresignClient(client),
		bucket:        configuration.Bucket,
		publicBaseURL: configuration.PublicBaseURL,
	}, nil
}

func (store *R2Store) PresignUpload(ctx context.Context, objectKey, contentType string, lifetime time.Duration) (UploadGrant, error) {
	result, err := store.presigner.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      aws.String(store.bucket),
			Key:         aws.String(objectKey),
			ContentType: aws.String(contentType),
			IfNoneMatch: aws.String("*"),
		},
		func(options *s3.PresignOptions) {
			options.Expires = lifetime
		},
	)
	if err != nil {
		return UploadGrant{}, fmt.Errorf("presign R2 upload: %w", err)
	}

	return UploadGrant{
		URL:       result.URL,
		ExpiresAt: time.Now().Add(lifetime),
	}, nil
}

func (store *R2Store) HeadObject(ctx context.Context, objectKey string) (ObjectInfo, error) {
	result, err := store.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(store.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) &&
			(apiError.ErrorCode() == "NotFound" || apiError.ErrorCode() == "NoSuchKey") {
			return ObjectInfo{}, ErrObjectNotFound
		}

		return ObjectInfo{}, fmt.Errorf("inspect R2 object: %w", err)
	}

	return ObjectInfo{
		ContentType: aws.ToString(result.ContentType),
		SizeBytes:   aws.ToInt64(result.ContentLength),
	}, nil
}

func (store *R2Store) DeleteObject(ctx context.Context, objectKey string) error {
	_, err := store.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(store.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("delete R2 object: %w", err)
	}

	return nil
}

// PutObjectIfAbsent uploads an object for internal tooling such as the data
// seeder. It never overwrites existing user content.
func (store *R2Store) PutObjectIfAbsent(ctx context.Context, objectKey, contentType string, body []byte) (bool, error) {
	_, err := store.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(store.bucket),
		Key:           aws.String(objectKey),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(int64(len(body))),
		IfNoneMatch:   aws.String("*"),
		Body:          bytes.NewReader(body),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) &&
			(apiError.ErrorCode() == "PreconditionFailed" || apiError.ErrorCode() == "ConditionalRequestConflict") {
			return false, nil
		}
		return false, fmt.Errorf("upload R2 object: %w", err)
	}
	return true, nil
}

func (store *R2Store) PublicURL(objectKey string) string {
	return store.publicBaseURL + "/" + strings.TrimLeft(objectKey, "/")
}
