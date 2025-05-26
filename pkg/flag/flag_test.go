package flag_test

import (
	"os"
	"testing"

	"github.com/niksmo/runlytics/pkg/flag"
	"github.com/stretchr/testify/assert"
)

func TestFlagSet(t *testing.T) {
	const (
		stringFlag = "sf"
		intFlag    = "if"
		boolFlag   = "bf"

		stringValue = "testStringValue"
		intValue    = "12345"
		boolValue   = "true"

		usage = ""
	)
	t.Run("Regular should no errors", func(t *testing.T) {
		args := []string{
			"-" + stringFlag, stringValue,
			"-" + intFlag, intValue,
			"-" + boolFlag, boolValue,
		}
		os.Args = append(os.Args[:1], args...)
		flagSet := flag.New()
		flagSet.String(stringFlag, "", usage)
		flagSet.Int(intFlag, 0, usage)
		flagSet.Bool(boolFlag, false, usage)
		flagSet.Parse()
		assert.True(t, flagSet.IsSet(stringFlag))
		assert.True(t, flagSet.IsSet(intFlag))
		assert.True(t, flagSet.IsSet(boolFlag))
	})

	t.Run("One flag not passed to arg", func(t *testing.T) {
		args := []string{
			"-" + intFlag, intValue,
			"-" + boolFlag, boolValue,
		}
		os.Args = append(os.Args[:1], args...)
		flagSet := flag.New()
		flagSet.String(stringFlag, "", usage) // not set in args
		flagSet.Int(intFlag, 0, usage)
		flagSet.Bool(boolFlag, false, usage)
		flagSet.Parse()
		assert.False(t, flagSet.IsSet(stringFlag))
		assert.True(t, flagSet.IsSet(intFlag))
		assert.True(t, flagSet.IsSet(boolFlag))
	})

}
