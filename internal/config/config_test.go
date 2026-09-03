package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateRejectsDevSecretsInProduction(t *testing.T) {
	t.Setenv("GO_ENV", "production")
	c := &Config{
		HttpPort:            3053,
		BadgerDBPath:        "/tmp/db",
		JWTSecret:           devJWTSecret,
		PublicBaseURL:       "https://app.example",
		OnlyOfficeJWTSecret: "real-secret",
	}
	require.Error(t, c.Validate())

	c.JWTSecret = "real-jwt"
	c.OnlyOfficeJWTSecret = devOOJWTSecret
	require.Error(t, c.Validate())

	c.OnlyOfficeJWTSecret = "real-oo"
	require.NoError(t, c.Validate())
}

func TestValidateAllowsDevSecretsLocally(t *testing.T) {
	t.Setenv("GO_ENV", "development")
	c := &Config{
		HttpPort:            3053,
		BadgerDBPath:        "/tmp/db",
		JWTSecret:           devJWTSecret,
		PublicBaseURL:       "http://localhost",
		OnlyOfficeJWTSecret: devOOJWTSecret,
	}
	require.NoError(t, c.Validate())
}
