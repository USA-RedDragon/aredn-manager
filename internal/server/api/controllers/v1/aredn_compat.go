package v1

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/USA-RedDragon/aredn-manager/internal/sdk"
	"github.com/USA-RedDragon/aredn-manager/internal/server/api/apimodels"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
			Enabled: true,
		},
		Interfaces: getInterfaces(),
		Hosts:      getHosts(),
		Services:   getServices(),
		LinkInfo:   getLinkInfo(),
	}

	c.JSON(http.StatusOK, sysinfo)
}

func getInterfaces() []apimodels.Interface {
	ret := []apimodels.Interface{}

	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to get interfaces: %v", err)
		return nil
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
	return ret
}

func getHosts() []apimodels.Host {
	regexMid, err := regexp.Compile("^mid\\d+\\..*")
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to compile regex: %v", err)
		return nil
	}
	regexDtd, err := regexp.Compile("^dtdlink\\..*")
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to compile regex: %v", err)
		return nil
	}
	parser := olsrd.NewHostsParser()
	err = parser.Parse()
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to parse hosts file: %v", err)
		return nil
	}
	hosts := parser.GetHosts()
	ret := []apimodels.Host{}
	for _, host := range hosts {
		match := regexMid.Match([]byte(host.Hostname))
		if match {
			continue
		}

		match = regexDtd.Match([]byte(host.Hostname))
		if match {
			continue
		}

		ret = append(ret, apimodels.Host{
			Name: host.Hostname,
			IP:   host.IP.String(),
		})
	}
	return ret
}

func getLinkInfo() map[string]apimodels.LinkInfo {
	ret := make(map[string]apimodels.LinkInfo)
	// http request http://localhost:9090/links
	resp, err := http.DefaultClient.Get("http://localhost:9090/links")
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to get links: %v", err)
		return nil
	}
	defer resp.Body.Close()
	// Grab the body as json
	var links apimodels.OlsrdLinks
	err = json.NewDecoder(resp.Body).Decode(&links)
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to decode links: %v", err)
		return nil
	}

	for _, link := range links.Links {
		// we need to take the hostname from the URL and resolve it to an IP
		ips, err := net.LookupIP(link.Hostname)
		if err != nil {
			fmt.Printf("GETSysinfo: Unable to resolve hostname: %v", err)
			continue
		}

		ret[ips[0].String()] = apimodels.LinkInfo{
			HelloTime:           link.HelloTime,
			LostLinkTime:        link.LostLinkTime,
			LinkQuality:         link.LinkQuality,
			VTime:               link.VTime,
			LinkCost:            link.LinkCost,
			LinkType:            link.LinkType,
			Hostname:            link.Hostname,
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
		}
	}

	return ret
}

func getServices() []apimodels.Service {
	parser := olsrd.NewServicesParser()
	err := parser.Parse()
	if err != nil {
		fmt.Printf("GETSysinfo: Unable to parse services file: %v", err)
		return nil
	}
	svcs := parser.GetServices()
	ret := []apimodels.Service{}
	for _, svc := range svcs {
		// we need to take the hostname from the URL and resolve it to an IP
		url, err := url.Parse(svc.URL)
		if err != nil {
			fmt.Printf("GETSysinfo: Unable to parse URL: %v", err)
			continue
		}
		ips, err := net.LookupIP(url.Hostname())
		if err != nil {
			fmt.Printf("GETSysinfo: Unable to resolve hostname: %v", err)
			continue
		}
		link := svc.URL
		// If the link ends with :0/, then it is a non-http link, so set link to ""
		if strings.HasSuffix(svc.URL, ":0/") {
			link = ""
		}
		ret = append(ret, apimodels.Service{
			Name:     svc.Name,
			IP:       ips[0].String(),
			Protocol: svc.Protocol,
			Link:     link,
		})
	}

	return ret
}
