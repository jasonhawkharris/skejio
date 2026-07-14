package password

import "golang.org/x/crypto/bcrypt"

// Cost is the bcrypt work factor used by Hash. It's a var (rather than using
// bcrypt.DefaultCost directly) so the test harness can lower it - at
// DefaultCost, every sign-up/login in the test suite costs ~60ms of pure
// bcrypt work, which dominates total test runtime.
var Cost = bcrypt.DefaultCost

func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), Cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func Verify(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
