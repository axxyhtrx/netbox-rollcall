package cmd

import (
	"fmt"
	"github.com/dean2021/go-nmap"
	"github.com/netbox-community/go-netbox/netbox/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net"
	"rollcall/app/netbox"
	"strings"
	"sync"
)

//Configuration file structure
type Config struct {
	Netbox struct {
		Netboxhost     string `yaml:"netboxhost"`
		Netboxapitoken string `yaml:"netboxapitoken"`
	} `yaml:"netbox"`
	Targets []struct {
		Vrf     string   `yaml:"vrf"`
		Subnets []string `yaml:"subnets"`
	} `yaml:"targets"`
	ScanThreads int `yaml:"scanthreads"`
}
type BoundedWaitGroup struct {
	wg sync.WaitGroup
	ch chan struct{}
}

// Bounded (with limit) wait group, to split target slice to fixed size slices.
func NewBoundedWaitGroup(cap int) BoundedWaitGroup {
	return BoundedWaitGroup{ch: make(chan struct{}, cap)}
}

func (bwg *BoundedWaitGroup) Add(delta int) {
	for i := 0; i > delta; i-- {
		<-bwg.ch
	}
	for i := 0; i < delta; i++ {
		bwg.ch <- struct{}{}
	}
	bwg.wg.Add(delta)
}

func (bwg *BoundedWaitGroup) Done() {
	bwg.Add(-1)
}

func (bwg *BoundedWaitGroup) Wait() {
	bwg.wg.Wait()
}

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "A brief description of your command",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		var C Config
		err := viper.Unmarshal(&C)
		// array with scan results
		var results []models.IPAddress
		ch := make(chan models.IPAddress)

		if err != nil {
			fmt.Println("unable to decode into struct, %v", err)
		}
		//perform nmap scanning with scanthreads number of threads.
		threads := viper.GetInt("scanthreads")
		w := NewBoundedWaitGroup(threads)
		for _, vrf := range C.Targets {
			for _, host := range GenerateIPs(vrf.Subnets) {
				w.Add(1)
				go ScanHost(vrf.Vrf, host, &w, ch)

				go func() {
					for v := range ch {
						results = append(results, v)
					}
				}()
			}
		}

		w.Wait()
		close(ch)
		client := netbox.NetboxLogin(viper.GetString("netbox.netboxapitoken"), viper.GetString("netbox.netboxhost"))
		netbox.GetIPAddresses(client)

		for _, host := range results {
			netbox.CreateIPAddress(client, host)
		}

	},
}

func GenerateIPs(subnet []string) []string {
	hosts := make([]string, 0)
	for _, prefix := range subnet {
		s := strings.Split(prefix, "/")
		prefixes, _ := Hosts(prefix)
		for _, host := range prefixes {
			hosts = append(hosts, host+"/"+s[1])
		}
	}
	return hosts

}

func ScanHost(vrf string, host string, w *BoundedWaitGroup, ch chan models.IPAddress) {
	defer w.Done()
	fmt.Println("start")
	s := strings.Split(host, "/")
	ipaddress := models.IPAddress{}
	scan := nmap.New()

	args := []string{"-sn"}
	scan.SetArgs(args...)
	scan.SetHosts(s[0])
	ipaddress.Address = &host
	if vrf != "" {
		var hostvrf models.NestedVRF
		hostvrf.Name = &vrf
		ipaddress.Vrf = &hostvrf
	}
	err := scan.Run()
	if err != nil {
		fmt.Println("Scan failed:", err)
		return
	}

	result, err := scan.Parse()
	if err != nil {
		fmt.Println("Parse scanner result:", err)
		return
	}

	for _, host := range result.Hosts {
		if host.Status.State == "up" {
			ch <- ipaddress
		}
	}

}

func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// remove network address and broadcast address
	lenIPs := len(ips)
	switch {
	case lenIPs < 2:
		return ips, nil

	default:
		return ips[1 : len(ips)-1], nil
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
func init() {
	rootCmd.AddCommand(scanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
