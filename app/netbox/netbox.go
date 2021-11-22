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

func NetboxLogin(token string, netboxHost string) (c *client.NetBoxAPI) {

	httpClient := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	transport := httptransport.NewWithClient(netboxHost, client.DefaultBasePath, []string{"https"}, httpClient)
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Token "+token)

	cli := client.New(transport, nil)

	return cli
}

func GetIPAddresses(c *client.NetBoxAPI) ([]models.IPAddress, error) {

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

func CreateIPAddress(c *client.NetBoxAPI, ip models.IPAddress) {
	tagname := "Scanned"
	tag := []*models.NestedTag{{Name: &tagname, Slug: &tagname}}

	address := ip.Address


	parameters := ipam.NewIpamIPAddressesCreateParams()
	vrfname := *ip.Vrf.Name

    vrfid, err := GetVRFByName(c, vrfname)
    if err != nil {
    	fmt.Println(err)
	}

	parameters.WithData(&models.WritableIPAddress{
		Address: address,
		Tags:    tag,
		Vrf:   &vrfid,
	})

	_ , er := c.Ipam.IpamIPAddressesCreate(parameters, nil)
	//If address already exist, try to update Description field
	if er != nil {
		fmt.Println(*address, " address already exist in vrf", vrfname)
	}

}

func GetVRFByName(c *client.NetBoxAPI, vrfname string) (vrfid int64, err error) {
	vrfparams := ipam.NewIpamVrfsListParams()
	vrfparams.SetName(&vrfname)
	vr, err := c.Ipam.IpamVrfsList(vrfparams, nil)
	if err != nil {
		fmt.Println("VRF Not Found in Netbox")
		return 0, err
	}
	VRFID := vr.Payload.Results[0].ID
	return VRFID, nil
}
