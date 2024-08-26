package viaq

const (
	SetMessageOnRoot     = `.message = del(._internal.message)`
	SetOpenShiftOnRoot   = `if exists(._internal.openshift) {.openshift = ._internal.openshift}`
	SetOpenShiftSequence = `._internal.openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")`
)
