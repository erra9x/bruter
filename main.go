package main

import (
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/vflame6/bruter/cmd"
	"github.com/vflame6/bruter/logger"
	"os"
)

var (
	app       = kingpin.New("bruter", "bruter is a network services bruteforce tool.")
	quietFlag = app.Flag("quiet", "Enable quiet mode, print results only").Short('q').Default("false").Bool()
	debugFlag = app.Flag("debug", "Enable debug mode, print all logs").Short('D').Default("false").Bool()

	// file output flags
	outputFlag = app.Flag("output", "Filename to write output in raw format").Short('o').Default("").String()

	// optimization flags
	parallelFlag = app.Flag("parallel", "Number of targets in parallel").Short('T').Default("16").Int()
	threadsFlag  = app.Flag("threads", "Number of threads per target").Short('t').Default("5").Int()
	delayFlag    = app.Flag("delay", "Delay in millisecond between each attempt. Will always use single thread if set").Short('d').Default("0").Int()
	timeoutFlag  = app.Flag("timeout", "Connection timeout in seconds").Default("5").Int()

	// wordlist flags
	usernameFlag = app.Flag("username", "Username or file with usernames").Short('u').Required().String()
	passwordFlag = app.Flag("password", "Password or file with passwords").Short('p').Required().String()

	// clickhouse
	// default port 9000
	clickhouseCommand   = app.Command("clickhouse", "clickhouse module")
	clickhouseTargetArg = clickhouseCommand.Arg("target", "Target host or file with targets. Format host or host:port, one per line").Required().String()
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

	if err := logger.Init(*quietFlag, *debugFlag); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// print program banner
	if !*quietFlag {
		cmd.PrintBanner()
	}

	// show which module is executed
	if !*quietFlag {
		logger.Infof("executing %s module", command)
	}

	s, err := cmd.CreateScanner(
		*timeoutFlag,
		*outputFlag,
		*parallelFlag,
		*threadsFlag,
		*delayFlag,
		*usernameFlag,
		*passwordFlag,
	)
	if err != nil {
		logger.Fatal(err)
	}

	if command == clickhouseCommand.FullCommand() {
		err = s.Run(command, *clickhouseTargetArg)
	}
	if err != nil {
		logger.Fatal(err)
	}
	s.Stop()

	// show which module is done its execution
	if !*quietFlag {
		logger.Infof("finished execution of %s module", command)
	}
}
