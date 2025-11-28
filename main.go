package main

import (
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/vflame6/bruter/cmd"
	"log"
	"os"
)

var (
	app       = kingpin.New("bruter", "bruter is a network services bruteforce tool.")
	quietFlag = app.Flag("quiet", "Enable quiet mode, print results only").Short('q').Bool()

	// file output flags
	outputFlag = app.Flag("output", "Filename to write output in raw format").Short('o').Default("").String()

	// optimization flags
	delayFlag = app.Flag("delay", "Delay between requests in milliseconds").Short('d').Default("0").Int()

	// wordlist flags
	usernameFlag = app.Flag("username", "Username or file with usernames").Short('u').Required().String()
	passwordFlag = app.Flag("password", "Password or file with passwords").Short('p').Required().String()

	// clickhouse
	clickhouseCommand   = app.Command("clickhouse", "clickhouse module")
	clickhouseTargetArg = clickhouseCommand.Arg("target", "Target host").Required().String()
	clickhousePortFlag  = clickhouseCommand.Flag("port", "port for ClickHouse service").Default("9000").Int()
)

func main() {
	// VERSION is linked to actual tag
	VERSION := "0.0.1"

	// kingpin settings
	app.Version(VERSION)
	app.Author("vflame6")
	app.HelpFlag.Short('h')
	app.UsageTemplate(kingpin.CompactUsageTemplate)

	// parse options
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	// print program banner
	if !*quietFlag {
		cmd.PrintBanner()
	}

	// show which module is executed
	if !*quietFlag {
		log.Println(fmt.Sprintf("[*] Executing %s module", command))
	}

	s, err := cmd.CreateScanner(*outputFlag, *delayFlag, *usernameFlag, *passwordFlag)
	if err != nil {
		log.Fatal(err)
	}

	if command == clickhouseCommand.FullCommand() {
		err = s.RunClickHouse(*clickhouseTargetArg, *clickhousePortFlag)
	}
	if err != nil {
		log.Fatal(err)
	}
	s.Stop()

	// show which module is done its execution
	if !*quietFlag {
		log.Println(fmt.Sprintf("[*] Finished execution of %s module", command))
	}
}
