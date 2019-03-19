# Useage

This is a trace tool for brocaar loraserver.

## create lora-trace.toml


```
name = "hogehoge"
network-server = "localhost:8000"
mqtt-server = "ssl://mqtt.hogehoge.com:8883"
mqtt-username = "username"
mqtt-password = "password"
topic-top = "dump"
```

## start

```
% cd <where you created tom config>
% lora-trace
```

