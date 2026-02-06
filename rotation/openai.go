package rotation

import (
	"context"
	"fmt"
	"time"
)

type openaiPlugin struct{}

func init() { GetGlobalRegistry().Register(&openaiPlugin{}) }

func (p *openaiPlugin) Name() string                    { return "openai" }
func (p *openaiPlugin) RotatableFields() []RotatableField { return []RotatableField{FieldSecretKey} }

func (p *openaiPlugin) Validate(cred CredentialInfo) error {
	if cred.APIType != "openai" {
		return fmt.Errorf("expected api_type openai, got %q", cred.APIType)
	}
	if cred.SecretKey == nil || *cred.SecretKey == "" {
		return fmt.Errorf("openai credential requires a secret key")
	}
	return nil
}

func (p *openaiPlugin) ConfigSchema() ConfigSchema {
	return ConfigSchema{Fields: []ConfigField{
		{Name: "organization_id", Description: "OpenAI organization ID", Required: true},
		{Name: "admin_key", Description: "Admin API key for key management", Required: true, Secret: true},
	}}
}

func (p *openaiPlugin) Rotate(_ context.Context, cred CredentialInfo, _ Config) (*Result, error) {
	// Stub: real implementation would call OpenAI admin API
	newKey := "sk-rotated-stub-" + cred.Name
	return &Result{
		NewSecretKey: &newKey,
		KeyID:        "key-" + cred.Name,
		OldKeyGrace:  5 * time.Minute,
		Metadata:     map[string]string{"stub": "true"},
	}, nil
}
