package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
)

// Validate returns an error if the new datasample is not valid:
// missing required data, incompatible values, etc.
func (o *NewDataSample) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Key, validation.Required, is.UUID),
		validation.Field(&o.DataManagerKeys, validation.Each(is.UUID)),
		validation.Field(&o.TestOnly, validation.NotNil),
		validation.Field(&o.Checksum, validation.Required, validation.Length(64, 64), is.Hexadecimal),
	)
}

func (p *RegisterDataSamplesParam) Validate() error {
	return validation.ValidateStruct(p, validation.Field(&p.Samples, validation.Required))
}

// Validate returns an error if the updated datasample is not valid:
// missing required data, incompatible values, etc.
func (o *UpdateDataSamplesParam) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Keys, validation.Required, validation.Each(is.UUID)),
		validation.Field(&o.DataManagerKeys, validation.Each(is.UUID)),
	)
}
