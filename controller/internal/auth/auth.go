package auth

type AuthService struct {
	apiKey string
}

func New(apiKey string) *AuthService {
	return &AuthService{apiKey: apiKey}
}

func (a *AuthService) Validate(key string) bool {
	if a.apiKey == "" || key == "" {
		return false
	}
	return key == a.apiKey
}

// secret returns the raw key; used internally by the HMAC middleware.
func (a *AuthService) secret() string { return a.apiKey }
