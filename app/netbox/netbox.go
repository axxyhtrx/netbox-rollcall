package netbox

import (
	"context"
	"crypto/tls"
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/ipam"
	"github.com/netbox-community/go-netbox/netbox/models"
	"net/http"
)

type IPaddressStatus struct {
	Prefix  string
	VRF     string
	Status  bool
}

func PushIPAddress(c []IPaddressStatus) {
	for _, ip := range c {
		if ip.Status == true {
			fmt.Println(ip.Prefix)
		}
	}
}


func He(token string, netboxHost string) ([]models.IPAddress, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	transport := httptransport.NewWithClient(netboxHost, client.DefaultBasePath, []string{"https"}, httpClient)
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Token "+token)

	c := client.New(transport, nil)
	result := make([]models.IPAddress, 0)
	params := ipam.NewIpamIPAddressesListParams()

	limit := int64(2)
	params.Limit = &limit
	params.SetContext(context.Background())
	for {
		offset := int64(0)
		if params.Offset != nil {
			offset = *params.Offset + limit
		}
		params.Offset = &offset
		list, err := c.Ipam.IpamIPAddressesList(params, nil)
		if err != nil {
			return nil, err
		}
		for _, prefix := range list.Payload.Results {
			result = append(result, *prefix)


			address := "99.99.66.99/30"

			resource := ipam.NewIpamIPAddressesCreateParams().WithDefaults()
			resource.Data.Address = &address
			_, err = c.Ipam.IpamIPAddressesCreate(resource, nil)
			if err != nil {
				fmt.Println(err)
			}
		}
		if list.Payload.Next == nil {
			break
		}
	}

	return result, nil

}



