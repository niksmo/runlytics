package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPollFlag(t *testing.T) {

	t.Run("Valid poll cmd arg", func(t *testing.T) {
		actual := getPollFlag(1)
		expected := float64(1)
		assert.Equal(t, expected, actual.Seconds())
	})

	t.Run("Valid poll env arg", func(t *testing.T) {
		expected := float64(5)
		os.Setenv(pollEnv, "5")
		actual := getPollFlag(77)
		assert.Equal(t, expected, actual.Seconds())

	})

	t.Run("Invalid poll env arg, cmd is valid", func(t *testing.T) {
		expected := float64(5)
		os.Setenv(pollEnv, "77,0")
		actual := getPollFlag(5)
		assert.Equal(t, expected, actual.Seconds())
	})

	t.Run("Invalid poll env and cmd args, should return default", func(t *testing.T) {
		expected := float64(pollDefault)
		os.Setenv(pollEnv, "77,0")
		actual := getPollFlag(-99)
		assert.Equal(t, expected, actual.Seconds())
	})
}
