package filter

import (
	"fmt"

	fapi "github.com/openshift/cluster-logging-operator/internal/builders/config/fluentd"
)

type RecordTransformerFilter struct {
	fapi.Filter
	Record Record
}

type Record map[string]string

func (r Record) AsList() []string {
	buf := []string{"<record>"}
	for k, v := range r {
		buf = append(buf, fmt.Sprintf("%s %v", k, v))
	}
	return append(buf, "</record>")
}
func (r Record) Set(key string, value interface{}) {
	r[key] = value.(string)
}

func NewRecordTransformerFilterBuilder(tag string) *RecordTransformerFilter {
	record := Record{}
	f := &RecordTransformerFilter{
		Filter: fapi.Filter{
			Tag: tag,
			Configuration: fapi.Configuration{
				Type: "record_transformer",
				AllowedKeys: fapi.NewSet(
					"enable_ruby",
					"record",
				),
				Config: map[string]interface{}{
					"record": record,
				},
			},
		},
		Record: record,
	}
	return f
}

func (f *RecordTransformerFilter) EnableRuby(enable bool) *RecordTransformerFilter {
	f.Set("enable_ruby", enable)
	return f
}
func (f *RecordTransformerFilter) AddToRecord(key, value string) *RecordTransformerFilter {
	f.Record.Set(key, value)
	return f
}
