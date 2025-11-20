package observability

import obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

// FilterMap returns a map of filter names to FilterSpec.
func FilterMap(filters Filters) map[string]*obs.FilterSpec {
	m := map[string]*obs.FilterSpec{}
	for i := range filters {
		m[filters[i].Name] = &filters[i]
	}
	return m
}

type Filters []obs.FilterSpec

// Names returns a slice of filter names
func (f Filters) Names() (names []string) {
	for _, f := range f {
		names = append(names, f.Name)
	}
	return names
}

func (f Filters) Map() map[string]*obs.FilterSpec {
	return FilterMap(f)
}
