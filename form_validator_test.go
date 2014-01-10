package entropy

import (
	"fmt"
	"testing"
)

func TestEmailValidator(t *testing.T) {
	ret, err := XValidators["email"].Verify("frankyang418@gmail.com")
	if ret {
		fmt.Println(ret)
	} else {
		t.Fatal(err)
	}
}
