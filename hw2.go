package main

import (
	"fmt"
	//"hw2/cachettl"
	"time"

	"github.com/Aizengott/udemi_hw2/cachettl"
)

func main() {
	cache1 := cachettl.New()
	cache1.Set("aa", 1, time.Millisecond*200)
	time.Sleep(time.Millisecond * 300)

	//cache1.Delete("aa")
	//cache1.DelExpired()

	fmt.Println(cache1.Get("aa"))
	// cache1.Delete("aa")
	// fmt.Println(cache1.Get("aa"))

}
