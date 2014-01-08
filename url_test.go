package entropy

import (
	"reflect"
	"testing"
)

type TestHandler struct {
	Handler
}

func TestUrl2Regexp(t *testing.T) {
	spec := NewURLSpec("/home/:p1/:p2", reflect.ValueOf(&TestHandler{}), "ename", "cname")
	exp, err := spec.Url2Regexp()
	if err != nil {
		t.Fatalf("%v", exp)
	} else {
		t.Logf("%v", exp)
	}
}

func TestUrlSetParams(t *testing.T) {
	spec := NewURLSpec("/home/:p1/:p2", reflect.ValueOf(&TestHandler{}), "ename", "cname")
	url, err := spec.UrlSetParams("abc", "edf")
	if err != nil {
		t.Fatalf("%v", err)
	} else {
		t.Logf("%v", url)
	}
}

func TestParseUrlParams(t *testing.T) {
	spec := NewURLSpec("/home/:p1/:p2", reflect.ValueOf(&TestHandler{}), "ename", "cname")
	args := spec.ParseUrlParams("/home/frank/yang")
	t.Logf("%v", args)
}
