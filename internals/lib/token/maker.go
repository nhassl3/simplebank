package token

// Maker is an interface for managing tokens
type Maker interface {
	CreateToken(username string, isAdmin bool) (string, error)
	VerifyToken(token string) (*Payload, error)
}
