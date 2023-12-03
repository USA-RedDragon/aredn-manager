package v1

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/USA-RedDragon/aredn-manager/internal/sdk"
	"github.com/USA-RedDragon/aredn-manager/internal/server/api/apimodels"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GETMetrics(c *gin.Context) {
	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		fmt.Println("GETMetrics: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	if config.MetricsPort == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "Metrics port is not configured"})
		return
	}

	nodeResp, err := http.DefaultClient.Get(fmt.Sprintf("http://%s:9100/metrics", config.MetricsNodeExporterHost))
	if err != nil {
		fmt.Printf("GETMetrics: Unable to get node-exporter metrics: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	defer nodeResp.Body.Close()
	nodeMetrics := ""
	buf := make([]byte, 128)
	n, err := nodeResp.Body.Read(buf)
	for err == nil || n > 0 {
		nodeMetrics += string(buf[:n])
		n, err = nodeResp.Body.Read(buf)
	}

	metricsResp, err := http.DefaultClient.Get(fmt.Sprintf("http://localhost:%d/metrics", config.MetricsPort))
	if err != nil {
		fmt.Printf("GETMetrics: Unable to get metrics: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	defer metricsResp.Body.Close()
	metrics := ""
	buf = make([]byte, 128)
	n, err = metricsResp.Body.Read(buf)
	for err == nil || n > 0 {
		metrics += string(buf[:n])
		n, err = metricsResp.Body.Read(buf)
	}

	// Combine the two responses and send them back
	c.String(http.StatusOK, fmt.Sprintf("%s\n%s", nodeMetrics, metrics))
}

func GETSysinfo(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Println("POSTLogin: Unable to get DB from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		fmt.Println("GETSysinfo: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	activeTunnels, err := models.CountActiveTunnels(db)
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to get active tunnels: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	var info syscall.Sysinfo_t
	err = syscall.Sysinfo(&info)
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to get system info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	interfacesChan := make(chan []apimodels.Interface)
	hostsChan := make(chan []apimodels.Host)
	servicesChan := make(chan []apimodels.Service)
	linkInfoChan := make(chan map[string]apimodels.LinkInfo)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		getInterfaces(interfacesChan)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		getHosts(hostsChan)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		getServices(servicesChan)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		getLinkInfo(linkInfoChan)
		wg.Done()
	}()

	wg.Wait()

	sysinfo := apimodels.SysinfoResponse{
		Longitude: config.Longitude,
		Latitude:  config.Latitude,
		Sysinfo: apimodels.Sysinfo{
			Uptime: secondsToClock(info.Uptime),
			Loadavg: [3]float64{
				float64(info.Loads[0]) / float64(1<<16),
				float64(info.Loads[1]) / float64(1<<16),
				float64(info.Loads[2]) / float64(1<<16),
			},
		},
		APIVersion: "1.11",
		MeshRF: apimodels.MeshRF{
			Status: "off",
		},
		Gridsquare: config.Gridsquare,
		Node:       config.ServerName,
		NodeDetails: apimodels.NodeDetails{
			MeshSupernode:        config.Supernode,
			Description:          "AREDN Cloud Tunnel",
			Model:                "Virtual",
			MeshGateway:          "1",
			BoardID:              "0x0000",
			FirmwareManufacturer: "github.com/USA-RedDragon/aredn-manager",
			FirmwareVersion:      sdk.Version,
		},
		Tunnels: apimodels.Tunnels{
			ActiveTunnelCount: activeTunnels,
		},
		LQM: apimodels.LQM{
			Enabled: false,
		},
		Interfaces: <-interfacesChan,
		Hosts:      <-hostsChan,
		Services:   <-servicesChan,
		LinkInfo:   <-linkInfoChan,
	}

	c.JSON(http.StatusOK, sysinfo)

	close(interfacesChan)
	close(hostsChan)
	close(servicesChan)
	close(linkInfoChan)
}

func getInterfaces(output chan []apimodels.Interface) {
	ret := []apimodels.Interface{}

	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to get interfaces: %v", err)
		return
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Printf("GETSysinfo: Unable to get addresses for interface %s: %v", iface.Name, err)
			continue
		}
		if iface.Name == "lo" || strings.HasPrefix(iface.Name, "wg") {
			continue
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				fmt.Printf("GETSysinfo: Unable to parse address %s: %v", addr.String(), err)
				continue
			}
			ret = append(ret, apimodels.Interface{
				Name: iface.Name,
				IP:   ip.String(),
				MAC:  iface.HardwareAddr.String(),
			})
		}
	}
	output <- ret
}

var (
	regexMid = regexp.MustCompile(`^mid\d+\..*`)
	regexDtd = regexp.MustCompile(`^dtdlink\..*`)
)

func getHosts(output chan []apimodels.Host) {
	parser := olsrd.NewHostsParser()
	err := parser.Parse()
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to parse hosts file: %v", err)
		return
	}
	hosts := parser.GetHosts()
	ret := []apimodels.Host{}
	chans := make([]chan []string, len(hosts))
	wg := sync.WaitGroup{}

	for i, host := range hosts {
		wg.Add(1)
		chans[i] = make(chan []string)
		go func(host *olsrd.AREDNHost, output chan []string) {
			match := regexMid.Match([]byte(host.Hostname))
			if match {
				wg.Done()
				return
			}

			match = regexDtd.Match([]byte(host.Hostname))
			if match {
				wg.Done()
				return
			}

			output <- []string{host.Hostname, host.IP.String()}

			wg.Done()
		}(host, chans[i])
	}

	wg.Wait()

	for _, ch := range chans {
		a := <-ch
		ret = append(ret, apimodels.Host{
			Name: a[0],
			IP:   a[1],
		})
		close(ch)
	}
	output <- ret
}

