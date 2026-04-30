package main

import (
	"fmt"
	"github.com/tuxnode/TinyCache/internal/cache"
)

func main() {
	c := &cache.Cache{}

	c.Add("name", []byte("test"))

	if v, ok := c.Get("name"); ok {
		fmt.Printf("%s \n", string(v))
	} else {
		fmt.Println("Unfind in Cache")
	}
}
