package bits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(Read(byte(0b1011), 2), byte(0))
	assert.Equal(Read(byte(0b1000), 3), byte(1))
}
