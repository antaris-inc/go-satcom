#!/bin/bash -e

NAME=go-satcom_e2e-loopback
docker kill $NAME || true

DOWNLINK_PORT=8888
UPLINK_PORT=8889

docker run -d -t --rm -p 8888:8888 -p 8889:8889 --name=$NAME --entrypoint /bin/sh subfuzion/netcat -c "nc -kvl $UPLINK_PORT | nc -kvl $DOWNLINK_PORT"

export TEST_E2E_DOWNLINK_ADDRESS=127.0.0.1:$DOWNLINK_PORT
export TEST_E2E_UPLINK_ADDRESS=127.0.0.1:$UPLINK_PORT
go test -v ./test/...

docker kill $NAME
