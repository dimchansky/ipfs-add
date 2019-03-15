package ipfs

import (
	"encoding/json"
	"testing"
)

func TestCid_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		jsonCid string
		cid     Cid
		wantErr bool
	}{
		{"normal json", `{"/":"zb2rhhFAEMepUBbGyP1k8tGfz7BSciKXP6GHuUeUsJBaK6cqG"}`, Cid("zb2rhhFAEMepUBbGyP1k8tGfz7BSciKXP6GHuUeUsJBaK6cqG"), false},
		{"error when empty string", `{"/":""}`, Cid(""), true},
		{"error when empty struct", `{}`, Cid(""), true},
		{"undefined when null", `null`, Cid(""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c Cid
			if err := json.Unmarshal([]byte(tt.jsonCid), &c); (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCid_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		cid     Cid
		want    string
		wantErr bool
	}{
		{"undefined cid", Cid(""), "null", false},
		{"normal cid", Cid("zb2rhhFAEMepUBbGyP1k8tGfz7BSciKXP6GHuUeUsJBaK6cqG"), `{"/":"zb2rhhFAEMepUBbGyP1k8tGfz7BSciKXP6GHuUeUsJBaK6cqG"}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.cid)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Marshal error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("Cid.MarshalJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
