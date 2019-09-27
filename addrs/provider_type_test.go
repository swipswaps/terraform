package addrs

// moved to regsrc

// import (
// 	"testing"

// 	"github.com/go-test/deep"
// )

// func TestParseProviderSourceStr(t *testing.T) {
// 	tests := map[string]struct {
// 		Want      ProviderType
// 		expectErr bool
// 	}{
// 		"registry.terraform.io/hashicorp/aws": {
// 			ProviderType{
// 				Name:      "aws",
// 				Namespace: "hashicorp",
// 				SourceURL: "registry.terraform.io",
// 			},
// 			false,
// 		},
// 		"hashicorp/aws": {
// 			ProviderType{
// 				Name:      "aws",
// 				Namespace: "hashicorp",
// 			},
// 			false,
// 		},
// 		"aws": {
// 			ProviderType{
// 				Name: "aws",
// 			},
// 			false,
// 		},
// 		"example.com/terraform/registry/hashicorp/aws": {
// 			ProviderType{
// 				Name:      "aws",
// 				Namespace: "hashicorp",
// 				SourceURL: "example.com/terraform/registry",
// 			},
// 			false,
// 		},
// 	}

// 	for name, test := range tests {
// 		got := ParseProviderSourceStr(name)
// 		for _, problem := range deep.Equal(got, test.Want) {
// 			t.Errorf(problem)
// 		}
// 	}
// }
