package tlsh

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestSomething(t *testing.T) {
}

func TestReal(t *testing.T) {
	foo, err := exec.Command("tlsh", "-f", "LICENSE").Output()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s", foo)
	if bar, err := Hash("LICENSE"); bar != string(foo[:]) {
		fmt.Printf("%s", bar)
		if err != nil {
			t.Error(err)
		}
		t.Error("Calculated hash doesn't match real hash")
	}
}
