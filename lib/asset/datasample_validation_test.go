package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type datasampleTestCase struct {
	datasample *NewDataSample
	valid      bool
}

type updateDataSampleTestCase struct {
	datasample *UpdateDataSamplesParam
	valid      bool
}

func TestNewDataSampleValidate(t *testing.T) {
	cases := map[string]datasampleTestCase{
		"empty": {&NewDataSample{}, false},
		"invalidDataSampleKey": {&NewDataSample{
			Key:             "not36chars",
			DataManagerKeys: []string{"834f47c3-2d95-4ccd-a718-7143b64e61c0", "c6cc913d-83a9-4a8e-a258-2901e1d5ebbc"},
			TestOnly:        false,
			Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		}, false},
		"invalidDataManagerKey": {&NewDataSample{
			Key:             "834f47c3-2d95-4ccd-a718-7143b64e61c0",
			DataManagerKeys: []string{"not36chars", "1d417d76-a2e1-46e7-aae5-9c7c165575fc"},
			TestOnly:        false,
			Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		}, false},
		"valid": {&NewDataSample{
			Key:             "834f47c3-2d95-4ccd-a718-7143b64e61c0",
			DataManagerKeys: []string{"3dd165f8-8822-481a-8bf9-23bf135152cf", "1d417d76-a2e1-46e7-aae5-9c7c165575fc"},
			TestOnly:        false,
			Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		}, true},
		"validTestOnly": {&NewDataSample{
			Key:             "834f47c3-2d95-4ccd-a718-7143b64e61c0",
			DataManagerKeys: []string{"3dd165f8-8822-481a-8bf9-23bf135152cf", "1d417d76-a2e1-46e7-aae5-9c7c165575fc"},
			TestOnly:        true,
			Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		}, true},
		"invalidChecksum": {&NewDataSample{
			Key:             "834f47c3-2d95-4ccd-a718-7143b64e61c0",
			DataManagerKeys: []string{"3dd165f8-8822-481a-8bf9-23bf135152cf", "1d417d76-a2e1-46e7-aae5-9c7c165575fc"},
			TestOnly:        false,
			Checksum:        "j2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		}, false},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.datasample.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.datasample.Validate(), name+" should be invalid")
		}
	}
}

func TestUpdateDataSampleValidate(t *testing.T) {
	cases := map[string]updateDataSampleTestCase{
		"empty": {&UpdateDataSamplesParam{}, false},
		"invalidDataSampleKey": {&UpdateDataSamplesParam{
			Keys:            []string{"not36chars", "08680966-97ae-4573-8b2d-6c4db2b3cdd2"},
			DataManagerKeys: []string{"834f47c3-2d95-4ccd-a718-7143b64e61c0", "08680966-97ae-4573-8b2d-6c4db2b3cdd2"},
		}, false},
		"invalidDataManagerKey": {&UpdateDataSamplesParam{
			Keys:            []string{"08680966-97ae-4573-8b2d-6c4db2b3cdd2", "3dd165f8-8822-481a-8bf9-23bf135152cf"},
			DataManagerKeys: []string{"not36chars", "08680966-97ae-4573-8b2d-6c4db2b3cdd2"},
		}, false},
		"valid": {&UpdateDataSamplesParam{
			Keys:            []string{"834f47c3-2d95-4ccd-a718-7143b64e61c0", "c6cc913d-83a9-4a8e-a258-2901e1d5ebbc"},
			DataManagerKeys: []string{"3dd165f8-8822-481a-8bf9-23bf135152cf", "1d417d76-a2e1-46e7-aae5-9c7c165575fc"},
		}, true},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.datasample.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.datasample.Validate(), name+" should be invalid")
		}
	}
}
