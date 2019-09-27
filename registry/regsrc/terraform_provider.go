package regsrc

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/hashicorp/terraform-svchost"
)

var (
	// DefaultProviderNamespace represents the namespace for canonical
	// HashiCorp-controlled providers.
	DefaultProviderNamespace = "-"
)

// TerraformProvider describes a Terraform Registry Provider source.
type TerraformProvider struct {
	RawHost      *FriendlyHost
	RawNamespace string
	RawName      string
	OS           string
	Arch         string
}

// NewTerraformProvider constructs a new provider source.
func NewTerraformProvider(name, os, arch, source string) *TerraformProvider {
	p := &TerraformProvider{}
	// TODO: sourceStr will be required and should be verified earlier on
	// I have no idea when or where
	_, namespace, sourceURL := parseProviderSourceStr(source)

	if os == "" {
		os = runtime.GOOS
	}
	if arch == "" {
		arch = runtime.GOARCH
	}

	p.RawHost = sourceURL
	p.RawNamespace = namespace
	p.RawName = name
	p.OS = os
	p.Arch = arch

	return p
}

// Provider returns just the registry ID of the provider
func (p *TerraformProvider) TerraformProvider() string {
	return fmt.Sprintf("%s/%s", p.RawNamespace, p.RawName)
}

// SvcHost returns the svchost.Hostname for this provider. The
// default PublicRegistryHost is returned.
func (p *TerraformProvider) SvcHost() (svchost.Hostname, error) {
	return svchost.ForComparison(p.RawHost.Raw)
}

func parseProviderSourceStr(source string) (name, namespace string, sourceURL *FriendlyHost) {
	// TODO in the "real world" source is required and should have already been validated
	if source == "" {
		return "", DefaultProviderNamespace, PublicRegistryHost
	}

	parts := strings.Split(source, "/")
	n := parts[len(parts)-1:]
	name = strings.Join(n, "")

	if len(parts) > 1 {
		ns := parts[len(parts)-2 : len(parts)-1]
		namespace = strings.Join(ns, "")
	} else {
		namespace = DefaultProviderNamespace
	}

	//TODO: the user-provided source url could have "/"
	// regscr.FriendlyHost might be insufficient?
	if len(parts) > 2 {
		s := parts[:len(parts)-2]
		sourceURL = NewFriendlyHost(strings.Join(s, "/"))
	} else {
		sourceURL = PublicRegistryHost
	}

	return
}
