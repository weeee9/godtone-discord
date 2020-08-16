package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDir(t *testing.T) {
	assert.NoError(t, tempSave("22345", nil))
}
