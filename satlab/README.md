# spaceframe

This package provides support for the Satlab Spaceframe protocol, which is used to communicate with the Satlab SRS4 S-band Transceiver.
The SRS4 datasheet describes the Spaceframe format: https://www.satlab.com/resources/SLDS-SRS4-1.0.pdf. 

## Quickstart

After importing this package, build a config and enframe your message:

```
	cfg := satlab.SpaceframeConfig{
		Type:            satlab.SPACEFRAME_TYPE_CSP,
		PayloadDataSize: 1024,
		CRCEnabled:      true,
	}

	msg := []byte("ping")
	frm, err := satlab.Enframe(msg, &cfg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Message = % x\n", msg)
	fmt.Printf("Encoded Spaceframe = % x\n", frm)
```

As part of the enframing process, the message is padded to the indicated `PayloadDataSize`.
This padding must be maintained, as this is a frame-oriented protocol.
As a convenience, you can use `cfg.FrameSize()` to inform how large your read buffer should be, as demonstrated below:

```
	// This io.Reader simulates an arbitrary data source, but we fill
	// it with the contents of frm for the sake of a working example.
	src := bytes.NewBuffer(frm)

	rb := make([]byte, cfg.FrameSize())
	if _, err = src.Read(rb); err != nil {
		panic(err)
	}

	msg, err = satlab.Deframe(rb, &cfg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Decoded message = % x\n", msg)
```
