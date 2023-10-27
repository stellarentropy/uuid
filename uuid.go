package uuid

import (
	"crypto/rand"
	"encoding/binary"
	"errors"

	"github.com/stellarentropy/isaac64"
)

var (
	errInvalid = errors.New("invalid UUID")

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

// Stringfy into the given byte buffer.
func (u UUID) hexify(buf []byte) {
	buf[8] = '-'
	buf[13] = '-'
	buf[18] = '-'
	buf[23] = '-'
	for i, j := range encode {
		const hex = "0123456789abcdef"
		buf[j+0] = hex[u[i]>>4]
		buf[j+1] = hex[u[i]&0x0f]
	}
}

func (u UUID) String() string {
	var buf [36]byte
	u.hexify(buf[:])
	return string(buf[:])
}

func (u UUID) MarshalBinary() ([]byte, error) {
	return u[:], nil
}

func (u *UUID) UnmarshalBinary(data []byte) error {
	if len(data) < len(u) {
		return errors.New("invalid length")
	}
	copy(u[:], data[:])
	return nil
}

func (u UUID) MarshalJSON() ([]byte, error) {
	var buf [38]byte
	buf[0] = '"'
	u.hexify(buf[1:])
	buf[37] = '"'
	return buf[:], nil
}

func (u *UUID) UnmarshalJSON(buf []byte) error {
	if len(buf) != 38 || buf[0] != '"' || buf[37] != '"' {
		return errInvalid
	}
	uuid, err := ParseBytes(buf[1:37])
	if err != nil {
		return err
	}
	*u = uuid
	return nil
}

// ParseBytes is like Parse but accepts a byte slice.
func ParseBytes(buf []byte) (UUID, error) {
	var u UUID
	if len(buf) != 36 {
		return u, errInvalid
	}
	if buf[8] != '-' || buf[13] != '-' || buf[18] != '-' || buf[23] != '-' {
		return u, errInvalid
	}
	for i, j := range encode {
		hi := nibbles[buf[j+0]]
		lo := nibbles[buf[j+1]]
		if hi == 0xff || lo == 0xff {
			return u, errInvalid
		}
		u[i] = hi<<4 | lo
	}
	return u, nil
}

// Parse decodes a UUID (any version) from its string representation.
func Parse(s string) (UUID, error) {
	return ParseBytes([]byte(s))
}

// MustParse is like Parse but panics if the string is invalid.
func MustParse(s string) UUID {
	uuid, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return uuid
}

// Gen is a version 4 UUID generator backed by a CSPRNG.
type Gen isaac64.Rand

// NewGen initializes and returns a new version 4 UUID generator.
func NewGen() *Gen {
	r := isaac64.New()
	if err := r.SeedFrom(rand.Reader); err != nil {
		panic(err)
	}
	return (*Gen)(r)
}

// NewV4 returns a fresh version 4 UUID.
func (g *Gen) NewV4() UUID {
	var u UUID
	r := (*isaac64.Rand)(g)
	binary.LittleEndian.PutUint64(u[0:], r.Uint64())
	binary.LittleEndian.PutUint64(u[8:], r.Uint64())
	u[6] = (u[6] & 0x0f) | 0x40
	u[8] = (u[8] & 0x3f) | 0x80
	return u
}
