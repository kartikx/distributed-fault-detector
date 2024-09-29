# Monitoring the bandwith

Install the nethogs package using Yum. Use the `-C` flag to monitor both TCP and UDP traffic.

```
sudo yum install nethogs
sudo nethogs -C
```

# Adding a bandwith limiting rule

Check [this](https://netbeez.net/blog/how-to-use-the-linux-traffic-control/) post for explanations about the parameters.

Command to limit the bandwidth to 1mbit/s:
```
sudo tc qdisc add dev ens33 root tbf rate 1mbit burst 32kbit latency 400ms
```

Command to remove all `tc` rules:
```
sudo tc qdisc del dev ens33 root
```

Command to see all `tc` rules:
```
sudo tc qdisc show dev ens33
```