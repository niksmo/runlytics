package config

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAddrFlag(t *testing.T) {
	t.Run("Valid address", func(t *testing.T) {
		type test struct {
			addr, want string
		}
		testList := []test{
			{addr: "127.0.0.1", want: "http://127.0.0.1"},
			{addr: "127.0.0.1:8080", want: "http://127.0.0.1:8080"},
			{addr: "http://some-domen.ru", want: "http://some-domen.ru"},
			{addr: "https://some-domen.ru", want: "https://some-domen.ru"},
		}
		for _, test := range testList {
			URL := getAddrFlag(test.addr)
			assert.Equal(t, test.want, URL.String())
		}
	})

	t.Run("Invalid address, should return default value", func(t *testing.T) {
		addrList := []string{
			"",
			"/127.0.0.1",
		}

		expected, _ := url.ParseRequestURI(addrDefault)
		for _, addr := range addrList {
			actualURL := getAddrFlag(addr)
			assert.Equal(t, expected, actualURL)
		}

	})

	t.Run("Should use ENV value", func(t *testing.T) {
		expected := "http://env-domen.test"
		os.Setenv(addrEnv, expected)
		URL := getAddrFlag("http://127.0.0.1:8080")
		assert.Equal(t, expected, URL.String())
	})

	t.Run("Wrong ENV value, should use cmd param", func(t *testing.T) {
		os.Setenv(addrEnv, "/wrong-env-domen.test")
		expected := "https://127.0.0.1"
		URL := getAddrFlag(expected)
		assert.Equal(t, expected, URL.String())
	})

	t.Run("Wrong ENV value, cmd param, should use default", func(t *testing.T) {
		os.Setenv(addrEnv, "/wrong-env-domen.test")
		actual := getAddrFlag("///localhost:8080")
		expected, _ := url.ParseRequestURI(addrDefault)
		assert.Equal(t, expected, actual)
	})
}
