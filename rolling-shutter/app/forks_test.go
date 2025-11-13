package app

import "testing"

func TestIsForkActive(t *testing.T) {
	testcases := []struct {
		name               string
		forkHeightGenesis  *int64
		override           *ForkHeightOverride
		currentBlockHeight int64
		currentEon         uint64
		want               bool
	}{
		{
			name:               "override height met",
			override:           &ForkHeightOverride{Height: int64Ptr(10)},
			currentBlockHeight: 10,
			want:               true,
		},
		{
			name:               "override height not met",
			override:           &ForkHeightOverride{Height: int64Ptr(10)},
			currentBlockHeight: 9,
			want:               false,
		},
		{
			name:       "override eon met",
			override:   &ForkHeightOverride{Eon: uint64Ptr(5)},
			currentEon: 5,
			want:       true,
		},
		{
			name:       "override eon not met",
			override:   &ForkHeightOverride{Eon: uint64Ptr(5)},
			currentEon: 4,
			want:       false,
		},
		{
			name:               "override height takes precedence over eon",
			override:           &ForkHeightOverride{Height: int64Ptr(10), Eon: uint64Ptr(2)},
			currentBlockHeight: 9,
			currentEon:         5,
			want:               false,
		},
		{
			name:               "override without fields never activates fork",
			override:           &ForkHeightOverride{},
			currentBlockHeight: 100,
			currentEon:         100,
			want:               false,
		},
		{
			name:               "genesis height met without override",
			forkHeightGenesis:  int64Ptr(7),
			currentBlockHeight: 7,
			want:               true,
		},
		{
			name:               "genesis height not met without override",
			forkHeightGenesis:  int64Ptr(7),
			currentBlockHeight: 6,
			want:               false,
		},
		{
			name:               "no override and no fork height",
			currentBlockHeight: 100,
			want:               false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isForkActive(tc.forkHeightGenesis, tc.override, tc.currentBlockHeight, tc.currentEon)
			if got != tc.want {
				t.Fatalf("isForkActive() = %t, want %t (test case: %+v)", got, tc.want, tc)
			}
		})
	}
}
