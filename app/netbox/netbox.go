package netbox

import (
	"context"
	"crypto/tls"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/ipam"
	"github.com/netbox-community/go-netbox/netbox/models"
	"log"
	"net/http"
)


func NetboxLogin(token string, netboxHost string) (c *client.NetBoxAPI){

	httpClient := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	transport := httptransport.NewWithClient(netboxHost, client.DefaultBasePath, []string{"https"}, httpClient)
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Token "+token)

	cli := client.New(transport, nil)

	return cli
}

func GetIPAddesses(c *client.NetBoxAPI) ([]models.IPAddress, error) {

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
		}
		if list.Payload.Next == nil {
			break
		}
	}

	return result, nil

}

func CreateIPAddress (c *client.NetBoxAPI, ip models.IPAddress) {
	tagname := "Scanned"
	tag := []*models.NestedTag{{Name: &tagname, Slug: &tagname}}
	address := ip.Address

	parameters := ipam.NewIpamIPAddressesCreateParams()
	parameters.WithData(&models.WritableIPAddress{
		Address: address,
		Tags:    tag,
	})

	_, err := c.Ipam.IpamIPAddressesCreate(parameters, nil)
	if err != nil {
		log.Fatalf("Error creating IP Address: ", err)
	}

}