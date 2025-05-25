package babel

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
)

const hostsDir = "/var/run/arednlink/hosts"
const servicesDir = "/var/run/arednlink/services"

type Parser struct {
	currentHosts    []*AREDNHost
	arednNodesCount int
	totalCount      int
	serviceCount    int
	isParsing       atomic.Bool
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) GetHosts() []*AREDNHost {
	return p.currentHosts
}

func (p *Parser) GetHostsCount() int {
	return len(p.currentHosts)
}

func (p *Parser) GetServiceCount() int {
	return p.serviceCount
}

func (p *Parser) GetAREDNHostsCount() int {
	return p.arednNodesCount
}

func (p *Parser) GetTotalHostsCount() int {
	return p.totalCount + p.arednNodesCount
}

func (p *Parser) GetHostsPaginated(page int, limit int, filter string) []*AREDNHost {
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

func (p *Parser) Parse() (err error) {
	if p.isParsing.Load() {
		return
	}
	p.isParsing.Store(true)
	defer p.isParsing.Store(false)
	hosts, arednCount, totalCount, serviceCount, err := parseHosts()
	if err != nil {
		return
	}
	p.arednNodesCount = arednCount
	p.totalCount = totalCount
	p.currentHosts = hosts
	p.serviceCount = serviceCount
	return
}

type HostData struct {
	Hostname string          `json:"hostname"`
	IP       net.IP          `json:"ip"`
	Services []*AREDNService `json:"services"`
}

type AREDNService struct {
	URL        string `json:"url"`
	Protocol   string `json:"protocol"`
	Name       string `json:"name"`
	ShouldLink bool   `json:"should_link"`
}

func (s *AREDNService) String() string {
	ret := ""
	ret += fmt.Sprintf("%s:\n\t", s.Name)
	ret += fmt.Sprintf("%s\t%s", s.Protocol, s.URL)
	return ret
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

func parseHosts() (ret []*AREDNHost, arednCount int, totalCount int, serviceCount int, err error) {
	regexAredn := regexp.MustCompile(`\s[^\.]+$`)
	err = fs.WalkDir(os.DirFS(hostsDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("Error reading hosts directory", "error", err)
			return err
		}
		if d.IsDir() {
			return nil // Skip directories
		}

		file := filepath.Join(hostsDir, path)

		entries, err := os.ReadFile(file)
		if err != nil {
			slog.Error("Error reading hosts directory entry", "entry", file, "error", err)
		}

		totalCount++

		var arednHost *AREDNHost

		for _, line := range strings.Split(string(entries), "\n") {
			// Ignore empty lines
			if len(strings.TrimSpace(line)) == 0 {
				continue
			}

			if regexAredn.Match([]byte(line)) {
				arednCount++
				arednHost = &AREDNHost{
					HostData: HostData{
						Hostname: strings.TrimSpace(strings.Fields(line)[0]),
						IP:       net.ParseIP(strings.TrimSpace(strings.Fields(line)[1])),
					},
				}
				if arednHost.IP == nil {
					slog.Warn("Invalid IP in hosts file", "line", line)
					continue
				}
				// Check if the same base filename exists under the services directory
				servicesFile := filepath.Join(servicesDir, arednHost.IP.To4().String())
				if services, err := os.ReadFile(servicesFile); err == nil {
					slog.Debug("Found services file for AREDN host", "file", servicesFile)
					var servicesList []*AREDNService
					for _, svcLine := range strings.Split(string(services), "\n") {
						line := strings.TrimSpace(svcLine)

						// Ignore empty lines
						if len(line) == 0 {
							continue
						}

						// Lines are of the form:
						// url|protocol|name

						// Split the line by '|'
						split := strings.Split(line, "|")
						if len(split) != 3 {
							continue
						}

						url, err := url.Parse(split[0])
						if err != nil {
							slog.Warn("Error parsing URL", "url", split[0], "error", err)
							continue
						}

						service := &AREDNService{
							URL:        url.String(),
							Protocol:   split[1],
							Name:       split[2],
							ShouldLink: url.Port() != "0",
						}

						serviceCount++

						servicesList = append(servicesList, service)
					}
				}
			} else {
				if arednHost == nil {
					slog.Warn("Found a host entry without a parent AREDN host", "line", line)
					continue
				}
				// This is a child of the last AREDN host
				child := HostData{
					Hostname: strings.TrimSpace(strings.Fields(line)[0]),
					IP:       net.ParseIP(strings.TrimSpace(strings.Fields(line)[1])),
				}
				if child.IP == nil {
					slog.Warn("Invalid IP in hosts file", "line", line)
					continue
				}
				arednHost.addChild(child)
			}
		}

		if arednHost != nil {
			ret = append(ret, arednHost)
		}
		return nil
	})

	return
}
