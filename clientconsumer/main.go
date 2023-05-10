package main

import (
	"fmt"
	"os"
)

func main() {
	if err := Run(); err != nil {
		fmt.Println("启动失败：", err)
		os.Exit(1)
	}
}
