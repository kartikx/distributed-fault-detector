# Monitoring the bandwith

Install the pv package using Yum. Monitor the packets with `tcpdump` and count the bytes with `pv`.

```
sudo yum install pv
sudo tcpdump -ni ens33 udp -w- |pv -i2 >/dev/null
```
