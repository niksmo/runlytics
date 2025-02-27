package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReportFlag(t *testing.T) {

	t.Run("Valid report cmd arg", func(t *testing.T) {
		actual := getReportFlag(1)
		expected := float64(1)
		assert.Equal(t, expected, actual.Seconds())
	})

	t.Run("Valid report env arg", func(t *testing.T) {
		expected := float64(5)
		os.Setenv(reportEnv, "5")
		actual := getReportFlag(77)
		assert.Equal(t, expected, actual.Seconds())

	})

	t.Run("Invalid report env arg, cmd is valid", func(t *testing.T) {
		expected := float64(5)
		os.Setenv(reportEnv, "77,0")
		actual := getReportFlag(5)
		assert.Equal(t, expected, actual.Seconds())
	})

	t.Run("Invalid report env and cmd args, should return default", func(t *testing.T) {
		expected := float64(reportDefault)
		os.Setenv(reportEnv, "77,0")
		actual := getReportFlag(-99)
		assert.Equal(t, expected, actual.Seconds())
	})
}
