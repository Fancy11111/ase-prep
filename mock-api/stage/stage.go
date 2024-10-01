package stage

import (
	"crypto/md5"
	"encoding/binary"
	"io"
	"math/rand"
)

type Stage[T any, S any] interface {
	CreateTestcase(token string, nr int) T
	GetSolution(token string, nr int) S
	ValidateSolution(token string, nr int, solution S) bool
}

func RandFromTokenAndTestcase(token string, nr int) *rand.Rand {
	h := md5.New()
	io.WriteString(h, token)
	seed := binary.BigEndian.Uint64(h.Sum(nil))
	return rand.New(rand.NewSource(int64(seed) + int64(nr)))
}
