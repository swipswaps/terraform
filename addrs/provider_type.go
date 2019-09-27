package addrs

// ProviderType encapsulates a single provider type. In the future this will be
// extended to include additional fields including Namespace and SourceURL
type ProviderType struct {
	Name   string
	Source string
	// Version
}
