package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	ErrObjectNotFound     = errors.New("objeto não encontrado no storage")
	ErrStorageUnavailable = errors.New("serviço de storage indisponível ou não configurado")
)

// StorageService define a interface para interagir com o Cloudflare R2 / S3.
type StorageService interface {
	GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, contentLength int64, expiresIn time.Duration) (string, error)
	GeneratePresignedDownloadURL(ctx context.Context, key string, filename string, download bool, expiresIn time.Duration) (string, error)
	ValidatePDFOCR(ctx context.Context, key string, minChars int) (bool, int, error)
	DeleteObject(ctx context.Context, key string) error
	ObjectExists(ctx context.Context, key string) (bool, error)
}

// Config contém as credenciais e parâmetros de conexão para o Cloudflare R2.
type Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicDomain    string
}

// LoadConfigFromEnv carrega a configuração a partir de variáveis de ambiente.
func LoadConfigFromEnv() Config {
	return Config{
		AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		BucketName:      os.Getenv("R2_BUCKET_NAME"),
		PublicDomain:    os.Getenv("R2_PUBLIC_DOMAIN"),
	}
}

// IsConfigured verifica se as credenciais mínimas do R2 estão presentes.
func (c Config) IsConfigured() bool {
	return c.AccountID != "" && c.AccessKeyID != "" && c.SecretAccessKey != "" && c.BucketName != ""
}

// R2StorageService implementa StorageService usando a API S3 compatível do Cloudflare R2.
type R2StorageService struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucketName    string
}

// NewR2StorageService instancia um novo serviço de storage conectado ao Cloudflare R2.
func NewR2StorageService(cfg Config) (*R2StorageService, error) {
	if !cfg.IsConfigured() {
		return nil, errors.New("credenciais do Cloudflare R2 incompletas (requer R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY e R2_BUCKET_NAME)")
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			SigningRegion:     "auto",
			HostnameImmutable: true,
		}, nil
	})

	awsCfg := aws.Config{
		Region:                      "auto",
		Credentials:                 credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		EndpointResolverWithOptions: customResolver,
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	presignClient := s3.NewPresignClient(s3Client)

	return &R2StorageService{
		client:        s3Client,
		presignClient: presignClient,
		bucketName:    cfg.BucketName,
	}, nil
}

// GeneratePresignedUploadURL gera uma URL pré-assinada para upload direto via HTTP PUT com Content-Length assinado.
func (s *R2StorageService) GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, contentLength int64, expiresIn time.Duration) (string, error) {
	putInput := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(contentLength),
	}

	req, err := s.presignClient.PresignPutObject(ctx, putInput, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return "", fmt.Errorf("falha ao gerar presigned upload url: %w", err)
	}

	return req.URL, nil
}

// GeneratePresignedDownloadURL gera uma URL pré-assinada para download ou preview inline do arquivo.
func (s *R2StorageService) GeneratePresignedDownloadURL(ctx context.Context, key string, filename string, download bool, expiresIn time.Duration) (string, error) {
	getInput := &s3.GetObjectInput{
		Bucket:              aws.String(s.bucketName),
		Key:                 aws.String(key),
		ResponseContentType: aws.String("application/pdf"),
	}

	if filename != "" {
		dispositionType := "inline"
		if download {
			dispositionType = "attachment"
		}
		cleanFilename := strings.ReplaceAll(filename, "\"", "")
		disposition := fmt.Sprintf("%s; filename=\"%s\"", dispositionType, cleanFilename)
		getInput.ResponseContentDisposition = aws.String(disposition)
	}

	req, err := s.presignClient.PresignGetObject(ctx, getInput, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return "", fmt.Errorf("falha ao gerar presigned download url: %w", err)
	}

	return req.URL, nil
}

// ValidatePDFOCR baixa os bytes do arquivo no R2 e valida se possui texto pesquisável (OCR),
// impondo limite estrito de 25MB para proteger o container contra DoS / OOM.
func (s *R2StorageService) ValidatePDFOCR(ctx context.Context, key string, minChars int) (bool, int, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, 0, fmt.Errorf("erro ao recuperar objeto do R2 para validação de OCR: %w", err)
	}
	defer out.Body.Close()

	const maxAllowedBytes = int64(25 * 1024 * 1024) // 25 MB
	if out.ContentLength != nil && *out.ContentLength > maxAllowedBytes {
		return false, 0, errors.New("o arquivo excede o limite máximo permitido de 25 MB")
	}

	buf, err := io.ReadAll(io.LimitReader(out.Body, maxAllowedBytes+1))
	if err != nil {
		return false, 0, fmt.Errorf("erro ao ler corpo do arquivo: %w", err)
	}

	if int64(len(buf)) > maxAllowedBytes {
		return false, 0, errors.New("o arquivo excede o limite máximo permitido de 25 MB")
	}

	if !IsPDFHeader(buf) {
		return false, 0, errors.New("o arquivo armazenado não possui assinatura válida de PDF")
	}

	reader := bytes.NewReader(buf)
	return ValidatePDFText(reader, int64(len(buf)), minChars)
}

// DeleteObject remove um objeto do Cloudflare R2.
func (s *R2StorageService) DeleteObject(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("falha ao excluir objeto do R2 (%s): %w", key, err)
	}

	return nil
}

// ObjectExists verifica se um objeto existe no bucket.
func (s *R2StorageService) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		var responseError interface{ HTTPStatusCode() int }
		if errors.As(err, &responseError) && responseError.HTTPStatusCode() == http.StatusNotFound {
			return false, nil
		}
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
