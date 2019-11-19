package tfconfig

import (
	"fmt"
	"log"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// ProviderRef is a reference to a provider configuration within a module.
// It represents the contents of a "provider" argument in a resource, or
// a value in the "providers" map for a module call.
type ProviderRef struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitempty"` // Empty if the default provider configuration is referenced
}

type ProviderRequirement struct {
	Name               string
	Source             string
	VersionConstraints []VersionConstraint
}

type VersionConstraint struct {
	Required  version.Constraints
	DeclRange hcl.Range
}

func decodeRequiredProvidersBlock(block *hcl.Block) (map[string]*ProviderRequirement, hcl.Diagnostics) {
	attrs, diags := block.Body.JustAttributes()
	reqs := make(map[string]*ProviderRequirement)
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
				reqs[name] = &ProviderRequirement{
					Name:               name,
					VersionConstraints: []VersionConstraint{req},
				}
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
			reqs[name] = pr
		}
	}

	return reqs, diags
}

func decodeVersionConstraint(attr *hcl.Attribute) (VersionConstraint, hcl.Diagnostics) {
	ret := VersionConstraint{
		DeclRange: attr.Range,
	}

	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return ret, diags
	}
	var err error
	val, err = convert.Convert(val, cty.String)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid version constraint",
			Detail:   fmt.Sprintf("A string value is required for %s.", attr.Name),
			Subject:  attr.Expr.Range().Ptr(),
		})
		return ret, diags
	}

	if val.IsNull() {
		// A null version constraint is strange, but we'll just treat it
		// like an empty constraint set.
		return ret, diags
	}

	if !val.IsWhollyKnown() {
		// If there is a syntax error, HCL sets the value of the given attribute
		// to cty.DynamicVal. A diagnostic for the syntax error will already
		// bubble up, so we will move forward gracefully here.
		return ret, diags
	}

	constraintStr := val.AsString()
	constraints, err := version.NewConstraint(constraintStr)
	if err != nil {
		// NewConstraint doesn't return user-friendly errors, so we'll just
		// ignore the provided error and produce our own generic one.
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid version constraint",
			Detail:   "This string does not use correct version constraint syntax.", // Not very actionable :(
			Subject:  attr.Expr.Range().Ptr(),
		})
		return ret, diags
	}

	ret.Required = constraints
	return ret, diags
}
