package arednlink

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
	needParse       atomic.Bool // set if we get a call to Parse() while already parsing. We'll run Parse() again after the current parse is done to ensure we have the latest data
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
		p.needParse.Store(true)
		return
	}
	p.isParsing.Store(true)
	hosts, arednCount, totalCount, serviceCount, err := parseHosts()
	if err != nil {
		return
	}
	p.isParsing.Store(false)
	p.arednNodesCount = arednCount
	p.totalCount = totalCount
	p.currentHosts = hosts
	p.serviceCount = serviceCount
	if p.needParse.Load() {
		go func() {
			p.needParse.Store(false)
			p.isParsing.Store(true)
			defer p.isParsing.Store(false)
			if err := p.Parse(); err != nil {
				slog.Error("Error re-parsing hosts", "error", err)
			}
		}()
	}
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
	Tag        string `json:"type"`
}

func (s *AREDNService) String() string {
	ret := fmt.Sprintf("%s:\n\t", s.Name)
	ret += fmt.Sprintf("%s\t%s", s.Protocol, s.URL)
	return ret
}

type AREDNHost struct {
	HostData
	Children []HostData `json:"children"`
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

var (
	regexAredn  = regexp.MustCompile(`\s[^\.]+$`)
	taggedRegex = regexp.MustCompile(`^(.*)\s+\[(.*)\]$`)
)

func parseHosts() (ret []*AREDNHost, arednCount int, totalCount int, serviceCount int, err error) {
	err = fs.WalkDir(os.DirFS(hostsDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("Error reading hosts directory", "error", err)
			return err
		}
		if d.IsDir() {
			return nil // Skip directories
		}

		file := filepath.Join(hostsDir, path)

		slog.Debug("parseHosts: Processing hosts file", "file", file)

		entries, err := os.ReadFile(file)
		if err != nil {
			slog.Error("parseHosts: Error reading hosts directory entry", "entry", file, "error", err)
		}

		totalCount++

		var arednHost *AREDNHost

		for _, line := range strings.Split(string(entries), "\n") {
			// Ignore empty lines
			if len(strings.TrimSpace(line)) == 0 {
				continue
			}

			slog.Debug("parseHosts: Processing hosts file line", "file", file, "line", line)

			fields := strings.Fields(line)
			if len(fields) < 2 {
				slog.Warn("parseHosts: Invalid AREDN host entry", "file", file, "line", line)
				continue
			}

			if regexAredn.Match([]byte(line)) && arednHost == nil {
				slog.Debug("parseHosts: Found AREDN host entry", "file", file, "line", line)
				arednHost = &AREDNHost{
					HostData: HostData{
						Hostname: strings.TrimSpace(fields[1]),
						IP:       net.ParseIP(strings.TrimSpace(fields[0])),
					},
				}
				if arednHost.IP == nil {
					slog.Warn("parseHosts: Invalid IP in hosts file", "file", file, "line", line)
					continue
				}
				// Check if the same base filename exists under the services directory
				servicesFile := filepath.Join(servicesDir, arednHost.IP.To4().String())
				if services, err := os.ReadFile(servicesFile); err == nil {
					slog.Debug("parseHosts: Found services file for AREDN host", "file", servicesFile)
					var servicesList []*AREDNService
					for _, svcLine := range strings.Split(string(services), "\n") {
						line := strings.TrimSpace(svcLine)

						// Ignore empty lines
						if len(line) == 0 {
							slog.Debug("parseHosts: Skipping empty line in services file", "file", servicesFile)
							continue
						}

						// Lines are of the form:
						// url|protocol|name

						// Split the line by '|'
						split := strings.Split(line, "|")
						if len(split) != 3 {
							slog.Warn("parseHosts: Invalid service line format", "line", line, "file", servicesFile)
							continue
						}

						url, err := url.Parse(split[0])
						if err != nil {
							slog.Warn("parseHosts: Error parsing URL", "url", split[0], "error", err)
							continue
						}

						// Name can have an optional tag suffix like 'Meshchat [chat]'
						name := split[2]
						tag := ""
						if matches := taggedRegex.FindStringSubmatch(name); len(matches) == 3 {
							name = matches[1]
							tag = matches[2]
						}

						service := &AREDNService{
							URL:        url.String(),
							Protocol:   split[1],
							Name:       name,
							ShouldLink: url.Port() != "0",
							Tag:        tag,
						}

						serviceCount++

						servicesList = append(servicesList, service)
					}
					arednHost.HostData.Services = append(arednHost.HostData.Services, servicesList...)
				}
			} else {
				slog.Debug("parseHosts: Found child host entry", "file", file, "line", line)
				if arednHost == nil {
					slog.Warn("parseHosts: Found a host entry without a parent AREDN host", "line", line)
					continue
				}
				// This is a child of the last AREDN host
				child := HostData{
					Hostname: strings.TrimSuffix(strings.TrimSpace(fields[1]), ".local.mesh"),
					IP:       net.ParseIP(strings.TrimSpace(fields[0])),
				}
				if child.IP == nil {
					slog.Warn("parseHosts: Invalid IP in hosts file", "line", line)
					continue
				}
				if strings.HasPrefix(child.Hostname, "lan.") ||
					strings.HasPrefix(child.Hostname, "dtdlink.") ||
					strings.HasPrefix(child.Hostname, "babel.") ||
					strings.HasPrefix(child.Hostname, "supernode.") {
					continue
				}
				arednHost.addChild(child)
			}
		}

		if arednHost != nil {
			ret = append(ret, arednHost)
			arednCount++
		}
		return nil
	})

	return
}