func getLinkInfo(output chan map[string]apimodels.LinkInfo) {
	ret := make(map[string]apimodels.LinkInfo)
	// http request http://localhost:9090/links
	resp, err := http.DefaultClient.Get("http://localhost:9090/links")
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to get links: %v\n", err)
		return
	}
	defer resp.Body.Close()
	// Grab the body as json
	var links apimodels.OlsrdLinks
	err = json.NewDecoder(resp.Body).Decode(&links)
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to decode links: %v\n", err)
		return
	}

	wg := sync.WaitGroup{}
	type linkInfo struct {
		LinkInfo apimodels.LinkInfo
		IP       string
	}
	chans := make([]chan linkInfo, len(links.Links))

	for i, link := range links.Links {
		wg.Add(1)
		chans[i] = make(chan linkInfo)
		go func(link apimodels.OlsrdLinkinfo, output chan linkInfo) {
			hosts, err := net.LookupAddr(link.RemoteIP)
			if err != nil {
				fmt.Printf("GETSysinfo: Unable to resolve hostname: %s\n%v\n", link.RemoteIP, err)
				wg.Done()
				return
			}

			hostname := ""
			if len(hosts) > 0 {
				hostname = hosts[0]
				// Strip off mid\d. from the hostname if it exists
				regex := regexp.MustCompile(`^[mM][iI][dD]\d+\.(.+)`)
				matches := regex.FindStringSubmatch(hostname)
				if len(matches) == 2 {
					hostname = matches[1]
				}
				// Strip off dtdlink. from the hostname if it exists
				regex = regexp.MustCompile(`^[dD][tT][dD][lL][iI][nN][kK]\.(.+)`)
				matches = regex.FindStringSubmatch(hostname)
				if len(matches) == 2 {
					hostname = matches[1]
				}
				// Make sure the hostname doesn't end with a period
				hostname = strings.TrimSuffix(hostname, ".")
				// Make sure the hostname doesn't end with .local.mesh
				hostname = strings.TrimSuffix(hostname, ".local.mesh")
			} else {
				wg.Done()
				return
			}

			ips, err := net.LookupIP(hostname)
			if err != nil {
				fmt.Printf("GETSysinfo: Unable to resolve hostname: %s\n%v\n", hostname, err)
				wg.Done()
				return
			}

			if len(ips) == 0 {
				wg.Done()
				return
			}

			linkType := ""
			if strings.HasPrefix(link.OLSRInterface, "tun") {
				linkType = "TUN"
			} else if strings.HasPrefix(link.OLSRInterface, "eth") {
				linkType = "DTD"
			} else {
				linkType = "UNKNOWN"
			}

			output <- linkInfo{
				LinkInfo: apimodels.LinkInfo{
					HelloTime:           link.HelloTime,
					LostLinkTime:        link.LostLinkTime,
					LinkQuality:         link.LinkQuality,
					VTime:               link.VTime,
					LinkCost:            link.LinkCost,
					LinkType:            linkType,
					Hostname:            hostname,
					PreviousLinkStatus:  link.PreviousLinkStatus,
					CurrentLinkStatus:   link.CurrentLinkStatus,
					NeighborLinkQuality: link.NeighborLinkQuality,
					SymmetryTime:        link.SymmetryTime,
					SeqnoValid:          link.SeqnoValid,
					Pending:             link.Pending,
					LossHelloInterval:   link.LossHelloInterval,
					LossMultiplier:      link.LossMultiplier,
					Hysteresis:          link.Hysteresis,
					Seqno:               link.Seqno,
					LossTime:            link.LossTime,
					ValidityTime:        link.ValidityTime,
					OLSRInterface:       link.OLSRInterface,
					LastHelloTime:       link.LastHelloTime,
					AsymmetryTime:       link.AsymmetryTime,
				},
				IP: ips[0].String(),
			}
			wg.Done()
		}(link, chans[i])
	}

	wg.Wait()

	for _, ch := range chans {
		link := <-ch
		ret[link.IP] = link.LinkInfo
		close(ch)
	}

	output <- ret
}

func getServices(output chan []apimodels.Service) {
	parser := olsrd.NewServicesParser()
	err := parser.Parse()
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to parse services file: %v\n", err)
		return
	}
	svcs := parser.GetServices()
	ret := []apimodels.Service{}
	chans := make([]chan apimodels.Service, len(svcs))
	wg := sync.WaitGroup{}
	for i, svc := range svcs {
		wg.Add(1)
		chans[i] = make(chan apimodels.Service)
		go func(svc *olsrd.AREDNService, output chan apimodels.Service) {
			// we need to take the hostname from the URL and resolve it to an IP
			url, err := url.Parse(svc.URL)
			if err != nil {
				fmt.Printf("GETSysinfo: Unable to parse URL: %v\n", err)
				wg.Done()
				return
			}
			ips, err := net.LookupIP(url.Hostname())
			if err != nil {
				fmt.Printf("GETSysinfo: Unable to resolve hostname: %s\n%v\n", url.Hostname(), err)
				wg.Done()
				return
			}
			link := svc.URL
			// If the link ends with :0/, then it is a non-http link, so set link to ""
			if strings.HasSuffix(svc.URL, ":0/") {
				link = ""
			}
			output <- apimodels.Service{
				Name:     svc.Name,
				IP:       ips[0].String(),
				Protocol: svc.Protocol,
				Link:     link,
			}
			wg.Done()
		}(svc, chans[i])
	}

	wg.Wait()

	for _, ch := range chans {
		ret = append(ret, <-ch)
		close(ch)
	}

	output <- ret
}
