package uuid

import (
	"crypto/rand"
	"errors"

	"github.com/skeeto/chacha-go"
)

var (
	invalidErr = errors.New("invalid UUID")

	// Maps UUID bytes into their string representation indices
	encode = [...]int{
		0, 2, 4, 6, 9, 11, 14, 16, 19, 21, 24, 26, 28, 30, 32, 34,
	}

	// Maps hexidecimal to nibbles
	nibbles = [...]byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	}
)

// UUID represents a UUID (any version). This is explicitly an array so
// that users can directly slice the binary representation.
type UUID [16]byte

func (u UUID) String() string {
	const hex = "0123456789abcdef"
	buf := make([]byte, 36)
	buf[8] = '-'
	buf[13] = '-'
	buf[18] = '-'
	buf[23] = '-'
	for i, j := range encode {
		buf[j+0] = hex[u[i]>>4]
		buf[j+1] = hex[u[i]&0x0f]
	}
	return string(buf)
}

func (u UUID) MarshalBinary() ([]byte, error) {
	return u[:], nil
}

func (u UUID) UnmarshalBinary(data []byte) error {
	if len(data) < len(u) {
		return errors.New("invalid length")
	}
	copy(u[:], data[:])
	return nil
}

// Parse a UUID (any version) from its string representation.
func Parse(s string) (UUID, error) {
	var u UUID
	if len(s) != 36 {
		return u, invalidErr
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return u, invalidErr
	}
	for i, j := range encode {
		hi := nibbles[s[j+0]]
		lo := nibbles[s[j+1]]
		if hi == 0xff || lo == 0xff {
			return u, invalidErr
		}
		u[i] = hi<<4 | lo
	}
	return u, nil
}

// Gen is a version 4 UUID generator backed by a CSPRNG.
type Gen struct {
	state *chacha.Cipher
}

// NewGen initializes and returns a new version 4 UUID generator.
func NewGen() (Gen, error) {
	var seed [40]byte
	if _, err := rand.Read(seed[:]); err != nil {
		return Gen{}, err
	}
	return Gen{chacha.New(seed[:32], seed[32:], 12)}, nil
}

// NewV4 returns a fresh version 4 UUID.
func (g Gen) NewV4() UUID {
	var u UUID
	// Technically this will EOF when the keystream is exhausted, but in
	// practice this will take so long (> 250,000 years) that it will
	// never happen.
	g.state.Read(u[:])
	u[6] = (u[6] & 0x0f) | 0x40
	u[8] = (u[8] & 0x3f) | 0x80
	return u
}
