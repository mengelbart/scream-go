package scream

import "testing"

func Test_ntpToQ16(t *testing.T) {
	type args struct {
		ntpTime uint64
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "1",
			args: args{ntpTime: 16438227384021503999},
			want: 1354127795,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ntpToQ16(tt.args.ntpTime); got != tt.want {
				t.Errorf("ntpToQ16() = %v, want %v", got, tt.want)
			}
		})
	}
}
