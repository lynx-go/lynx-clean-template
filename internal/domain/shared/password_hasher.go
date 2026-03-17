package shared

// PasswordHasher is the port for hashing and verifying passwords.
// Implementations are provided by the infrastructure layer.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}
