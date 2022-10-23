package mscfb

type Validation int

const (
	ValidationPermissive Validation = iota
	ValidationStrict     Validation = iota
)

func (v Validation) IsStrict() bool {
	return v == ValidationStrict
}
