package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kanocz/goparse"
)

func main() {
	s, err := goparse.GetFileStructs(os.Args[1], "req", "req")
	if nil != err {
		fmt.Println("Error parsing file:", err)
		return
	}

	result, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println(string(result))
}
