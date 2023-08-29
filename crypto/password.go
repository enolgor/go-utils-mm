package crypto

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

func ComparePassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

func ShouldUpdateCost(hashed string, cost int) (bool, error) {
	hashedCost, err := bcrypt.Cost([]byte(hashed))
	return hashedCost < cost || err != nil, err
}

func OptimalCost(target time.Duration) int {
	cost := bcrypt.MinCost
	start := time.Now()
	HashPassword("microbenchmark", cost)
	end := time.Now()
	dur := end.Sub(start)
	for dur < target {
		cost = cost + 1
		dur = 2 * dur
	}
	cost = cost - 1
	if cost > bcrypt.MaxCost {
		return bcrypt.MaxCost
	}
	return cost
}

type passwordHasher struct {
	cost int
}

type PasswordHasher interface {
	Cost() int
	HashPassword(password string) (string, error)
	ComparePassword(hashed, password string) error
	ShouldUpdateCost(hashed string) (bool, error)
}

func NewPasswordHasher(target time.Duration) PasswordHasher {
	return &passwordHasher{cost: OptimalCost(target)}
}

func (ph *passwordHasher) Cost() int {
	return ph.cost
}

func (ph *passwordHasher) HashPassword(password string) (string, error) {
	return HashPassword(password, ph.cost)
}

func (ph *passwordHasher) ComparePassword(hashed, password string) error {
	return ComparePassword(hashed, password)
}

func (ph *passwordHasher) ShouldUpdateCost(hashed string) (bool, error) {
	hashedCost, err := bcrypt.Cost([]byte(hashed))
	return hashedCost < ph.cost || err != nil, err
}
