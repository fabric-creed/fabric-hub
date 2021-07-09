package sw

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrivateSignPublicVerify(t *testing.T) {
	msg := []byte("hello world")
	s, err := NewSimpleCSP("./test/server.key", "./test/server.crt")
	assert.Nil(t, err)
	signer, err := s.Sign(msg)
	assert.Nil(t, err)
	fmt.Println(string(signer))

	valid, err := s.Verify(signer, msg)
	assert.Nil(t, err)
	assert.True(t, valid)
}
