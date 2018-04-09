package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/czerwonk/oSnap/api"
	"github.com/czerwonk/oSnap/config"
)

const version = "0.4.0"

var (
	showVersion = flag.Bool("version", false, "Print version information")
	configFile  = flag.String("config.file", "config.yml", "Path to config file")
	debug       = flag.Bool("debug", false, "Prints API requests and responses to STDOUT")
	purgeOnly   = flag.Bool("purge-only", false, "Only deleting old snapshots without creating a new one")
	dry         = flag.Bool("dry", false, "Print names of VMs instead of creating actual snapshots")
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: oSnap [ ... ]\n\nParameters:")
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

	cfg, err := loadConfig()
	if err != nil {
		log.Println("could not load config file.", err)
		os.Exit(2)
	}

	err = run(cfg)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func loadConfig() (*config.Config, error) {
	b, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return nil, err
	}

	return config.Load(bytes.NewReader(b))
}

func printVersion() {
	fmt.Println("oSnap - oVirt Snapshot Creator")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): Daniel Czerwonk")
}

func run(cfg *config.Config) error {
	client, err := connectAPI(cfg)
	if err != nil {
		return err
	}

	vms, err := client.GetVMs()
	if err != nil {
		return err
	}

	var snapped []api.Vm
	if !*purgeOnly {
		snapped = createSnapshots(vms, cfg, client)
	}

	if *dry {
		return nil
	}

	var success int
	if *purgeOnly {
		success = purgeOldSnapshots(vms, cfg, client)
	} else {
		success = purgeOldSnapshots(snapped, cfg, client)
	}

	if success != len(vms) {
		return fmt.Errorf("One or more errors occurred. See output above for more detail.")
	}

	return nil
}

func connectAPI(cfg *config.Config) (*api.Client, error) {
	opts := []api.Option{
		api.WithClusterFilter(cfg.Cluster),
		api.WithIncludes(cfg.Includes),
		api.WithExcludes(cfg.Excludes),
	}

	client, err := api.NewClient(cfg.API.URL, cfg.API.User, cfg.API.Password, cfg.API.Insecure, *debug, opts...)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func createSnapshots(vms []api.Vm, cfg *config.Config, a *api.Client) []api.Vm {
	snapshots := make([]*api.Snapshot, 0)
	for _, vm := range vms {
		log.Printf("%s: Creating snapshot for VM", vm.Name)
		if *dry {
			continue
		}

		s, err := a.CreateSnapshot(vm.ID, cfg.Description)
		if err != nil {
			log.Printf("%s: Snapshot failed - %v)\n", vm.Name, err)
		}

		snapshots = append(snapshots, s)
		log.Printf("%s: Snapshot job created. (ID: %s)\n", vm.Name, s.ID)
	}

	return monitorSnapshotCreation(snapshots, a)
}

func monitorSnapshotCreation(snapshots []*api.Snapshot, client *api.Client) []api.Vm {
	complete := make([]api.Vm, 0)

	for _, s := range snapshots {
		x, err := waitForCompletion(s, client)
		if err != nil {
			log.Printf("%s: Snapshot failed - %v)\n", s.VM.Name, err)
		} else {
			log.Printf("%s: Snapshot completed\n", x.VM.Name)
			complete = append(complete, x.VM)
		}
	}

	return complete
}

func waitForCompletion(snapshot *api.Snapshot, client *api.Client) (*api.Snapshot, error) {
	log.Printf("Waiting for snapshot %s to finish...\n", snapshot.ID)

	for {
		s, err := client.GetSnapshot(snapshot.VM.ID, snapshot.ID)
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(s.Status, "fail") || strings.HasPrefix(s.Status, "error") {
			return nil, fmt.Errorf(s.Status)
		}

		if s.Status == "ok" {
			return s, nil
		}

		time.Sleep(30 * time.Second)
	}
}
