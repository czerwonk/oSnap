package main

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/czerwonk/osnap/api"
)

func purgeOldSnapshots(snapshots []*api.Snapshot, api *api.ApiClient) int {
	success := 0

	for _, s := range snapshots {
		err := purgeVmSnapshots(s, api)
		if err != nil {
			log.Printf("%s: Purging failed - %v\n", s.Vm.Name, err)
		} else {
			success++
		}
	}

	return success
}
func purgeVmSnapshots(snapshot *api.Snapshot, a *api.ApiClient) error {
	log.Printf("%s: Purging old snapshots\n", snapshot.Vm.Name)

	snaps, err := a.GetCreatedSnapshots(snapshot.Vm.Id)
	if err != nil {
		return err
	}

	l := len(snaps)
	d := l - 1 - *keep

	if d > 0 {
		return purgeSnapshots(snaps[0:d], &snapshot.Vm, a)
	} else {
		log.Printf("%s: Nothing to purge.\n", snapshot.Vm.Name)
	}

	return nil
}

func purgeSnapshots(snapshots []api.Snapshot, vm *api.Vm, a *api.ApiClient) error {
	log.Printf("%s: Purging %v old snapshots...\n", vm.Name, len(snapshots))

	for _, s := range snapshots {
		err := deleteSnapshot(&s, vm, a)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteSnapshot(s *api.Snapshot, vm *api.Vm, a *api.ApiClient) error {
	log.Printf("%s: Delete snapshot %s\n", vm.Name, s.Id)

	try := 0
	for try < 10 {
		err := a.DeleteSnapshot(vm.Id, s.Id)
		if err == nil {
			return nil
		}

		if !strings.HasPrefix(err.Error(), "409") {
			return err
		}

		log.Println("Conflict occured. Retry in 60 seconds.")
		try++
		time.Sleep(1 * time.Minute)
	}

	return errors.New("Max retries reached.")
}
