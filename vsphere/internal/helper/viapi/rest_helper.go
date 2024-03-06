package viapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/vmware/govmomi/vapi/rest"
)

// RestRequest makes a rest request to endpoint and returns the given generic format from response
func RestRequest[T map[string]interface{} | []interface{}](client *rest.Client, method, endpoint string, body interface{}) (T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	var res T
	var err error
	var buf io.Reader

	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return res, fmt.Errorf("error trying to convert body to json: %s", err)
		}

		buf = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequest(method, client.URL().String()+endpoint, buf)
	if err != nil {
		return res, fmt.Errorf("error generating http request with payload: %s", err)
	}

	err = client.Do(ctx, req, &res)
	if err != nil && err != io.EOF {
		return res, fmt.Errorf("error making http request with payload: %s", err)
	}

	return res, nil
}
