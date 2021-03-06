package ascgen

import (
	"fmt"
	"os"
	"testing"

	ct "github.com/daviddengcn/go-colortext"
)

func TestShow(t *testing.T) {
	f, err := os.Open("pic.jpg")
	if err != nil {
		t.Skip()
		return
	}
	err = ShowFile(os.Stdout, f, Console{6, 14, 120}, false, false)
	if err != nil {
		fmt.Println(err)
	}
}

func TestShowColor(t *testing.T) {
	f, err := os.Open("cc.png")
	if err != nil {
		t.Skip(err)
	}
	err = ShowFile(os.Stdout, f, Console{6, 14, 120}, true, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetConsoleColor(t *testing.T) {
	for b := uint32(0); b < 6; b++ {
		for g := uint32(0); g < 6; g++ {
			for r := uint32(0); r < 6; r++ {
				m, l, h, j, k := GetConsoleColor(r*0x33/4, g*0x33/4, b*0x33/4)
				ct.ChangeColor(l, h, j, k)
				c := clist64[m : m+1]
				c = c + c + c + c
				if k {
					fmt.Printf("%s%x", c, (j-1)<<1)
				} else {
					fmt.Printf("%s%x", c, (j - 1))
				}
				ct.ResetColor()
				fmt.Printf("#%02x%02x%02x ", r*0x33, g*0x33, b*0x33)
			}
			fmt.Print("\n")
		}
	}
}
