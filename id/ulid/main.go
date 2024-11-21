package main

import (
	"fmt"

	"github.com/oklog/ulid/v2"
)

func main() {
	res := ulid.Make()
	fmt.Println(res, ulid.Make().String())
	// 01JCTFTC8W S34X5KXJFFTB26CR
	// 01JCTFTC8X SM9W3WANHKNNYHKB
}
