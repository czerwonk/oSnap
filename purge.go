package main

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/czerwonk/oSnap/api"
	"github.com/czerwonk/oSnap/config"
)

func purgeOldSnapshots(vms []api.Vm, cfg *config.Config, client *api.Client) int {
	success := 0

	for _, vm := range vms {
		err := purgeVMSnapshots(&vm, cfg, client)
		if err != nil {
			log.Printf("%s: Purging failed - %v\n", vm.Name, err)
		} else {
			success++
		}
	}

	return success
}
func purgeVMSnapshots(vm *api.Vm, cfg *config.Config, client *api.Client) error {
	log.Printf("%s: Purging old snapshots\n", vm.Name)

	snaps, err := client.GetCreatedSnapshots(vm.ID)
	if err != nil {
		return err
	}

	l := len(snaps)
	d := l - 1 - cfg.Keep

	if d < 1 {
		log.Printf("%s: Nothing to purge.\n", vm.Name)
		return nil
	}

	return purgeSnapshots(snaps[0:d], vm, client)
}

func purgeSnapshots(snapshots []api.Snapshot, vm *api.Vm, client *api.Client) error {
	log.Printf("%s: Purging %v old snapshots...\n", vm.Name, len(snapshots))

	for _, s := range snapshots {
		err := deleteSnapshot(&s, vm, client)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteSnapshot(s *api.Snapshot, vm *api.Vm, client *api.Client) error {
	log.Printf("%s: Delete snapshot %s\n", vm.Name, s.ID)

	try := 0
	for try < 10 {
		err := client.DeleteSnapshot(vm.ID, s.ID)
		if err == nil {
			return nil
		}

		if !strings.HasPrefix(err.Error(), "409") {
			return err
		}

		log.Println("Conflict occurred. Retry in 60 seconds.")
		try++
		time.Sleep(1 * time.Minute)
	}

	return errors.New("Max retries reached.")
}
