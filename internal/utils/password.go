package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memory      = 64 * 1024
	iterations  = 3
	parallelism = 8
	saltLength  = 16
	keyLength   = 32
)

type argon2Params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
	salt        []byte
}

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
	ErrNoRandom            = errors.New("no random source available")
)

func HashPassword(password string, salt string) string {
	var params = argon2Params{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		saltLength:  saltLength,
		keyLength:   keyLength,
		salt:        make([]byte, saltLength),
	}
	// Fill the salt with cryptographically secure random bytes.
	_, err := rand.Read(params.salt)
	if err != nil {
		slog.Error("HashPassword: Unable to read random bytes", "error", err)
	}

	bytes := argon2.IDKey([]byte(password+salt), params.salt, params.iterations, params.memory, params.parallelism, params.keyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(params.salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(bytes)

	// Return a string using the standard encoded hash representation.
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.memory, params.iterations, params.parallelism, b64Salt, b64Hash)
}

func VerifyPassword(password, compareHash string, pwsalt string) (bool, error) {
	vals := strings.Split(compareHash, "$")
	const argon2Vals = 6
	if len(vals) != argon2Vals {
		return false, ErrInvalidHash
	}

	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return false, ErrInvalidHash
	}
	if version != argon2.Version {
		return false, ErrIncompatibleVersion
	}

	p := &argon2Params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return false, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return false, ErrInvalidHash
	}

	saltLen := len(salt)
	if saltLen < 0 || saltLen > math.MaxUint32 {
		return false, ErrInvalidHash
	}
	p.saltLength = uint32(saltLen)

	hash, err := base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return false, ErrInvalidHash
	}

	hashLen := len(hash)
	if hashLen < 0 || hashLen > math.MaxUint32 {
		return false, ErrInvalidHash
	}
	p.keyLength = uint32(hashLen)

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey([]byte(password+pwsalt), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

const allowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const allowedNumbers = "0123456789"
const allowedSpecial = "!@#$%^&*-_"

func RandomPassword(length int, minNumbers, minSpecial int) (string, error) {
	b := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(allowedChars))))
		if err != nil {
			return "", ErrNoRandom
		}
		b[i] = allowedChars[num.Int64()]
	}

	for i := 0; i < minNumbers; i++ {
		randInt, err := rand.Int(rand.Reader, big.NewInt(int64(length)))
		if err != nil {
			return "", ErrNoRandom
		}

		rollInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(allowedNumbers))))
		if err != nil {
			return "", ErrNoRandom
		}

		b[randInt.Int64()] = allowedNumbers[rollInt.Int64()]
	}
	for i := 0; i < minSpecial; i++ {
		randInt, err := rand.Int(rand.Reader, big.NewInt(int64(length)))
		if err != nil {
			return "", ErrNoRandom
		}

		rollInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(allowedSpecial))))
		if err != nil {
			return "", ErrNoRandom
		}

		b[randInt.Int64()] = allowedSpecial[rollInt.Int64()]
	}
	return string(b), nil
}
