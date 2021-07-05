package nonoconfig

import (
	"fmt"
	"os"
)

func ExampleSingleValues() {
	nnc := NewNoNoConfig("./testdata/config.yaml")

	var s string
	err := nnc.Config(&s, "single_string")
	if err != nil {
		fmt.Println("Unable to get single_string, %w", err)
		os.Exit(1)
	}
	fmt.Println(s)
	// Output: SingleString
}
