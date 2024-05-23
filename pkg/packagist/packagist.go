package packagist

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type PackagistResponse struct {
	PackageNames []string `json:"packageNames"`
}

type PackageResponse struct {
	Package Package `json:"package"`
}

type Package struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Time        time.Time `json:"time"`
	Type        string    `json:"type"`
	Repository  string    `json:"repository"`
	Downloads   Downloads `json:"downloads"`
	Favers      int       `json:"favers"`
}

type Downloads struct {
	Total   int `json:"total"`
	Monthly int `json:"monthly"`
	Daily   int `json:"daily"`
}

type Packagist struct {
	Vendor   string
	client   *http.Client
	Packages []string
}

func NewPackagist(vendor string) *Packagist {
	return &Packagist{
		Vendor: vendor,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *Packagist) FetchPackages() (*Packagist, error) {
	url := fmt.Sprintf("https://packagist.org/packages/list.json?vendor=%s", p.Vendor)

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch packages: %w", err)
	}
	defer resp.Body.Close()

	var response PackagistResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	p.Packages = response.PackageNames
	return p, nil
}

func (p *Packagist) FetchDetails(pkg string) (*Package, error) {
	url := fmt.Sprintf("https://packagist.org/packages/%v.json", pkg)

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package details: %w", err)
	}
	defer resp.Body.Close()

	var response PackageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response.Package, nil
}
