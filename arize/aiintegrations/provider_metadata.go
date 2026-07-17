package aiintegrations

import (
	"encoding/json"
	"fmt"

	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
)

// ProviderMetadata is the typed oneOf wrapper for provider-specific
// configuration on CreateRequest.ProviderMetadata and
// UpdateRequest.ProviderMetadata. Set exactly one of AWS or GCP.
//
// On the wire this serializes as the bare provider-metadata object with
// kind="aws" or kind="gcp" (the discriminator is set automatically).
//
// On UpdateRequest, an empty wrapper (both AWS and GCP nil) marshals as JSON
// null — the OpenAPI "Pass null to remove" signal — to clear the existing
// provider metadata. On CreateRequest, that shape will be rejected by the
// server when the provider requires metadata (awsBedrock, vertexAI).
type ProviderMetadata struct {
	// AWS, when non-nil, configures AWS Bedrock provider metadata. Required
	// when CreateRequest.Provider is AIIntegrationProviderAWSBedrock.
	AWS *AWSProviderMetadata
	// GCP, when non-nil, configures Vertex AI (GCP) provider metadata.
	// Required when CreateRequest.Provider is AIIntegrationProviderVertexAI.
	GCP *GCPProviderMetadata
}

// MarshalJSON serializes the active variant, automatically setting the kind
// discriminator. Returns an error if both AWS and GCP are non-nil.
func (p ProviderMetadata) MarshalJSON() ([]byte, error) {
	switch {
	case p.AWS != nil && p.GCP != nil:
		return nil, fmt.Errorf("aiintegrations: ProviderMetadata.AWS and ProviderMetadata.GCP are mutually exclusive; set exactly one")
	case p.AWS != nil:
		m := *p.AWS
		m.Kind = generated.AwsProviderMetadataKindAWS
		return json.Marshal(m)
	case p.GCP != nil:
		m := *p.GCP
		m.Kind = generated.GcpProviderMetadataKindGCP
		return json.Marshal(m)
	default:
		return []byte("null"), nil
	}
}
