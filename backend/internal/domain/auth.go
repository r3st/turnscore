package domain

import "time"

// TokenPair holds an access token and a refresh token.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // access token lifetime in seconds
}

// Claims holds the JWT payload for authenticated users and raters.
// ⚠️ No name, no email — IDs only.
type Claims struct {
	Role         string `json:"role"`           // "user" | "rater"
	TournamentID string `json:"tid,omitempty"`  // raters only
}

// GoogleUserInfo holds the data returned by Google's UserInfo endpoint.
// Only used internally during OAuth callback — never stored in JWT.
type GoogleUserInfo struct {
	Sub     string // Google subject ID — used as stable identifier
	Email   string
	Name    string
	Picture string
}

// RaterLoginInput holds the credentials submitted by a rater.
type RaterLoginInput struct {
	TournamentSlug string
	Nickname       string
	Code           string
}

// RefreshInput holds a refresh token for re-issuance.
type RefreshInput struct {
	RefreshToken string
}

// JWTConfig holds parameters for token issuance.
type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}
