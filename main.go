package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/czerwonk/osnap/api"
)

const version = "0.1"

var (
	showVersion     = flag.Bool("version", false, "Print version information")
	apiUrl          = flag.String("api.url", "https://localhost/ovirt-engine/api/", "API REST Endpoint")
	apiUser         = flag.String("api.user", "user@internal", "API username")
	apiPass         = flag.String("api.pass", "", "API password")
	cluster         = flag.String("cluster", "", "Cluster to filter")
	apiInsecureCert = flag.Bool("api.insecure-cert", false, "Skip verification for untrusted SSL/TLS certificates")
	desc            = flag.String("desc", "oSnap generated snapshot", "Description to use for the snapshot")
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: osnap [ ... ]\n\nParameters:")
		fmt.Println()
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	err := createSnapshots()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Println("osnap - oVirt Snapshot Creator")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): Daniel Czerwonk")
}

func createSnapshots() error {
	a := api.NewClient(*apiUrl, *apiUser, *apiPass, *apiInsecureCert)

	vms, err := a.GetVms(*cluster)
	if err != nil {
		return err
	}

	for _, vm := range vms {
		log.Printf("Creating snapshot for VM: %s", vm.Name)
		err = a.CreateSnapshot(vm.Id, *desc)
		if err != nil {
			return err
		}
	}

	return nil
}
