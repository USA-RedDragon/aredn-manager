package olsrd

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
)

const hostsFile = "/var/run/hosts_olsr"

type HostsParser struct {
	currentHosts    []*AREDNHost
	arednNodesCount int
	isParsing       atomic.Bool
}

func NewHostsParser() *HostsParser {
	return &HostsParser{}
}

func (p *HostsParser) GetHosts() []*AREDNHost {
	return p.currentHosts
}

func (p *HostsParser) GetHostsCount() int {
	return len(p.currentHosts)
}

func (p *HostsParser) GetAREDNHostsCount() int {
	return p.arednNodesCount
}

func (p *HostsParser) GetHostsPaginated(page int, limit int, filter string) []*AREDNHost {
	ret := []*AREDNHost{}
	for _, host := range p.currentHosts {
		filter = strings.ToLower(filter)
		hostNameLower := strings.ToLower(host.Hostname)
		if strings.Contains(hostNameLower, filter) {
			ret = append(ret, host)
		}
	}
	start := (page - 1) * limit
	end := start + limit
	if start > len(ret) {
		return []*AREDNHost{}
	}
	if end > len(ret) {
		end = len(ret)
	}
	return ret[start:end]
}

func (p *HostsParser) Parse() (err error) {
	if p.isParsing.Load() {
		return
	}
	p.isParsing.Store(true)
	defer p.isParsing.Store(false)
	hosts, arednCount, err := parseHosts()
	if err != nil {
		return
	}
	p.arednNodesCount = arednCount
	p.currentHosts = hosts
	return
}

type HostData struct {
	Hostname string          `json:"hostname"`
	IP       net.IP          `json:"ip"`
	Services []*AREDNService `json:"services"`
}

type AREDNHost struct {
	HostData
	Children []HostData `json:"children"`
}

type orphans struct {
	ip     net.IP
	parent net.IP
}

func (h *AREDNHost) addChild(child HostData) {
	h.Children = append(h.Children, child)
}

func (h *AREDNHost) String() string {
	ret := fmt.Sprintf("%s: %s\n", h.Hostname, h.IP)
	for _, child := range h.Children {
		ret += fmt.Sprintf("\t%s: %s\n", child.Hostname, child.IP)
	}
	return ret
}

// ParseHosts parses the hosts file and returns a map of hostname to IP
//
// Text on a line after a # is ignored
// Lines with only whitespace or that are empty are ignored
//
//nolint:golint,gocyclo
func parseHosts() (ret []*AREDNHost, arednCount int, err error) {
	hostsFile, err := os.ReadFile(hostsFile)
	if err != nil {
		return
	}
	orphanedChildren := make(map[string]orphans)
	for _, line := range strings.Split(string(hostsFile), "\n") {
		// Ignore comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		var parentIP net.IP
		if strings.Contains(line, "#") {
			splt := strings.Split(line, "#")
			line = splt[0]
			splt[1] = strings.TrimSpace(splt[1])
			if splt[1] == "myself" {
				split2 := strings.Fields(line)
				if len(split2) != 2 {
					fmt.Printf("Invalid line in hosts file: %s\n", line)
					continue
				}
				split2[0] = strings.TrimSpace(split2[0])
				parentIP = net.ParseIP(strings.TrimSpace(split2[0]))
			} else {
				parentIP = net.ParseIP(strings.TrimSpace(splt[1]))
			}
		}

		// Ignore empty lines
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// Split the line into hostname and IP
		// Note that the separator may be any amount of whitespace or tabs
		split := strings.Fields(line)
		if len(split) != 2 {
			fmt.Printf("Invalid line in hosts file: %s\n", line)
			continue
		}

		split[0] = strings.TrimSpace(split[0])
		split[1] = strings.TrimSpace(split[1])

		if split[1] == "localhost" {
			continue
		}

		if strings.Contains(split[1], ".") {
			if regexp.MustCompile(`mid\d+\.`).MatchString(split[1]) {
				arednCount++
			}
			continue
		}

		if split[1] == "" {
			fmt.Printf("Invalid hostname in hosts file: %s\n", line)
			continue
		}

		ip := net.ParseIP(split[0])
		if ip == nil {
			fmt.Printf("Invalid IP in hosts file: %s\n", line)
			continue
		}

		// If the parentIP is not the same as the IP, then we need to treat this as a child
		if parentIP.String() != ip.String() {
			// Find the parent
			var parent *AREDNHost
			for _, host := range ret {
				if host.IP.Equal(parentIP) {
					parent = host
					break
				}
			}
			if parent == nil {
				orphanedChildren[split[1]] = orphans{
					ip:     ip,
					parent: parentIP,
				}
				continue
			}
			parent.addChild(HostData{
				Hostname: split[1],
				IP:       ip,
			})
			continue
		}

		host := &AREDNHost{
			HostData: HostData{
				Hostname: split[1],
				IP:       ip,
			},
		}

		// Search orphanedChildren for children of this host
		orphansToRemove := []HostData{}
		for hostname, orphan := range orphanedChildren {
			if orphan.parent.Equal(ip) {
				child := HostData{
					Hostname: hostname,
					IP:       orphan.ip,
				}
				host.addChild(child)
				orphansToRemove = append(orphansToRemove, child)
			}
		}
		for _, orphan := range orphansToRemove {
			delete(orphanedChildren, orphan.Hostname)
		}

		// Add the hostname and IP to the map
		ret = append(ret, host)
	}

	svcs := NewServicesParser()
	err = svcs.Parse()
	if err != nil {
		fmt.Printf("Error parsing services: %v\n", err)
		return
	}

	services := svcs.GetServices()
	foundServices := []*AREDNService{}

	// We need to go through the hosts and each of their children and add the services
	// Remove the services from the servicesCopy list as we find them
	for hostIdx, host := range ret {
		for childIdx, child := range host.Children {
			for _, svc := range services {
				url, err := url.Parse(svc.URL)
				if err != nil {
					fmt.Printf("Error parsing URL: %v\n", err)
					continue
				}

				serviceAlreadyFound := false
				for _, foundSvc := range foundServices {
					if foundSvc == svc {
						serviceAlreadyFound = true
					}
				}
				if serviceAlreadyFound {
					continue
				}
				if strings.EqualFold(child.Hostname, url.Hostname()) {
					ret[hostIdx].Children[childIdx].Services = append(ret[hostIdx].Children[childIdx].Services, svc)
					foundServices = append(foundServices, svc)
				}
			}
		}
		for _, svc := range services {
			url, err := url.Parse(svc.URL)
			if err != nil {
				fmt.Printf("Error parsing URL: %v\n", err)
				continue
			}
			serviceAlreadyFound := false
			for _, foundSvc := range foundServices {
				if foundSvc == svc {
					serviceAlreadyFound = true
				}
			}
			if serviceAlreadyFound {
				continue
			}
			if strings.EqualFold(host.Hostname, url.Hostname()) {
				ret[hostIdx].Services = append(ret[hostIdx].Services, svc)
				foundServices = append(foundServices, svc)
			}
		}
	}

	return
}
