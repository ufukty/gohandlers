package version

import "fmt"

var Version string

func Main() error {
	fmt.Println(Version)
	return nil
}
