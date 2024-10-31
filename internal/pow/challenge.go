package pow

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"math"
	"math/big"
	"strconv"
)

const (
	delimiter = ":"
	base      = 10 // base system to format int64
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrInvalidNonce     = errors.New("invalid nonce")

	one = big.NewInt(1) //nolint:gochecknoglobals //read-only used
)

type Challenge struct {
	targetInt     *big.Int
	Signature     []byte `json:"sig"`
	Rand          []byte `json:"rand"`
	UnixTimestamp int64  `json:"unix"`
	Nonce         int64  `json:"nonce"`
	Difficulty    uint8  `json:"dif"`
}

// NewChallenge method create new challenge struct
// 0 difficulty mean POW is disabled.
func NewChallenge(rand []byte, difficulty uint8, unixTimestamp int64, secret []byte) *Challenge {
	c := &Challenge{Rand: rand, Difficulty: difficulty, UnixTimestamp: unixTimestamp}
	c.Signature = c.sig(secret)

	return c
}

func (c *Challenge) VerifySign(secret []byte) error {
	if !bytes.Equal(c.Signature, c.sig(secret)) {
		return fmt.Errorf("%w: %x", ErrInvalidSignature, c.Signature)
	}

	return nil
}

// ComputeNonce method calculates and sets a valid Nonce according to the established Difficulty.
func (c *Challenge) ComputeNonce(ctx context.Context) error {
	if c.Difficulty == 0 {
		return nil
	}

	var (
		sha   = sha256.New()
		nonce = big.NewInt(c.Nonce)
	)

	defer func() {
		c.Nonce = nonce.Int64()
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if c.verifyNonce(sha, nonce) {
				return nil
			}

			sha.Reset()

			nonce.Add(nonce, one)
		}
	}
}

func (c *Challenge) VerifyNonce() error {
	if !c.verifyNonce(sha256.New(), big.NewInt(c.Nonce)) {
		return fmt.Errorf("%w: %d", ErrInvalidNonce, c.Nonce)
	}

	return nil
}

func (c *Challenge) verifyNonce(sha hash.Hash, nonceInBytes *big.Int) bool {
	defer sha.Reset()

	sha.Write(c.Rand)
	sha.Write(nonceInBytes.Bytes())

	var hashInt big.Int

	hashInt.SetBytes(sha.Sum(nil))

	return hashInt.Cmp(c.target()) == -1
}

func (c *Challenge) buildHashingBytes() []byte {
	buf := bytes.Buffer{}
	buf.Write(c.Rand)
	buf.WriteString(delimiter)
	buf.WriteString(strconv.FormatInt(int64(c.Difficulty), base))
	buf.WriteString(delimiter)
	buf.WriteString(strconv.FormatInt(c.UnixTimestamp, base))

	return buf.Bytes()
}

func (c *Challenge) sig(secret []byte) []byte {
	b := c.buildHashingBytes()

	sha := hmac.New(sha256.New, secret)

	sha.Write(b)

	return sha.Sum(nil)
}

func (c *Challenge) target() *big.Int {
	if c.targetInt == nil {
		c.targetInt = big.NewInt(1)
		c.targetInt.Lsh(c.targetInt, math.MaxUint8-uint(c.Difficulty))
	}

	return c.targetInt
}
