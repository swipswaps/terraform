package configs

import (
	"fmt"
	"log"

	version "github.com/hashicorp/go-version"
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
	attrs, diags := block.Body.JustAttributes()
	var reqs []*ProviderRequirement
	for name, attr := range attrs {
		expr, err := attr.Expr.Value(nil)
		if err != nil {
			log.Printf("[TRACE] expr err in decodeRequiredProvidersBlock: %s\n", err.Error())
			panic("buhbye")
		}
		if expr.Type().IsPrimitiveType() {
			req, reqDiags := decodeVersionConstraint(attr)
			diags = append(diags, reqDiags...)
			if !diags.HasErrors() {
				reqs = append(reqs, &ProviderRequirement{
					Name:               name,
					VersionConstraints: []VersionConstraint{req},
				})
			}
		} else if expr.Type().IsObjectType() {
			// This is incomplete: the "name" here is the user-supplied map key, not the type name
			pr := &ProviderRequirement{Name: name}
			if expr.Type().HasAttribute("version") {
				constraintStr, err := version.NewConstraint(expr.GetAttr("version").AsString())
				if err != nil {
					// NewConstraint doesn't return user-friendly errors, so we'll just
					// ignore the provided error and produce our own generic one.
					versionDiags := &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid version constraint",
						Detail:   "This string does not use correct version constraint syntax.", // Not very actionable :(
						Subject:  attr.Expr.Range().Ptr(),
					}
					diags = append(diags, versionDiags)
				}
				vc := VersionConstraint{
					DeclRange: attr.Range,
					Required:  constraintStr,
				}
				pr.VersionConstraints = append(pr.VersionConstraints, vc)
			}
			if expr.Type().HasAttribute("source") {
				pr.Source = expr.GetAttr("source").AsString()
			}
			reqs = append(reqs, pr)
		}
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
