package main

import (
	"fmt"

	"kubehcl.sh/kubehcl/cli"
)


func main() {
	cmd := cli.CreateRootCMD()
	if err := cmd.Execute(); err!= nil {
		fmt.Println(err)
	}

}
