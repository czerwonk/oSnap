package api

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"regexp"
	"strings"

	ovirt_api "github.com/czerwonk/ovirt_api/api"
)

const snapshotSuffix = " - created by oSnap"

type Option func(*Client)

// Client encapsulates communication with the oVirt API
type Client struct {
	debug         bool
	client        *ovirt_api.Client
	clusterFilter string
	includes      []*regexp.Regexp
	excludes      []*regexp.Regexp
}

// WithClusterFilter sets the cluster filter
func WithClusterFilter(filter string) Option {
	return func(c *Client) {
		c.clusterFilter = filter
	}
}

// WithIncludes sets the include filters
func WithIncludes(filters []string) Option {
	return func(c *Client) {
		for _, f := range filters {
			r := regexp.MustCompile(f)
			c.includes = append(c.includes, r)
		}
	}
}

// WithExcludes sets the exclude filters
func WithExcludes(filters []string) Option {
	return func(c *Client) {
		for _, f := range filters {
			r := regexp.MustCompile(f)
			c.excludes = append(c.excludes, r)
		}
	}
}

// NewClient creates a new API client
func NewClient(url, user, pass string, insecureCert, debug bool, o ...Option) (*Client, error) {
	opts := []ovirt_api.ClientOption{}
	if insecureCert {
		opts = append(opts, ovirt_api.WithInsecure())
	}

	if debug {
		opts = append(opts, ovirt_api.WithDebug())
	}

	a, err := ovirt_api.NewClient(url, user, pass, opts...)
	if err != nil {
		return nil, err
	}

	c := &Client{
		client:   a,
		debug:    debug,
		includes: make([]*regexp.Regexp, 0),
		excludes: make([]*regexp.Regexp, 0),
	}

	for _, option := range o {
		option(c)
	}

	return c, nil
}

// GetVMs retrieves the list of VMs
func (c *Client) GetVMs() ([]Vm, error) {
	clusterID, err := c.getClusterID(c.clusterFilter)
	if err != nil {
		return nil, err
	}

	vms := Vms{}
	err = c.client.SendAndParse("vms", "GET", &vms, nil)
	if err != nil {
		return nil, err
	}

	res := make([]Vm, 0)
	for _, v := range vms.VM {
		if (v.Cluster.ID == clusterID || len(c.clusterFilter) == 0) && c.shouldProcessVM(v.Name) {
			res = append(res, v)
		}
	}

	return res, nil
}

func (c *Client) shouldProcessVM(name string) bool {
	if len(c.includes) == 0 && len(c.excludes) == 0 {
		return true
	}

	for _, exclude := range c.excludes {
		if exclude.MatchString(name) {
			return false
		}
	}

	for _, include := range c.includes {
		if include.MatchString(name) {
			return true
		}
	}

	return false
}

func (c *Client) getClusterID(name string) (string, error) {
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
			return cluster.ID, nil
		}
	}

	return "", fmt.Errorf("Unknown cluster %s", name)
}

// CreateSnapshot creates a snapshot
func (c *Client) CreateSnapshot(vmID, desc string) (*Snapshot, error) {
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
	err = c.client.SendAndParse(fmt.Sprintf("vms/%s/snapshots", vmID), "POST", &res, r)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// GetSnapshot returns details for a snapshot
func (c *Client) GetSnapshot(vmID, snapshotid string) (*Snapshot, error) {
	res := Snapshot{}
	err := c.client.SendAndParse(fmt.Sprintf("vms/%s/snapshots/%s", vmID, snapshotid), "GET", &res, nil)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// GetCreatedSnapshots returns all snapshots created by oSnap for the specified VM
func (c *Client) GetCreatedSnapshots(vmID string) ([]Snapshot, error) {
	res := Snapshots{}
	err := c.client.SendAndParse(fmt.Sprintf("vms/%s/snapshots", vmID), "GET", &res, nil)
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

// DeleteSnapshot deletes a snapshot
func (c *Client) DeleteSnapshot(vmId, snapShotId string) error {
	_, err := c.client.SendRequest(fmt.Sprintf("vms/%s/snapshots/%s", vmId, snapShotId), "DELETE", nil)
	if err != nil {
		return err
	}

	return nil
}
