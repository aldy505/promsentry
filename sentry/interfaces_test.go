package sentry

import (
	"encoding/json"
	"flag"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	update   = flag.Bool("update", false, "update .golden files")
	generate = flag.Bool("gen", false, "generate missing .golden files")
)

func TestUserIsEmpty(t *testing.T) {
	tests := []struct {
		input User
		want  bool
	}{
		{input: User{}, want: true},
		{input: User{ID: "foo"}, want: false},
		{input: User{Email: "foo@example.com"}, want: false},
		{input: User{IPAddress: "127.0.0.1"}, want: false},
		{input: User{Username: "My Username"}, want: false},
		{input: User{Name: "My Name"}, want: false},
		{input: User{Segment: "My Segment"}, want: false},
		{input: User{Data: map[string]string{"foo": "bar"}}, want: false},
		{input: User{ID: "foo", Email: "foo@example.com", IPAddress: "127.0.0.1", Username: "My Username", Name: "My Name", Segment: "My Segment", Data: map[string]string{"foo": "bar"}}, want: false},
	}

	for _, test := range tests {
		assertEqual(t, test.input.IsEmpty(), test.want)
	}
}

func TestUserMarshalJson(t *testing.T) {
	tests := []struct {
		input User
		want  string
	}{
		{input: User{}, want: `{}`},
		{input: User{ID: "foo"}, want: `{"id":"foo"}`},
		{input: User{Email: "foo@example.com"}, want: `{"email":"foo@example.com"}`},
		{input: User{IPAddress: "127.0.0.1"}, want: `{"ip_address":"127.0.0.1"}`},
		{input: User{Username: "My Username"}, want: `{"username":"My Username"}`},
		{input: User{Name: "My Name"}, want: `{"name":"My Name"}`},
		{input: User{Segment: "My Segment"}, want: `{"segment":"My Segment"}`},
		{input: User{Data: map[string]string{"foo": "bar"}}, want: `{"data":{"foo":"bar"}}`},
	}

	for _, test := range tests {
		got, err := json.Marshal(test.input)
		if err != nil {
			t.Fatal(err)
		}

		assertEqual(t, string(got), test.want)
	}
}
func TestEventWithDebugMetaMarshalJSON(t *testing.T) {
	t.Skip()
	event := NewEvent()
	event.DebugMeta = &DebugMeta{
		SdkInfo: &DebugMetaSdkInfo{
			SdkName:           "test",
			VersionMajor:      1,
			VersionMinor:      2,
			VersionPatchlevel: 3,
		},
		Images: []DebugMetaImage{
			{
				Type:        "macho",
				ImageAddr:   "0xabcd0000",
				ImageSize:   32768,
				DebugID:     "42DB5B96-5144-4079-BE09-45E2142CA3E5",
				DebugFile:   "foo.dSYM",
				CodeID:      "A7AF6477-9130-4EB7-ADFE-AD0F57001DBD",
				CodeFile:    "foo.dylib",
				ImageVmaddr: "0x0",
				Arch:        "arm64",
			},
			{
				Type: "proguard",
				UUID: "982E62D4-6493-4E43-864B-6523C79C7064",
			},
		},
	}

	got, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	want := `{"sdk":{},"user":{},` +
		`"debug_meta":{` +
		`"sdk_info":{"sdk_name":"test","version_major":1,"version_minor":2,"version_patchlevel":3},` +
		`"images":[` +
		`{"type":"macho",` +
		`"image_addr":"0xabcd0000",` +
		`"image_size":32768,` +
		`"debug_id":"42DB5B96-5144-4079-BE09-45E2142CA3E5",` +
		`"debug_file":"foo.dSYM",` +
		`"code_id":"A7AF6477-9130-4EB7-ADFE-AD0F57001DBD",` +
		`"code_file":"foo.dylib",` +
		`"image_vmaddr":"0x0",` +
		`"arch":"arm64"` +
		`},` +
		`{"type":"proguard","uuid":"982E62D4-6493-4E43-864B-6523C79C7064"}` +
		`]}}`

	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("Event mismatch (-want +got):\n%s", diff)
	}
}

func TestMechanismMarshalJSON(t *testing.T) {
	mechanism := &Mechanism{
		Type:        "some type",
		Description: "some description",
		HelpLink:    "some help link",
		Data: map[string]interface{}{
			"some data":         "some value",
			"some numeric data": 12345,
		},
	}

	got, err := json.Marshal(mechanism)
	if err != nil {
		t.Fatal(err)
	}

	want := `{"type":"some type","description":"some description","help_link":"some help link",` +
		`"data":{"some data":"some value","some numeric data":12345}}`

	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("Event mismatch (-want +got):\n%s", diff)
	}
}

func TestMechanismMarshalJSON_withHandled(t *testing.T) {
	mechanism := &Mechanism{
		Type:        "some type",
		Description: "some description",
		HelpLink:    "some help link",
		Data: map[string]interface{}{
			"some data":         "some value",
			"some numeric data": 12345,
		},
	}
	mechanism.SetUnhandled()

	got, err := json.Marshal(mechanism)
	if err != nil {
		t.Fatal(err)
	}

	want := `{"type":"some type","description":"some description","help_link":"some help link",` +
		`"handled":false,"data":{"some data":"some value","some numeric data":12345}}`

	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("Event mismatch (-want +got):\n%s", diff)
	}
}
