package rotation

import (
	"context"
	"fmt"
	"time"
)

type supabasePlugin struct{}

func init() { GetGlobalRegistry().Register(&supabasePlugin{}) }

func (p *supabasePlugin) Name() string { return "supabase" }
func (p *supabasePlugin) RotatableFields() []RotatableField {
	return []RotatableField{FieldSecretKey, FieldPublicKey}
}

func (p *supabasePlugin) Validate(cred CredentialInfo) error {
	if cred.APIType != "supabase" {
		return fmt.Errorf("expected api_type supabase, got %q", cred.APIType)
	}
	if cred.URL == nil || *cred.URL == "" {
		return fmt.Errorf("supabase credential requires a URL")
	}
	if (cred.SecretKey == nil || *cred.SecretKey == "") && (cred.PublicKey == nil || *cred.PublicKey == "") {
		return fmt.Errorf("supabase credential requires at least one key")
	}
	return nil
}

func (p *supabasePlugin) ConfigSchema() ConfigSchema {
	return ConfigSchema{Fields: []ConfigField{
		{Name: "project_ref", Description: "Supabase project reference", Required: true},
		{Name: "access_token", Description: "Supabase management API token", Required: true, Secret: true},
		{Name: "rotate_service_role", Description: "Also rotate the service role key", Required: false},
	}}
}

func (p *supabasePlugin) Rotate(_ context.Context, cred CredentialInfo, _ Config) (*Result, error) {
	// Stub: real implementation would call Supabase management API
	newSecret := "sbp_rotated-stub-" + cred.Name
	newPublic := "eyJ-rotated-stub-" + cred.Name
	return &Result{
		NewSecretKey: &newSecret,
		NewPublicKey: &newPublic,
		KeyID:        "supa-" + cred.Name,
		OldKeyGrace:  2 * time.Minute,
		Metadata:     map[string]string{"stub": "true"},
	}, nil
}
