package viaq

const (
	SetClusterID         = `._internal.openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"`
	SetTimestampField    = `ts = del(._internal.timestamp); if !exists(._internal."@timestamp") {._internal."@timestamp" = ts}`
	SetMessageOnRoot     = `.message = del(._internal.message)`
	SetOpenShiftOnRoot   = `if exists(._internal.openshift) {.openshift = ._internal.openshift}`
	SetOpenShiftSequence = `._internal.openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")`
)
