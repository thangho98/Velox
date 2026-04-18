//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"

	"github.com/thawng/velox/pkg/nameparser"
)

func main() {
	names := []string{
		"[YamiS] One Piece - Tập 1010.mkv",
		"[SubsPlease] Frieren - Beyond Journey's End - 28 (1080p) [21345F].mkv",
		"[Erai-raws] Isekai - 05 (1080p).mkv",
	}

	for _, n := range names {
		p := nameparser.Parse(n)
		res, _ := json.Marshal(p)
		fmt.Printf("%s => %s\n", n, string(res))
	}
}
