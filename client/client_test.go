package client_test

import (
	"context"
	"testing"

	"github.com/arpsch/xm/client"
	"github.com/pkg/errors"
)

func TestIPAPI_GetCountryNameByIP(t *testing.T) {

	tt := []struct {
		name   string
		input  string
		want   string
		expErr error
	}{
		{
			name:   "validate FR (France) country IP",
			input:  "20.188.40.63",
			want:   "FR",
			expErr: nil,
		},
		{
			name:   "invalid IP",
			input:  "20.188.40.63.0",
			want:   "",
			expErr: errors.New("Invalid IP"),
		},
	}
	ctx := context.Background()

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {
			country, err := client.IPAPI_GetCountryNameByIP(ctx, tc.input)
			if err != nil && tc.expErr != nil && tc.expErr.Error() != err.Error() {
				t.Errorf("expected error %v, actual error %v", tc.expErr, err)
			}

			if err == nil && country != tc.want {
				t.Errorf("expected country %v, received country %v", tc.want, country)
			}
		})
	}
}
