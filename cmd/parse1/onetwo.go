package main

import (
	"fmt"
	"log"
	"os"
)

type Block struct {
	Type string
	Data string
	Kids []Block
}

const RunBlock = "run"
const OpaqueBlock = "opaque"
const ParagraphBlock = "paragraph"

func parse(s string) {
	var stack []*Block = Block{Type: BlockRun}

	var sentinelSeen bool
	var inBlock bool

	for _, r := range s {
		switch r {
		case '¶':
			stack = append(state, &Block{Type: ParagraphBlock})
		case '‖':
			stack = append(state, &Block{Type: RunBlock})
		case '◇':
		case '†':
		case '⁝':
		case '‣':
			stack = append(state, &Block{Type: RunBlock})
		case '{':
			if !sentinelSeen {
				// go into opaque
			}
		case '}':
			if sentinelSeen {
				panic("close to early")
			}
			if inBlock {
				stack = stack[:len(stack)-1]
			}
		}
	}
	var run []Block
}

func main() {
	bs, err := os.ReadFile("slides2.gba")
	if err != nil {
		log.Fatal(err)
	}

	var run []Block

	fmt.Fprint(os.Stdout, string(bs))
}
