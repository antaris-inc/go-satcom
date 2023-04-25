# go-satcom

This repository contains golang libraries supporting satellite communication protocols:

* [csp](./csp) provides support for the [Cubesat Space Protocol (CSP)](https://github.com/libcsp/libcsp)
* [satlab](./satlab) provides support for [Satlab Spaceframes](https://www.satlab.com/resources/SLDS-SRS4-1.0.pdf)
* [openlst](./openlst) provides support for [OpenLST](https://github.com/OpenLST/openlst)

Additionally, the `Socket` and `Adapter` abstractions here help work with full communications channels.
Take a look at the examples in `socket_test.go` and the `test/` directory.

Feel free to open a Github issue with feedback/questions, or open a PR!

This project is licensed under APACHE 2.0 by Antaris, Inc.
