package entropy

import (
	"fmt"
	"strings"
)

type Blueprint struct {
	Prefix        string
	BeforeFilters []Filter
	NamedHandlers map[string]*URLSpec
	AfterFilters  []Filter
}

func NewBlueprint(prefix string) *Blueprint {
	return &Blueprint{
		Prefix:        prefix,
		BeforeFilters: make([]Filter, 0),
		NamedHandlers: make(map[string]*URLSpec, 0),
		AfterFilters:  make([]Filter, 0),
	}
}

func (self *Blueprint) Before(filter Filter) {
	self.BeforeFilters = append(self.BeforeFilters, filter)
}

func (self *Blueprint) After(filter Filter) {
	self.AfterFilters = append(self.AfterFilters, filter)
}

func (self *Blueprint) Handle(pattern string, eName string, cName string, handler Handler) {
	//pattern:/home/str:action/int:id
	if !strings.HasSuffix(pattern, "$") {
		pattern = pattern + "$"
	}
	if !strings.HasPrefix(pattern, "^") {
		pattern = "^" + pattern
	}
	if _, exist := self.NamedHandlers[eName]; exist {
		panic(fmt.Sprintf("Here is a handler named %s in blueprint %s", eName, self.Prefix))
	}
	self.NamedHandlers[eName] = NewURLSpec(pattern, handler, eName, cName)
}
