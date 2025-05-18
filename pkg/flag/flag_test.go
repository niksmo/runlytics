package flag_test

import (
	"os"
	"testing"

	"github.com/niksmo/runlytics/pkg/flag"
	"github.com/stretchr/testify/assert"
)

func TestFlagSet(t *testing.T) {
	const (
		STRING_FLAG = "sf"
		INT_FLAG    = "if"
		BOOL_FLAG   = "bf"

		STRING_VALUE = "testStringValue"
		INT_VALUE    = "12345"
		BOOL_VALUE   = "true"

		EMPTY_USAGE = ""
	)
	t.Run("Regular should no errors", func(t *testing.T) {
		args := []string{
			"-" + STRING_FLAG, STRING_VALUE,
			"-" + INT_FLAG, INT_VALUE,
			"-" + BOOL_FLAG, BOOL_VALUE,
		}
		os.Args = append(os.Args[:1], args...)
		flagSet := flag.New()
		flagSet.String(STRING_FLAG, "", EMPTY_USAGE)
		flagSet.Int(INT_FLAG, 0, EMPTY_USAGE)
		flagSet.Bool(BOOL_FLAG, false, EMPTY_USAGE)
		flagSet.Parse()
		assert.True(t, flagSet.IsSet(STRING_FLAG))
		assert.True(t, flagSet.IsSet(INT_FLAG))
		assert.True(t, flagSet.IsSet(BOOL_FLAG))
	})

	t.Run("One flag not passed to arg", func(t *testing.T) {
		args := []string{
			"-" + INT_FLAG, INT_VALUE,
			"-" + BOOL_FLAG, BOOL_VALUE,
		}
		os.Args = append(os.Args[:1], args...)
		flagSet := flag.New()
		flagSet.String(STRING_FLAG, "", EMPTY_USAGE) // not set in args
		flagSet.Int(INT_FLAG, 0, EMPTY_USAGE)
		flagSet.Bool(BOOL_FLAG, false, EMPTY_USAGE)
		flagSet.Parse()
		assert.False(t, flagSet.IsSet(STRING_FLAG))
		assert.True(t, flagSet.IsSet(INT_FLAG))
		assert.True(t, flagSet.IsSet(BOOL_FLAG))
	})

}
