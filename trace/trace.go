package trace

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"io"
)

const (
	MAP_MAX = 5
)

type TraceObject struct {
	maps map[string]bool
	nsClient *NSClient
	mqtt *MqttClient
	dataHandler func(lorawan.EUI64, []byte) error
	onHandler func(lorawan.EUI64) error
	offHandler func(lorawan.EUI64) error	
}

func defaultDataHandler(deveui lorawan.EUI64, bytes []byte) error {
	log.WithField("deveui", deveui).Trace("called defautlDataHandler")
	return nil
}

func (t *TraceObject)sendFrameHandler() func(lorawan.EUI64, []byte) error {
	return func(deveui lorawan.EUI64, bytes []byte) error {
		log.WithField("deveui", deveui).Trace("called sendFrameHandler")
		return t.mqtt.TraceFrame(deveui, bytes)
	}
}

func (t *TraceObject)traceOnHandler() func(lorawan.EUI64) error {
	return func(deveui lorawan.EUI64) error {
		log.WithField("deveui", deveui).Trace("called traceOnHanlder")		
		eui := deveui.String()
		if _, ok := t.maps[eui]; !ok {
			if len(t.maps) > MAP_MAX-1 {
				var keys []string
				for k := range t.maps {
					keys = append(keys, k)
				}
				log.WithField("traceList", keys).Errorf("Reached max lists (%d)", MAP_MAX)
				return nil
			}
			t.maps[eui] = true
			log.WithField("deveui", eui).Trace("added deveui to map")
			go t.TraceDevice(deveui)		
		} else {
			log.WithField("deveui", eui).Warn("alreay found in lists")
		}
		return nil
	}
}

func (t *TraceObject) traceOffHandler() func(lorawan.EUI64) error {
	return func(deveui lorawan.EUI64) error {
		log.WithField("deveui", deveui).Trace("called traceOffHanlder")
		eui := deveui.String()
		if _, ok := t.maps[eui]; ok {
			delete(t.maps, eui)
			log.WithField("deveui", eui).Trace("deleted deveui from map")
		} else {
			log.WithField("deveui", eui).Warn("not found in lists")
		}
		return nil
	}
}

func NewTrace(nsClient *NSClient, mqttClient *MqttClient) *TraceObject {
	obj := &TraceObject{
		maps: make(map[string]bool),
		nsClient: nsClient,
		mqtt: mqttClient,
	}
	obj.dataHandler = obj.sendFrameHandler()
	obj.onHandler = obj.traceOnHandler()
	obj.offHandler = obj.traceOffHandler()
	
	mqttClient.SetTraceOnHandler(obj.onHandler)
	mqttClient.SetTraceOffHandler(obj.offHandler)	

	return obj
}

func (t *TraceObject) Start() error {
	if err := t.mqtt.Connect(); err != nil {
		log.Fatal(err)
	}
	go t.mqtt.Run()
	return nil
}

func (t *TraceObject)TraceDevice(deveui lorawan.EUI64) error {
	req := ns.StreamFrameLogsForDeviceRequest{
		DevEui: deveui[:],
	}

	eui := deveui.String()

	ctx, cancel := context.WithCancel(context.Background())
	stream, err := t.nsClient.cl.StreamFrameLogsForDevice(ctx, &req)
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		appUp, appDown, err := convertUplinkAndDownlinkFrames(resp.GetUplinkFrameSet(), resp.GetDownlinkFrame(), true)

		if appUp != nil {
			phyJSON := appUp.PhyPayloadJson
			b, err := json.Marshal(appUp.TxInfo)
			if err != nil {
				log.Error(err)
			}
			txInfoJSON := string(b)
			b, err = json.Marshal(appUp.RxInfo)
			if err != nil {
				log.Error(err)
			}
			rxInfoJSON := string(b)
			str := fmt.Sprintf("{\"uplink\":{\"tx_info\":%s,\"rx_info\":%s,\"phy_payload\":%s}}\n", txInfoJSON, rxInfoJSON, phyJSON)
			t.dataHandler(deveui, []byte(str))
		}
		if appDown != nil {
			phyJSON := appDown.PhyPayloadJson
			b, err := json.Marshal(appDown.TxInfo)
			if err != nil {
				log.Error(err)
			}
			txInfoJSON := string(b)
			str := fmt.Sprintf("{\"downlink\":{\"tx_info\":%s,\"phy_payload\":%s}}\n", txInfoJSON, phyJSON)
			t.dataHandler(deveui, []byte(str))
		}
		if _, ok := t.maps[eui]; !ok {
			cancel()
		}
	}
	return nil
}

