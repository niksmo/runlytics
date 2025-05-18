package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type envValue interface {
	Set(string) error
}

type stringValue string

func (s *stringValue) Set(v string) error {
	*s = stringValue(v)
	return nil
}

type intValue int

func (i *intValue) Set(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*i = intValue(v)
	return nil
}

type boolValue bool

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	*b = boolValue(v)
	return nil
}

type parseError []error

func (pe parseError) Error() string {
	var b strings.Builder
	li := len(pe) - 1
	for i, e := range pe {
		b.WriteString(e.Error())
		if i != li {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

type EnvSet struct {
	vMap map[string]envValue
	vSet map[string]struct{}
}

func New() *EnvSet {
	return &EnvSet{}
}

func (es *EnvSet) Parse() error {
	var errs parseError
	for n, v := range es.vMap {
		s, ok := os.LookupEnv(n)
		if !ok {
			continue
		}

		err := v.Set(s)
		if err != nil {
			err = fmt.Errorf(
				"invalid value %q for env variable %s: %w", s, n, err,
			)
			errs = append(errs, err)
			continue
		}
		if es.vSet == nil {
			es.vSet = make(map[string]struct{})
		}
		es.vSet[n] = struct{}{}
	}

	if errs != nil {
		return errs
	}
	return nil
}

func (es *EnvSet) IsSet(name string) bool {
	_, isSet := es.vSet[name]
	return isSet
}

func (es *EnvSet) String(name string) *string {
	p := new(string)
	es.registerEnv(name, (*stringValue)(p))
	return p
}

func (es *EnvSet) Int(name string) *int {
	p := new(int)
	es.registerEnv(name, (*intValue)(p))
	return p
}

func (es *EnvSet) Bool(name string) *bool {
	p := new(bool)
	es.registerEnv(name, (*boolValue)(p))
	return p
}

func (es *EnvSet) registerEnv(name string, value envValue) {
	if es.vMap == nil {
		es.vMap = make(map[string]envValue)
	}
	es.vMap[name] = value
}
