package main

import (
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
)

func Test_loadKey(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   types.AccAddress
		wantErr bool
		errMsg string
	}{
		{
			name: "valid key file",
			args: args{
				file: "testdata/private_key_file",
			},
			want:   "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
			want1:  types.MustAccAddressFromHex("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"),
		},
		{
			name: "invalid key file",
			args: args{
				file: "testdata/invalid_private_key_file",
			},
			wantErr: true,
			errMsg: "key file too short, want 42 hex characters",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := loadKey(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err!=nil && (err.Error() != tt.errMsg) {
				t.Errorf("loadKey() errorMsg = %v, wantErrMsg %v", err.Error(), tt.errMsg)
				return
			}
			if got != tt.want {
				t.Errorf("loadKey() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("loadKey() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
