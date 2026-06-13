package live

import (
	"encoding/base64"
	"fmt"
)

// PCM audio constants — fixed by the spec.
const (
	SampleRateIn = 16000
	SampleRateOut = 24000
	Channels = 1
	Encoding = "pcm_s16le"
	BytesPerSample = 2 // Int16
)

// PCMDurationMs returns the playback duration of a raw PCM byte buffer
// at the given sample rate, in milliseconds. Buffer length must be a
// multiple of 2 (one sample = 2 bytes). Returns 0 for empty input.
func PCMDurationMs(pcmLen int, sampleRate int) int {
	if pcmLen == 0 || sampleRate == 0 {
		return 0
	}
	samples := pcmLen / BytesPerSample
	return (samples * 1000) / sampleRate
}

// EncodePCMBase64 returns the base64 (standard, padded) encoding of raw PCM bytes.
func EncodePCMBase64(pcm []byte) string {
	return base64.StdEncoding.EncodeToString(pcm)
}

// DecodePCMBase64 decodes a base64 PCM payload and validates that the
// resulting length is a multiple of 2 (whole Int16 samples).
func DecodePCMBase64(s string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}
	if len(raw)%BytesPerSample != 0 {
		return nil, fmt.Errorf("pcm length %d not a multiple of %d", len(raw), BytesPerSample)
	}
	return raw, nil
}
