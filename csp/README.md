# csp

This package provides support for encoding and decoding Cubest Space Protocol packets.
The upstream project located at https://github.com/libcsp/libcsp is authoritative in the wire formats implemented here.
Note that libcsp goes much further to implement a full messaging system, while this repo simply aims to support client compatibility.

This repo contains support for CSP v1 and v2, the primary difference coming in the expected header length due to larger fields.
It is common for a client to solely work with either v1 or v2, so it is encouraged that clients of this code assign a `csp` package name in the import statement:

```
import csp "github.com/antaris-inc/go-satcom/csp/v2"
```

## Quickstart

After importing the required version of the CSP lib, simply construct a header and encode a message:

```
	outgoing := csp.Packet{
		PacketHeader: csp.PacketHeader{
			Priority:        1,
			Destination:     12,
			DestinationPort: 13,
			Source:          14,
			SourcePort:      15,
		},
		Data: []byte("ping"),
	}

	pkt := outgoing.ToBytes() // returns a byte slice containing your encoded CSP packet
	fmt.Printf("Encoded Packet = % x\n", pkt)
```

The PacketHeader fields are used by the server to provide routing and QoS once your send the encoded packet to the appropriate endpoint.
This library does nothing with these fields other then provide encoding and decoding support.

Decoding a byte slice as a Packet is also straightforward:

```
	var incoming csp.Packet
	if err := incoming.FromBytes(pkt); err != nil {
		panic(err)
	}

	// incoming now contains your decoded Packet
	fmt.Printf("Decoded Packet = %+v\n", incoming)
```

The test files contained in this repo contain additional examples.
