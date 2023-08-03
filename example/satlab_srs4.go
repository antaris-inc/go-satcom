package example

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/antaris-inc/go-satcom"
	"github.com/antaris-inc/go-satcom/crc"
	"github.com/antaris-inc/go-satcom/satlab"
)

// Builds an example FrameConfig that may be used to exchange
// CSP messages with a Satlab SRS4 tranceiver.
func MakeSatlabSRS4FrameConfig() (satcom.FrameConfig, error) {
	// Set up the base config with proper ASM and frame size
	cfg := satcom.FrameConfig{
		FrameSyncMarker: satlab.SATLAB_ASM,
		FrameSize:       223,
	}

	// First adapter is the basic Satlab SpaceFrame, which prepends
	// a header and pads out to an expected length.
	cfg.Adapters = append(
		cfg.Adapters,
		&satlab.SpaceframeAdapter{
			satlab.SpaceframeConfig{
				Type:            satlab.SPACEFRAME_TYPE_CSP,
				PayloadDataSize: 217,
			},
		},
	)

	// Assuming CRC is enabled on the transceiver, this additional
	// adapter will add/remove/verify CRC checksums.
	crc32Adapter, err := crc.NewCRC32Adapter(crc.CRC32AdapterConfig{
		Algorithm: crc.CRC32c,
	})
	if err != nil {
		return cfg, err
	}
	cfg.Adapters = append(
		cfg.Adapters,
		crc32Adapter,
	)

	return cfg, nil
}

// Use the example Satlab SRS4 FrameConfig to generate a idleframe using
// an empty payload. Any additional features will be applied correctly
// but the SRS4 tranceiver will discard it upon receipt due to a lack
// of payload data. This does NOT include the ASM
func NewSatlabSRS4IdleFrame_Empty() ([]byte, error) {
	cfg, err := MakeSatlabSRS4FrameConfig()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	fs, err := satcom.NewFrameSender(cfg, buf)
	if err != nil {
		return nil, err
	}

	if err := fs.Send([]byte{}); err != nil {
		return nil, err
	}

	return buf.Bytes()[len(cfg.FrameSyncMarker):], nil
}

// Generates an idle frame using randomized data. This method results in an
// invalid CRC checksum, which will then cause the SRS4 trancevier to discard
// the frame altogether. This approach is less commonly used than the "empty"
// approach documented above. This does NOT include the ASM.
func NewSatlabSRS4IdleFrame_Rand() ([]byte, error) {
	// Just need this for FrameSize value
	cfg, _ := MakeSatlabSRS4FrameConfig()

	frm := make([]byte, cfg.FrameSize)
	n, err := rand.Read(frm)
	if err != nil {
		return nil, fmt.Errorf("rand read failed: %v", err)
	} else if n != cfg.FrameSize {
		return nil, errors.New("read incorrect number of random bytes")
	}

	return frm, nil
}
