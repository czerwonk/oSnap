package api

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type ApiClient struct {
	url          string
	user         string
	pass         string
	insecureCert bool
	debug        bool
	client       *http.Client
}

const snapshotSuffix = " - created by oSnap"

func NewClient(url, user, pass string, insecureCert, debugMode bool) *ApiClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureCert},
	}
	c := &http.Client{Transport: tr}

	return &ApiClient{url: url, user: user, pass: pass, insecureCert: insecureCert, client: c, debug: debugMode}
}

func (c *ApiClient) GetVms(clusterFilter, vmFilter string) ([]Vm, error) {
	clusterId, err := c.getClusterId(clusterFilter)
	if err != nil {
		return nil, err
	}

	vms := Vms{}
	err = c.sendAndParse("vms", "GET", &vms, nil)
	if err != nil {
		return nil, err
	}

	res := make([]Vm, 0)
	for _, v := range vms.Vm {
		if (v.Cluster.Id == clusterId || len(clusterFilter) == 0) && (v.Name == vmFilter || len(vmFilter) == 0) {
			res = append(res, v)
		}
	}

	return res, nil
}

func (c *ApiClient) getClusterId(name string) (string, error) {
	if len(name) == 0 {
		return "", nil
	}

	clusters := Clusters{}
	err := c.sendAndParse(fmt.Sprintf("clusters?search=%s", name), "GET", &clusters, nil)
	if err != nil {
		return "", err
	}

	for _, cluster := range clusters.Cluster {
		if cluster.Name == name {
			return cluster.Id, nil
		}
	}

	return "", errors.New("Unknown cluster " + name)
}

func (c *ApiClient) CreateSnapshot(vmId, desc string) (*Snapshot, error) {
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
	err = c.sendAndParse(fmt.Sprintf("vms/%s/snapshots", vmId), "POST", &res, r)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *ApiClient) GetSnapshot(vmId, snapshotid string) (*Snapshot, error) {
	res := Snapshot{}
	err := c.sendAndParse(fmt.Sprintf("vms/%s/snapshots/%s", vmId, snapshotid), "GET", &res, nil)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *ApiClient) GetCreatedSnapshots(vmId string) ([]Snapshot, error) {
	res := Snapshots{}
	err := c.sendAndParse(fmt.Sprintf("vms/%s/snapshots", vmId), "GET", &res, nil)
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

func (c *ApiClient) DeleteSnapshot(vmId, snapShotId string) error {
	_, err := c.sendRequest(fmt.Sprintf("vms/%s/snapshots/%s", vmId, snapShotId), "DELETE", nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *ApiClient) sendAndParse(path, method string, res interface{}, body io.Reader) error {
	b, err := c.sendRequest(path, method, body)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(b, res)
	return err
}

func (c *ApiClient) sendRequest(path, method string, body io.Reader) ([]byte, error) {
	uri := strings.Trim(c.url, "/") + "/" + strings.Trim(path, "/")
	log.Println(method, uri)

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/xml")
	req.SetBasicAuth(c.user, c.pass)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, errors.New(fmt.Sprintf(resp.Status))
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Println(resp.Status)
	if c.debug {
		log.Println(string(b))
	}

	return b, err
}