func convertUplinkAndDownlinkFrames(up *gw.UplinkFrameSet, down *gw.DownlinkFrame, decodeMACCommands bool) (*pb.UplinkFrameLog, *pb.DownlinkFrameLog, error) {
	var phy lorawan.PHYPayload

	if up != nil {
		if err := phy.UnmarshalBinary(up.PhyPayload); err != nil {
			return nil, nil, errors.Wrap(err, "unmarshal phypayload error")
		}
	}

	if down != nil {
		if err := phy.UnmarshalBinary(down.PhyPayload); err != nil {
			return nil, nil, errors.Wrap(err, "unmarshal phypayload error")
		}
	}

	if decodeMACCommands {
		switch v := phy.MACPayload.(type) {
		case *lorawan.MACPayload:
			if err := phy.DecodeFOptsToMACCommands(); err != nil {
				return nil, nil, errors.Wrap(err, "decode fopts to mac-commands error")
			}

			if v.FPort != nil && *v.FPort == 0 {
				if err := phy.DecodeFRMPayloadToMACCommands(); err != nil {
					return nil, nil, errors.Wrap(err, "decode frmpayload to mac-commands error")
				}
			}
		}
	}

	phyJSON, err := json.Marshal(phy)
	if err != nil {
		return nil, nil, errors.Wrap(err, "marshal phypayload error")
	}

	if up != nil {
		uplinkFrameLog := pb.UplinkFrameLog{
			TxInfo:         up.TxInfo,
			PhyPayloadJson: string(phyJSON),
		}

		for _, rxInfo := range up.RxInfo {
			var mac lorawan.EUI64
			copy(mac[:], rxInfo.GatewayId)

			upRXInfo := pb.UplinkRXInfo{
				GatewayId:         mac.String(),
				Time:              rxInfo.Time,
				TimeSinceGpsEpoch: rxInfo.TimeSinceGpsEpoch,
				Timestamp:         rxInfo.Timestamp,
				Rssi:              rxInfo.Rssi,
				LoraSnr:           rxInfo.LoraSnr,
				Channel:           rxInfo.Channel,
				RfChain:           rxInfo.RfChain,
				Board:             rxInfo.Board,
				Antenna:           rxInfo.Antenna,
				Location:          rxInfo.Location,
				FineTimestampType: rxInfo.FineTimestampType,
			}

			switch rxInfo.FineTimestampType {
			case gw.FineTimestampType_ENCRYPTED:
				fineTS := rxInfo.GetEncryptedFineTimestamp()
				if fineTS != nil {
					upRXInfo.FineTimestamp = &pb.UplinkRXInfo_EncryptedFineTimestamp{
						EncryptedFineTimestamp: &pb.EncryptedFineTimestamp{
							AesKeyIndex: fineTS.AesKeyIndex,
							EncryptedNs: fineTS.EncryptedNs,
							FpgaId:      hex.EncodeToString(fineTS.FpgaId),
						},
					}
				}
			case gw.FineTimestampType_PLAIN:
				fineTS := rxInfo.GetPlainFineTimestamp()
				if fineTS != nil {
					upRXInfo.FineTimestamp = &pb.UplinkRXInfo_PlainFineTimestamp{
						PlainFineTimestamp: fineTS,
					}
				}
			}

			uplinkFrameLog.RxInfo = append(uplinkFrameLog.RxInfo, &upRXInfo)
		}

		return &uplinkFrameLog, nil, nil
	}

	if down != nil {
		downlinkFrameLog := pb.DownlinkFrameLog{
			PhyPayloadJson: string(phyJSON),
		}

		if down.TxInfo != nil {
			var mac lorawan.EUI64
			copy(mac[:], down.TxInfo.GatewayId[:])

			downlinkFrameLog.TxInfo = &pb.DownlinkTXInfo{
				GatewayId:         mac.String(),
				Immediately:       down.TxInfo.Immediately,
				TimeSinceGpsEpoch: down.TxInfo.TimeSinceGpsEpoch,
				Timestamp:         down.TxInfo.Timestamp,
				Frequency:         down.TxInfo.Frequency,
				Power:             down.TxInfo.Power,
				Modulation:        down.TxInfo.Modulation,
				Board:             down.TxInfo.Board,
				Antenna:           down.TxInfo.Antenna,
			}

			if lora := down.TxInfo.GetLoraModulationInfo(); lora != nil {
				downlinkFrameLog.TxInfo.ModulationInfo = &pb.DownlinkTXInfo_LoraModulationInfo{
					LoraModulationInfo: lora,
				}
			}

			if fsk := down.TxInfo.GetFskModulationInfo(); fsk != nil {
				downlinkFrameLog.TxInfo.ModulationInfo = &pb.DownlinkTXInfo_FskModulationInfo{
					FskModulationInfo: fsk,
				}
			}
		}

		return nil, &downlinkFrameLog, nil
	}

	return nil, nil, nil
}
