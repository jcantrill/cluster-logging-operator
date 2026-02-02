package operators

type OperatorType string

const (
	OperatorTypeMove        OperatorType = "move"
	OperatorTypeRegexParser OperatorType = "regex_parser"
)

type Operator interface {
	GetType() OperatorType
	GetTypeName() string
}
