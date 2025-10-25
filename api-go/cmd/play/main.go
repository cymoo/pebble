package main

import (
	"fmt"

	"github.com/go-ego/gse"
)

func main() {
	text := "我来到北京清华大学"
	testGseCut(text)
}

func testGseCut(text string) {
	seg, _ := gse.New("zh")
	segments := seg.Cut(text)
	fmt.Println(segments)
}
