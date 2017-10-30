package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"log"
	flags "github.com/jessevdk/go-flags"
	"github.com/robfig/cron"
	"syscall"
	"os/signal"
)

const (
	// application informations
	Name    = "go-azure-dns-sync"
	Author  = "Markus Blaschke"
	Version = "0.1.0"
)

var (
	argparser *flags.Parser
	args []string
)

var opts struct {
	Verbose     []bool  `short:"v"  long:"verbose"      description:"verbose mode"`
	AzureConfig string  `           long:"azure-config" description:"azure configuration file" default:"/etc/kubernetes/azure.json"`
	Config      string  `           long:"config"       description:"DNS configuration file"   default:"/etc/azure-dns-sync/config.yml"`
	UpdateTime  string  `           long:"update-time"  description:"update time"              default:"@every 10m"`
}

func handleArgParser() {
	var err error
	argparser = flags.NewParser(&opts, flags.Default)
	args, err = argparser.Parse()

	// check if there is an parse error
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println()

			message := fmt.Sprintf("%v", r)

			if len(opts.Verbose) >= 2 {
				fmt.Println(message)
				debug.PrintStack()
			} else {
				fmt.Println(message)
			}
			os.Exit(255)
		}
	}()

	// signal channel (listen on SIGINT, SIGTERM and SIGHUP)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGHUP)

	handleArgParser()

	// Init azure configuration
	azureConf := AzureConfiguration{}
	if err := azureConf.ReadFromConfig(opts.AzureConfig); err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	rc, err := azureConf.GetClient()
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	// read dns configuration
	conf := DnsConfiguration{}
	if err := conf.ReadFromConfig(opts.Config); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	conf.SetAzureClient(rc)

	// create cron
	cron := cron.New()
	cron.AddFunc(opts.UpdateTime, func() {
		if err := conf.Run(); err != nil {
			log.Fatalf("Error: %v\n", err)
		}
	})

	// run cron
	log.Printf("Starting %s daemon version %v\n", Name, Version)
	cron.Start()
	s := <-c
	log.Printf("Got signal: %v\n", s)
	cron.Stop()

	os.Exit(0)
}
