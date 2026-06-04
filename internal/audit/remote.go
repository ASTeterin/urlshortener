package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RemoteReceiver struct {
	url    string
	client *http.Client
}

func NewRemoteReceiver(url string) *RemoteReceiver {
	return &RemoteReceiver{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (r *RemoteReceiver) Notify(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, r.url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("remote audit failed: %d", resp.StatusCode)
	}
	return nil
}
