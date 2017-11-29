package api

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"strings"

	"regexp"

	"github.com/czerwonk/ovirt_api"
)

type Api struct {
	debug         bool
	client        *ovirt_api.ApiClient
	ClusterFilter string
	VmFilter      *regexp.Regexp
	SkipFilter    *regexp.Regexp
}

const snapshotSuffix = " - created by oSnap"

func New(url, user, pass string, insecureCert, debug bool) (*Api, error) {
	c, err := ovirt_api.NewClient(url, user, pass, insecureCert)
	if err != nil {
		return nil, err
	}

	c.Debug = debug

	return &Api{client: c, debug: debug}, nil
}

func (c *Api) GetVms() ([]Vm, error) {
	clusterId, err := c.getClusterId(c.ClusterFilter)
	if err != nil {
		return nil, err
	}

	vms := Vms{}
	err = c.client.SendAndParse("vms", "GET", &vms, nil)
	if err != nil {
		return nil, err
	}

	res := make([]Vm, 0)
	for _, v := range vms.Vm {
		if (v.Cluster.Id == clusterId || len(c.ClusterFilter) == 0) && c.shouldProcessVm(v.Name) {
			res = append(res, v)
		}
	}

	return res, nil
}

func (c *Api) shouldProcessVm(name string) bool {
	if c.SkipFilter != nil && c.SkipFilter.MatchString(name) {
		return false
	}

	if c.VmFilter != nil {
		return c.VmFilter.MatchString(name)
	}

	return true
}

func (c *Api) getClusterId(name string) (string, error) {
	if len(name) == 0 {
		return "", nil
	}

	clusters := Clusters{}
	err := c.client.SendAndParse(fmt.Sprintf("clusters?search=%s", name), "GET", &clusters, nil)
	if err != nil {
		return "", err
	}

	for _, cluster := range clusters.Cluster {
		if cluster.Name == name {
			return cluster.Id, nil
		}
	}

	return "", fmt.Errorf("Unknown cluster %s", name)
}

func (c *Api) CreateSnapshot(vmId, desc string) (*Snapshot, error) {
	s := &Snapshot{Description: desc + snapshotSuffix, PersistMemoryState: false}
	b, err := xml.Marshal(s)
	if err != nil {
		return nil, err
	}

	if c.debug {
		log.Println(string(b))
	}

	r := bytes.NewReader(b)
	res := Snapshot{}
	err = c.client.SendAndParse(fmt.Sprintf("vms/%s/snapshots", vmId), "POST", &res, r)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Api) GetSnapshot(vmId, snapshotid string) (*Snapshot, error) {
	res := Snapshot{}
	err := c.client.SendAndParse(fmt.Sprintf("vms/%s/snapshots/%s", vmId, snapshotid), "GET", &res, nil)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Api) GetCreatedSnapshots(vmId string) ([]Snapshot, error) {
	res := Snapshots{}
	err := c.client.SendAndParse(fmt.Sprintf("vms/%s/snapshots", vmId), "GET", &res, nil)
	if err != nil {
		return nil, err
	}

	snaps := make([]Snapshot, 0)
	for _, s := range res.Snapshot {
		if strings.HasSuffix(s.Description, snapshotSuffix) {
			snaps = append(snaps, s)
		}
	}

	return snaps, err
}

func (c *Api) DeleteSnapshot(vmId, snapShotId string) error {
	_, err := c.client.SendRequest(fmt.Sprintf("vms/%s/snapshots/%s", vmId, snapShotId), "DELETE", nil)
	if err != nil {
		return err
	}

	return nil
}
