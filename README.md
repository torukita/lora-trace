# Useage

This is a trace tool for brocaar loraserver.

## create lora-trace.toml


```
name = "hogehoge"
network-server = "localhost:8000"
mqtt-server = "ssl://mqtt.hogehoge.com:8883"
mqtt-username = "username"
mqtt-password = "password"
mqtt-trace-topic = "dump"
```

## start

```
% cd <where you created tom config>
% lora-trace
```

## add/delete deveui

```
% lora-trace add <deveui>
or
% lora-trace delete <deveui>
```

## mqtt sample

```
$ mosquitto_sub -h mqtt.hogehoge.com -u username -P password -t dump/frame/<deveui> | jq .
{
  "uplink": {
    "tx_info": {
      "frequency": 927600000,
      "ModulationInfo": {
        "LoraModulationInfo": {
          "bandwidth": 125,
          "spreading_factor": 8,
          "code_rate": "4/5"
        }
      }
    },
    "rx_info": [
      {
        "gateway_id": "00001c497bcaafa1",
        "time": {
          "seconds": 1552985781,
          "nanos": 942538000
        },
        "time_since_gps_epoch": {
          "seconds": 1237020999,
          "nanos": 942000000
        },
        "timestamp": 3400064060,
        "rssi": -52,
        "lora_snr": 11,
        "channel": 5,
        "rf_chain": 1,
        "location": {},
        "FineTimestamp": null
      }
    ],
    "phy_payload": {
      "mhdr": {
        "mType": "UnconfirmedDataUp",
        "major": "LoRaWANR1"
      },
      "macPayload": {
        "fhdr": {
          "devAddr": "008805ea",
          "fCtrl": {
            "adr": false,
            "adrAckReq": false,
            "ack": false,
            "fPending": false,
            "classB": false
          },
          "fCnt": 6,
          "fOpts": null
        },
        "fPort": 5,
        "frmPayload": [
          {
            "bytes": "Gw=="
          }
        ]
      },
      "mic": "56c276e6"
    }
  }
}
```
