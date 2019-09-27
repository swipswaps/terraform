package configs

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"

	"github.com/hashicorp/terraform/addrs"
)

// Provider represents a "provider" block in a module or file. A provider
// block is a provider configuration, and there can be zero or more
// configurations for each actual provider.
type Provider struct {
	Name       string
	NameRange  hcl.Range
	Alias      string
	AliasRange *hcl.Range // nil if no alias set

	Version VersionConstraint

	Config hcl.Body

	DeclRange hcl.Range
}

func decodeProviderBlock(block *hcl.Block) (*Provider, hcl.Diagnostics) {
	content, config, diags := block.Body.PartialContent(providerBlockSchema)

	provider := &Provider{
		Name:      block.Labels[0],
		NameRange: block.LabelRanges[0],
		Config:    config,
		DeclRange: block.DefRange,
	}

	if attr, exists := content.Attributes["alias"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &provider.Alias)
		diags = append(diags, valDiags...)
		provider.AliasRange = attr.Expr.Range().Ptr()

		if !hclsyntax.ValidIdentifier(provider.Alias) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid provider configuration alias",
				Detail:   fmt.Sprintf("An alias must be a valid name. %s", badIdentifierDetail),
			})
		}
	}

	if attr, exists := content.Attributes["version"]; exists {
		var versionDiags hcl.Diagnostics
		provider.Version, versionDiags = decodeVersionConstraint(attr)
		diags = append(diags, versionDiags...)
	}

	// Reserved attribute names
	for _, name := range []string{"count", "depends_on", "for_each", "source"} {
		if attr, exists := content.Attributes[name]; exists {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Reserved argument name in provider block",
				Detail:   fmt.Sprintf("The provider argument name %q is reserved for use by Terraform in a future version.", name),
				Subject:  &attr.NameRange,
			})
		}
	}

	// Reserved block types (all of them)
	for _, block := range content.Blocks {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Reserved block type name in provider block",
			Detail:   fmt.Sprintf("The block type name %q is reserved for use by Terraform in a future version.", block.Type),
			Subject:  &block.TypeRange,
		})
	}

	return provider, diags
}

// Addr returns the address of the receiving provider configuration, relative
// to its containing module.
func (p *Provider) Addr() addrs.ProviderConfig {
	return addrs.ProviderConfig{
		Type:  p.Name,
		Alias: p.Alias,
	}
}

func (p *Provider) moduleUniqueKey() string {
	if p.Alias != "" {
		return fmt.Sprintf("%s.%s", p.Name, p.Alias)
	}
	return p.Name
}

// ProviderRequirement represents a declaration of a dependency on a particular
// provider version and source without actually configuring that provider.
// TODO: Add ranges for diagnostics
type ProviderRequirement struct {
	Name               string
	Source             string
	VersionConstraints []VersionConstraint
}

func decodeRequiredProvidersBlock(block *hcl.Block) ([]*ProviderRequirement, hcl.Diagnostics) {
	content, _, diags := block.Body.PartialContent(providerRequirementBlockSchema)
	reqs := make([]*ProviderRequirement, len(content.Blocks))

	for _, block := range content.Blocks {
		pr := &ProviderRequirement{
			Name: block.Labels[0],
		}
		if attr, exists := content.Attributes["version"]; exists {
			version, versionDiags := decodeVersionConstraint(attr)
			pr.VersionConstraints = append(pr.VersionConstraints, version)
			diags = append(diags, versionDiags...)
		}
		if attr, exists := content.Attributes["source"]; exists {
			sourceDiags := pr.decodeProviderSource(attr)
			diags = append(diags, sourceDiags...)
		}
		reqs = append(reqs, pr)
	}

	return reqs, diags
}

func (pr *ProviderRequirement) decodeProviderSource(attr *hcl.Attribute) (diags hcl.Diagnostics) {
	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		diags = append(diags, diags...)
		return
	}
	var err error
	val, err = convert.Convert(val, cty.String)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid source constraint",
			Detail:   fmt.Sprintf("A string value is required for %s.", attr.Name),
			Subject:  attr.Expr.Range().Ptr(),
		})
		return
	}
	pr.Source = val.AsString()
	return
}

var providerBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "alias",
		},
		{
			Name: "version",
		},

		// Attribute names reserved for future expansion.
		{Name: "count"},
		{Name: "depends_on"},
		{Name: "for_each"},
		{Name: "source"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		// _All_ of these are reserved for future expansion.
		{Type: "lifecycle"},
		{Type: "locals"},
	},
}

var providerRequirementBlockSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "provider", LabelNames: []string{"name"}},
	},
}
