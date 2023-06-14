package epoch

import "time"

type Duration struct {
	time.Duration
}

func (k *Duration) UnmarshalText(b []byte) error {
	dur, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	k.Duration = dur
	return nil
}

func (k *Duration) Equal(b *Duration) bool {
	aNs := k.Abs().Nanoseconds()
	bNs := b.Abs().Nanoseconds()
	return aNs == bNs
}

func (k *Duration) MarshalText() ([]byte, error) { //nolint: unparam
	return []byte(k.String()), nil
}
