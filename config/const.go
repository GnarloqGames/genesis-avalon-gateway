package config

const (
	EnvPrefix string = "AVALOND"

	EnvEnvironment       string = "ENVIRONMENT"
	EnvLogLevel          string = "LOG_LEVEL"
	EnvLogKind           string = "LOG_KIND"
	EnvGatewayHost       string = "GATEWAY_HOST"
	EnvGatewayPort       string = "GATEWAY_PORT"
	EnvNatsAddress       string = "NATS_ADDRESS"
	EnvNatsEncoding      string = "NATS_ENCODING"
	EnvCouchbaseURL      string = "COUCHBASE_URL"
	EnvCouchbaseBucket   string = "COUCHBASE_BUCKET"
	EnvCouchbaseUsername string = "COUCHBASE_USERNAME"
	EnvCouchbasePassword string = "COUCHBASE_PASSWORD"
	EnvOidcProvider      string = "OIDC_PROVIDER"
	EnvOidcClientId      string = "OIDC_CLIENT_ID"

	FlagEnvironment       string = "environment"
	FlagLogLevel          string = "log-level"
	FlagLogKind           string = "log-kind"
	FlagGatewayHost       string = "host"
	FlagGatewayPort       string = "port"
	FlagNatsAddress       string = "nats-address"
	FlagNatsEncoding      string = "nats-encoding"
	FlagCouchbaseURL      string = "couchbase-url"
	FlagCouchbaseBucket   string = "couchbase-bucket"
	FlagCouchbaseUsername string = "couchbase-username"
	FlagCouchbasePassword string = "couchbase-password"
	FlagOidcProvider      string = "oidc-provider"
	FlagOidcClientId      string = "oidc-client-id"
)
