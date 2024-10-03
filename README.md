# Distributed Fault Detector
This repository implements a distributed fault detection service on a cluster of VMs. Any failure on a VM is detected within 10 seconds.

We implement the following two modes of failure detection:
1. **PingAck**: A naive implementation where all nodes periodically ping each other. Any failed acknowledgement results in a failure prediction.
2. **SWIM**: An elaborate implementation of the [SWIM](https://en.wikipedia.org/wiki/SWIM_Protocol) protocol. A failed acknowledgement disseminates suspicion among the group. A persistent suspicion is eventually converted into a failure detection.
