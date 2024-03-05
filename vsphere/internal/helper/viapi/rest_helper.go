package viapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/vmware/govmomi/vapi/rest"
)

func GetRestBodyResponse[T any](client *rest.Client, endpoint string) (T, error) {
	var resMap T

	fullURL := client.URL().String() + endpoint
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return resMap, fmt.Errorf("error generating http request: %s", err)
	}

	if err = client.Do(context.Background(), req, &resMap); err != nil {
		return resMap, fmt.Errorf("error making http request: %s", err)
	}

	log.Printf("[DEBUG] res here: %+v\n", resMap)

	return resMap, nil
}

func RestUpdateRequest(client *rest.Client, method, endpoint string, body interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	var buf *bytes.Buffer

	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("error trying to convert body to json: %s", err)
		}

		buf = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequest(method, client.URL().String()+endpoint, buf)
	if err != nil {
		return fmt.Errorf("error generating http request with payload: %s", err)
	}

	if err = client.Do(ctx, req, nil); err != nil {
		return fmt.Errorf("error making http request with payload: %s", err)
	}

	return nil
}
