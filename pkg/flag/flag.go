package flag

import (
	"flag"
	"os"
)

type FlagSet struct {
	*flag.FlagSet
	vSet map[string]struct{}
}

func New() *FlagSet {
	return &FlagSet{
		FlagSet: flag.NewFlagSet(os.Args[0], flag.ExitOnError),
	}
}

func (fs *FlagSet) Parse() {
	fs.FlagSet.Parse(os.Args[1:])
	fs.setVisited()
}

func (fs *FlagSet) IsSet(name string) bool {
	_, isSet := fs.vSet[name]
	return isSet
}

func (fs *FlagSet) setVisited() {
	fs.vSet = make(map[string]struct{}, fs.FlagSet.NFlag())
	fs.FlagSet.Visit(func(f *flag.Flag) {
		fs.vSet[f.Name] = struct{}{}
	})
}
