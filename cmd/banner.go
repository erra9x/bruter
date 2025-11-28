package cmd

import "fmt"

// Banner string
var Banner = "    __               __           \n   / /_  _______  __/ /____  _____\n  / __ \\/ ___/ / / / __/ _ \\/ ___/\n / /_/ / /  / /_/ / /_/  __/ /    \n/_.___/_/   \\__,_/\\__/\\___/_/     \n                                  "

// PrintBanner is a function to print program banner
func PrintBanner() {
	fmt.Println(Banner)
}
