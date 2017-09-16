package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/czerwonk/osnap/api"
)

const version = "0.2.0"

var (
	showVersion     = flag.Bool("version", false, "Print version information")
	apiUrl          = flag.String("api.url", "https://localhost/ovirt-engine/api/", "API REST Endpoint")
	apiUser         = flag.String("api.user", "user@internal", "API username")
	apiPass         = flag.String("api.pass", "", "API password")
	apiInsecureCert = flag.Bool("api.insecure-cert", false, "Skip verification for untrusted SSL/TLS certificates")
	cluster         = flag.String("cluster", "", "Cluster name to filter")
	vm              = flag.String("vm", "", "VM name to filter")
	desc            = flag.String("desc", "oSnap generated snapshot", "Description to use for the snapshot")
	keep            = flag.Int("keep", 7, "Number of snapshots to keep")
	debug           = flag.Bool("debug", false, "Prints API requests and responses to STDOUT")
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
	fmt.Println("oSnap - oVirt Snapshot Creator")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): Daniel Czerwonk")
}

func createSnapshots() error {
	a := api.NewClient(*apiUrl, *apiUser, *apiPass, *apiInsecureCert, *debug)

	vms, err := a.GetVms(*cluster, *vm)
	if err != nil {
		return err
	}

	snapshots := make([]*api.Snapshot, 0)
	for _, vm := range vms {
		log.Printf("%s: Creating snapshot for VM", vm.Name)
		s, err := a.CreateSnapshot(vm.Id, *desc)
		if err != nil {
			return err
		}

		snapshots = append(snapshots, s)
		log.Printf("%s: Snapshot job created. (ID: %s)\n", vm.Name, s.Id)
	}

	snapshots = monitorSnapshotCreation(snapshots, a)
	purgeSucceeded := purgeOldSnapshots(snapshots, a)

	if len(snapshots) != len(vms) || purgeSucceeded != len(vms) {
		return errors.New("One or more errors occured. See output above for more detail.")
	}

	return nil
}

func monitorSnapshotCreation(snapshots []*api.Snapshot, a *api.ApiClient) []*api.Snapshot {
	complete := make([]*api.Snapshot, 0)

	for _, s := range snapshots {
		x, err := waitForCompletion(s, a)
		if err != nil {
			log.Printf("%s: Snapshot failed - %v)\n", s.Vm.Name, err)
		} else {
			log.Printf("%s: Snapshot completed\n", x.Vm.Name)
			complete = append(complete, x)
		}
	}

	return complete
}

func waitForCompletion(snapshot *api.Snapshot, a *api.ApiClient) (*api.Snapshot, error) {
	log.Printf("Waiting for snapshot %s to finish...\n", snapshot.Id)

	for {
		s, err := a.GetSnapshot(snapshot.Vm.Id, snapshot.Id)
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(s.Status, "fail") || strings.HasPrefix(s.Status, "error") {
			return nil, errors.New(s.Status)
		}

		if s.Status == "ok" {
			return s, nil
		}

		time.Sleep(30 * time.Second)
	}
}
