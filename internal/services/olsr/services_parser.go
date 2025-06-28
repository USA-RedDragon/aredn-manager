package olsr

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
)

const servicesFile = "/var/run/services_olsr"

type MeshService struct {
	URL        string `json:"url"`
	Protocol   string `json:"protocol"`
	Name       string `json:"name"`
	ShouldLink bool   `json:"should_link"`
}

func (s *MeshService) String() string {
	ret := ""
	ret += fmt.Sprintf("%s:\n\t", s.Name)
	ret += fmt.Sprintf("%s\t%s", s.Protocol, s.URL)
	return ret
}

type ServicesParser struct {
	currentServices []*MeshService
}

func NewServicesParser() *ServicesParser {
	return &ServicesParser{}
}

func (p *ServicesParser) GetServicesCount() int {
	return len(p.currentServices)
}

func (p *ServicesParser) Parse() (err error) {
	services, err := parseServices()
	if err != nil {
		return
	}
	p.currentServices = services
	return
}

func (p *ServicesParser) GetServices() []*MeshService {
	return p.currentServices
}

func parseServices() (ret []*MeshService, err error) {
	servicesFile, err := os.ReadFile(servicesFile)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(servicesFile), "\n") {
		// Ignore comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.Contains(line, "#") {
			line = strings.Split(line, "#")[0]
		}

		line = strings.TrimSpace(line)

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

		service := &MeshService{
			URL:        url.String(),
			Protocol:   split[1],
			Name:       split[2],
			ShouldLink: url.Port() != "0",
		}

		// Add the hostname and IP to the map
		ret = append(ret, service)
	}

	return
}
