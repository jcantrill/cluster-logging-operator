package operators

type Move struct {
	Id   string       `json:"id,omitempty"`
	Type OperatorType `json:"type"`
	From string       `json:"string"`
	To   string       `json:"to"`
}

func NewMove(from, to string) Move {
	return Move{
		Type: OperatorTypeMove,
		From: from,
		To:   to,
	}
}

func (r Move) GetType() OperatorType {
	return r.Type
}

func (r Move) GetTypeName() string {
	return r.Id
}
