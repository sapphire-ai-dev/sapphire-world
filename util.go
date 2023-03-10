package world

import "fmt"

func printErr(err error) {
	if err != nil {
		fmt.Println(err)
    }
}
