package deploy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/dexidp/dex/server"
	dexstorage "github.com/dexidp/dex/storage"
	"github.com/pkg/errors"
	kotsv1beta1 "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
	dextypes "github.com/replicatedhq/kots/pkg/identity/types/dex"
	"github.com/replicatedhq/kots/pkg/ingress"
	yaml "gopkg.in/yaml.v2"
)

func getDexConfig(ctx context.Context, identitySpec kotsv1beta1.IdentitySpec, identityConfigSpec kotsv1beta1.IdentityConfigSpec) ([]byte, error) {
	config := dextypes.Config{
		Issuer: dexIssuerURL(identityConfigSpec),
		Storage: dextypes.Storage{
			Type: "postgres",
			Config: dextypes.Postgres{
				SSL: dextypes.SSL{
					Mode: "disable", // TODO ssl
				},
			},
		},
		Web: dextypes.Web{
			HTTP: "0.0.0.0:5556",
		},
		Frontend: server.WebConfig{
			Issuer: "KOTS",
		},
		OAuth2: dextypes.OAuth2{
			SkipApprovalScreen:    true,
			AlwaysShowLoginScreen: identitySpec.OAUTH2AlwaysShowLoginScreen,
		},
		Expiry: dextypes.Expiry{
			IDTokens:    identitySpec.IDTokensExpiration,
			SigningKeys: identitySpec.SigningKeysExpiration,
		},
		StaticClients: []dexstorage.Client{
			{
				ID:           "kotsadm",
				Name:         "kotsadm",
				SecretEnv:    "DEX_CLIENT_SECRET",
				RedirectURIs: identitySpec.OIDCRedirectURIs,
			},
		},
		EnablePasswordDB: false,
	}

	connectors := []kotsv1beta1.DexConnector{}
	for _, connector := range identityConfigSpec.DexConnectors.Value {
		if len(identitySpec.SupportedProviders) == 0 || stringInSlice(connector.Type, identitySpec.SupportedProviders) {
			connectors = append(connectors, connector)
		}
	}

	if len(connectors) == 0 {
		return nil, errors.New("at lease one dex connector is required")
	}

	if len(connectors) > 0 {
		dexConnectors, err := DexConnectorsToDexTypeConnectors(connectors)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal dex connectors")
		}
		config.StaticConnectors = dexConnectors
	}

	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate dex config")
	}

	marshalledConfig, err := yaml.Marshal(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal dex config")
	}

	buf := bytes.NewBuffer(nil)
	t, err := template.New("dex-config").Funcs(template.FuncMap{
		"OIDCIdentityCallbackURL": func() string { return dexCallbackURL(identityConfigSpec) },
	}).Parse(string(marshalledConfig))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse dex config for templating")
	}
	if err := t.Execute(buf, nil); err != nil {
		return nil, errors.Wrap(err, "failed to execute template")
	}

	return buf.Bytes(), nil
}

func DexConnectorsToDexTypeConnectors(conns []kotsv1beta1.DexConnector) ([]dextypes.Connector, error) {
	dexConnectors := []dextypes.Connector{}
	for _, conn := range conns {
		f, ok := server.ConnectorsConfig[conn.Type]
		if !ok {
			return nil, errors.Errorf("unknown connector type %q", conn.Type)
		}

		connConfig := f()
		if len(conn.Config.Raw) != 0 {
			if err := json.Unmarshal(conn.Config.Raw, connConfig); err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal connector config")
			}
		}

		dexConnectors = append(dexConnectors, dextypes.Connector{
			Type:   conn.Type,
			Name:   conn.Name,
			ID:     conn.ID,
			Config: connConfig,
		})
	}
	return dexConnectors, nil
}

func dexIssuerURL(identityConfigSpec kotsv1beta1.IdentityConfigSpec) string {
	if identityConfigSpec.IdentityServiceAddress != "" {
		return identityConfigSpec.IdentityServiceAddress
	}
	return fmt.Sprintf("%s/dex", ingress.GetAddress(identityConfigSpec.IngressConfig))
}

func dexCallbackURL(identityConfigSpec kotsv1beta1.IdentityConfigSpec) string {
	return fmt.Sprintf("%s/callback", dexIssuerURL(identityConfigSpec))
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
