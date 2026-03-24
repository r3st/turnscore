module github.com/r3st/turnscore

go 1.24

require (
	// HTTP & API
	github.com/gin-gonic/gin v1.10.0
	github.com/oapi-codegen/runtime v1.1.1

	// Database
	gorm.io/gorm v1.25.11
	gorm.io/driver/postgres v1.5.9
	github.com/golang-migrate/migrate/v4 v4.17.1

	// Configuration & CLI
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0

	// Auth
	github.com/golang-jwt/jwt/v5 v5.2.1
	golang.org/x/oauth2 v0.23.0

	// Logging
	github.com/rs/zerolog v1.33.0

	// Storage
	github.com/aws/aws-sdk-go-v2 v1.30.0
	github.com/aws/aws-sdk-go-v2/config v1.27.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.63.0

	// QR & PDF
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/jung-kurt/gofpdf v1.16.2

	// Utilities
	github.com/google/uuid v1.6.0

	// Testing
	github.com/stretchr/testify v1.9.0
	go.uber.org/mock v0.4.0
)
