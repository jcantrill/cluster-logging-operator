package operators

type RegexParser struct {
	Id    string       `json:"id,omitempty"`
	Type  OperatorType `json:"type"`
	Regex string       `json:"regex"`

	//Output is the id of the operator to pass the result directly
	Output string `json:"output,omitempty"`

	ParseFrom string     `json:"parse_from,omitempty"`
	Timestamp *Timestamp `json:"timestamp,omitempty"`
	Cache     *Cache     `json:"cache,omitempty"`
}

func NewRegexParser(id, pattern string, init func(parser *RegexParser)) *RegexParser {
	p := &RegexParser{
		Id:    id,
		Type:  OperatorTypeRegexParser,
		Regex: pattern,
	}
	if init != nil {
		init(p)
	}
	return p
}

func (r RegexParser) GetType() OperatorType {
	return r.Type
}

func (r RegexParser) GetTypeName() string {
	return r.Id
}

type Cache struct {
	Size uint `json:"size,omitempty"`
}

type TimestampLayoutType string

const (
	TimestampLayoutTypeGoTime TimestampLayoutType = "gotime"
)

type Timestamp struct {
	ParseFrom  string              `json:"parse_from"`
	LayoutType TimestampLayoutType `json:"layout_type"`
	Layout     string              `json:"layout"`
}
