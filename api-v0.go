package hostdb

// APIv0Config contains the configuration for HostDB API v0
type APIv0Config struct {

	// this should be map[HostdbDoctype]field
	// describes the required context fields for a given type
	ContextFields map[string][]string `mapstructure:"context_fields"`

	// the default number of records to return on each page
	DefaultLimit int `mapstructure:"default_limit"`

	// fields to show when calling list endpoint
	ListFields []string `mapstructure:"list_fields"`

	// valid query parameters, grouped by type
	// query-param (e.g. flavor) =>
	//   record-type (e.g. openstack) =>
	//     place-to-find-data (e.g. context) =>
	//       data-location (e.g. .flavor)
	QueryParams map[string]map[string]APIv0QueryParam `mapstructure:"query_params"`
}

// APIv0QueryParam defines where a given data point can be found;
// * in the context JSON
// * in the data JSON
// * or in the database table itself
// and the Display Name (text which should be displayed in UIs)
type APIv0QueryParam struct {
	Context     string `mapstructure:"context"`
	Data        string `mapstructure:"data"`
	DisplayName string `mapstructure:"_name"`
	Table       string `mapstructure:"table"`
}
