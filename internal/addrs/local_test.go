package addrs

import (
	"strings"
	"testing"
)

func Test_Local(t *testing.T) {
	Test := []Local{
		{
			Name: "kubehcl",
		},
		{
			Name: "kubehcl",
		},
		{
			Name: "version",
		},
		{
			Name: "application",
		},
	}

	for i := 1; i < len(Test)-1; i++ {
		if Test[i].UniqueKey() == Test[i+1] {
			t.Errorf("2 different default keys are equal: %s, %s", Test[i].String(), Test[i+1].String())
		}
	}

	for i := 1; i < len(Test); i++ {

		if !strings.HasPrefix(Test[i].String(), "local.") {
			t.Errorf("Annotation addr must start with tag this starts with: %s", Test[i].String())
		}
	}
}
