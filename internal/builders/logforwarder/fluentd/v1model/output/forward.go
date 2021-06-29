package output

type ForwardOutputBuilder struct{

}

func NewForwardOutputBuilder() *ForwardOutputBuilder {
	return &ForwardOutputBuilder{}
}

func (b *ForwardOutputBuilder) AsList() []string {
	return []string{}
}