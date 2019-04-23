// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

type simulateMetadataParams struct {
	RSSI              float32          `protobuf:"fixed32,1,opt,name=rssi,proto3" json:"rssi,omitempty"`
	SNR               float32          `protobuf:"fixed32,2,opt,name=snr,proto3" json:"snr,omitempty"`
	Timestamp         uint32           `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Time              *time.Time       `protobuf:"bytes,4,opt,name=time,proto3,stdtime" json:"time,omitempty"`
	LoRaWANVersion    ttnpb.MACVersion `protobuf:"varint,5,opt,name=lorawan_version,proto3,enum=ttn.lorawan.v3.MACVersion" json:"lorawan_version"`
	LoRaWANPHYVersion ttnpb.PHYVersion `protobuf:"varint,6,opt,name=lorawan_phy_version,proto3,enum=ttn.lorawan.v3.MACVersion" json:"lorawan_phy_version"`
	BandID            string           `protobuf:"bytes,7,opt,name=band_id,proto3,stdtime" json:"band_id,omitempty"`
	Frequency         uint64           `protobuf:"varint,8,opt,name=frequency,proto3" json:"frequency,omitempty"`
	ChannelIndex      uint32           `protobuf:"varint,9,opt,name=channel_index,proto3" json:"channel_index,omitempty"`
	Bandwidth         uint32           `protobuf:"varint,10,opt,name=bandwidth,proto3" json:"bandwidth,omitempty"`
	SpreadingFactor   uint32           `protobuf:"varint,11,opt,name=spreading_factor,proto3" json:"spreading_factor,omitempty"`
	DataRateIndex     uint32           `protobuf:"varint,12,opt,name=data_rate_index,proto3" json:"data_rate_index,omitempty"`
}

func (m *simulateMetadataParams) setDefaults() error {
	if m.Time == nil || m.Time.IsZero() {
		now := time.Now()
		m.Time = &now
	}
	if m.Timestamp == 0 {
		m.Timestamp = uint32(m.Time.UnixNano() / 1000)
	}
	if m.BandID == "" {
		m.BandID = band.EU_863_870
	}
	if m.LoRaWANPHYVersion == ttnpb.PHY_UNKNOWN {
		m.LoRaWANPHYVersion = ttnpb.PHY_V1_0_2_REV_B
	}
	phy, err := band.GetByID(m.BandID)
	if err != nil {
		return err
	}
	phy, err = phy.Version(m.LoRaWANPHYVersion)
	if err != nil {
		return err
	}
	if m.Frequency == 0 {
		m.Frequency = phy.UplinkChannels[int(m.ChannelIndex)].Frequency
	} else if m.ChannelIndex == 0 {
		for i, ch := range phy.UplinkChannels {
			if ch.Frequency == m.Frequency {
				m.ChannelIndex = uint32(i)
				break
			}
		}
	}
	if m.Bandwidth == 0 || m.SpreadingFactor == 0 {
		drIdx := int(m.DataRateIndex)
		if drIdx < int(phy.UplinkChannels[0].MinDataRate) || drIdx > int(phy.UplinkChannels[0].MaxDataRate) {
			drIdx = int(phy.UplinkChannels[0].MaxDataRate)
		}
		dr := phy.DataRates[drIdx].Rate.GetLoRa()
		m.SpreadingFactor, m.Bandwidth = dr.SpreadingFactor, dr.Bandwidth
	} else if m.DataRateIndex == 0 {
		for i, dr := range phy.DataRates {
			if dr.Rate.GetLoRa().SpreadingFactor == m.SpreadingFactor && dr.Rate.GetLoRa().Bandwidth == m.Bandwidth {
				m.DataRateIndex = uint32(i)
				break
			}
		}
	}
	return nil
}

type simulateJoinRequestParams struct {
	JoinEUI  types.EUI64     `protobuf:"bytes,1,opt,name=join_eui,proto3" json:"join_eui"`
	DevEUI   types.EUI64     `protobuf:"bytes,2,opt,name=dev_eui,proto3" json:"dev_eui"`
	DevNonce types.DevNonce  `protobuf:"bytes,3,opt,name=dev_nonce,proto3" json:"dev_nonce"`
	AppKey   types.AES128Key `protobuf:"bytes,4,opt,name=app_key,proto3" json:"app_key"`
	NwkKey   types.AES128Key `protobuf:"bytes,5,opt,name=nwk_key,proto3" json:"nwk_key"`
}

type simulateDataUplinkParams struct {
	DevAddr     types.DevAddr   `protobuf:"bytes,1,opt,name=dev_addr,proto3" json:"dev_addr"`
	FNwkSIntKey types.AES128Key `protobuf:"bytes,2,opt,name=f_nwk_s_int_key,proto3" json:"f_nwk_s_int_key"`
	SNwkSIntKey types.AES128Key `protobuf:"bytes,3,opt,name=s_nwk_s_int_key,proto3" json:"s_nwk_s_int_key"`
	NwkSEncKey  types.AES128Key `protobuf:"bytes,4,opt,name=nwk_s_enc_key,proto3" json:"nwk_s_enc_key"`
	AppSKey     types.AES128Key `protobuf:"bytes,5,opt,name=app_s_key,proto3" json:"app_s_key"`
	ADR         bool            `protobuf:"varint,6,opt,name=adr,proto3" json:"adr,omitempty"`
	ADRAckReq   bool            `protobuf:"varint,7,opt,name=adr_ack_req,json=adrAckReq,proto3" json:"adr_ack_req,omitempty"`
	Confirmed   bool            `protobuf:"varint,8,opt,name=confirmed,proto3" json:"confirmed,omitempty"`
	Ack         bool            `protobuf:"varint,9,opt,name=ack,proto3" json:"ack,omitempty"`
	FCnt        uint32          `protobuf:"varint,10,opt,name=f_cnt,json=fCnt,proto3" json:"f_cnt,omitempty"`
	FPort       uint32          `protobuf:"varint,11,opt,name=f_port,json=fPort,proto3" json:"f_port,omitempty"`
	FRMPayload  []byte          `protobuf:"bytes,12,opt,name=frm_payload,json=frmPayload,proto3" json:"frm_payload,omitempty"`
	ConfFCnt    uint32          `protobuf:"varint,13,opt,name=conf_f_cnt,json=confFCnt,proto3" json:"conf_f_cnt,omitempty"`
	TxDRIdx     uint32          `protobuf:"varint,14,opt,name=tx_dr_idx,json=txDRIdx,proto3" json:"tx_dr_idx,omitempty"`
	TxChIdx     uint32          `protobuf:"varint,15,opt,name=tx_ch_idx,json=txChIdx,proto3" json:"tx_ch_idx,omitempty"`
}

var (
	simulateUplinkFlags      = util.FieldFlags(&simulateMetadataParams{})
	simulateJoinRequestFlags = util.FieldFlags(&simulateJoinRequestParams{})
	simulateDataUplinkFlags  = util.FieldFlags(&simulateDataUplinkParams{})
)

func simulateDownlinkFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Duration("timeout", 20*time.Second, "how long to wait for downlinks")
	flagSet.Int("downlinks", 1, "how many downlinks to expect")
	return flagSet
}

func simulate(cmd *cobra.Command, forUp func(*ttnpb.UplinkMessage) error, forDown func(*ttnpb.DownlinkMessage) error) error {
	gtwID, err := getGatewayID(cmd.Flags(), nil, true)
	if err != nil {
		return err
	}

	var uplinkParams simulateMetadataParams
	if err := util.SetFields(&uplinkParams, simulateUplinkFlags); err != nil {
		return err
	}

	uplinkParams.setDefaults()

	upMsg := &ttnpb.UplinkMessage{
		Settings: ttnpb.TxSettings{
			DataRate: ttnpb.DataRate{
				Modulation: &ttnpb.DataRate_LoRa{
					LoRa: &ttnpb.LoRaDataRate{
						Bandwidth:       uplinkParams.Bandwidth,
						SpreadingFactor: uplinkParams.SpreadingFactor,
					},
				},
			},
			CodingRate: "4/5",
			Frequency:  uplinkParams.Frequency,
			Timestamp:  uplinkParams.Timestamp,
			Time:       uplinkParams.Time,
		},
		RxMetadata: []*ttnpb.RxMetadata{
			{
				GatewayIdentifiers: *gtwID,
				Time:               uplinkParams.Time,
				Timestamp:          uplinkParams.Timestamp,
				RSSI:               uplinkParams.RSSI,
				SNR:                uplinkParams.SNR,
			},
		},
	}

	if err = forUp(upMsg); err != nil {
		return err
	}

	gs, err := api.Dial(ctx, config.GatewayServerGRPCAddress)
	if err != nil {
		return err
	}
	timeout, _ := cmd.Flags().GetDuration("timeout")
	linkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	linkCtx = rpcmetadata.MD{ID: gtwID.GatewayID}.ToOutgoingContext(linkCtx)
	link, err := ttnpb.NewGtwGsClient(gs).LinkGateway(linkCtx)
	if err != nil {
		return err
	}

	// Send dummy up to start stream:
	if err = link.Send(&ttnpb.GatewayUp{}); err != nil {
		return err
	}

	sendTime := time.Now()
	if err = link.Send(&ttnpb.GatewayUp{UplinkMessages: []*ttnpb.UplinkMessage{upMsg}}); err != nil {
		return err
	}

	logger.Info("Sent uplink")

	expect, _ := cmd.Flags().GetInt("downlinks")
	for i := 0; i < expect; i++ {
		down, err := link.Recv()
		if err != nil {
			return err
		}
		timestampDifference := uint32(uint64(down.GetDownlinkMessage().GetScheduled().GetTimestamp()) + 1<<32 - uint64(upMsg.GetRxMetadata()[0].GetTimestamp()))
		logger.Infof("Received downlink (after %s) for transmission %s relative to uplink", time.Since(sendTime), time.Duration(timestampDifference)*1000)

		if err = forDown(down.DownlinkMessage); err != nil {
			return err
		}

		if err = io.Write(os.Stdout, config.OutputFormat, down.DownlinkMessage); err != nil {
			return err
		}
	}

	return ctx.Err()
}

func processDownlink(dev *ttnpb.EndDevice) func(lastUpMsg *ttnpb.UplinkMessage, downMsg *ttnpb.DownlinkMessage) error {
	return func(lastUpMsg *ttnpb.UplinkMessage, downMsg *ttnpb.DownlinkMessage) (err error) {
		phy, err := band.GetByID(dev.FrequencyPlanID)
		if err != nil {
			return err
		}
		phy, err = phy.Version(dev.LoRaWANPHYVersion)
		if err != nil {
			return err
		}

		downMsg.Payload = &ttnpb.Message{}
		if err = lorawan.UnmarshalMessage(downMsg.RawPayload, downMsg.Payload); err != nil {
			return err
		}
		switch downMsg.Payload.MType {
		case ttnpb.MType_JOIN_ACCEPT:
			joinAcceptPayload := downMsg.Payload.GetJoinAcceptPayload()

			var devEUI, joinEUI types.EUI64
			var devNonce types.DevNonce
			if joinReq := lastUpMsg.GetPayload().GetJoinRequestPayload(); joinReq != nil {
				devEUI, joinEUI = joinReq.DevEUI, joinReq.JoinEUI
				devNonce = joinReq.DevNonce
			} else if rejoinReq := lastUpMsg.GetPayload().GetRejoinRequestPayload(); rejoinReq != nil {
				devEUI, joinEUI = rejoinReq.DevEUI, rejoinReq.JoinEUI
				devNonce = types.DevNonce{byte(rejoinReq.RejoinCnt), byte(rejoinReq.RejoinCnt >> 8)}
			}

			var key types.AES128Key
			if dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
				key = *dev.GetRootKeys().GetNwkKey().Key
			} else {
				key = *dev.GetRootKeys().GetAppKey().Key
			}

			payload, err := crypto.DecryptJoinAccept(key, joinAcceptPayload.GetEncrypted())
			if err != nil {
				return err
			}

			joinAcceptBytes := payload[:len(payload)-4]
			downMsg.Payload.MIC = payload[len(payload)-4:]

			if err = lorawan.UnmarshalJoinAcceptPayload(joinAcceptBytes, joinAcceptPayload); err != nil {
				return err
			}

			var expectedMIC [4]byte
			if dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 && joinAcceptPayload.OptNeg {
				jsIntKey := crypto.DeriveJSIntKey(key, devEUI)
				// TODO: Support RejoinRequest (https://github.com/TheThingsNetwork/lorawan-stack/issues/536)
				expectedMIC, err = crypto.ComputeJoinAcceptMIC(
					jsIntKey,
					0xFF,
					*dev.JoinEUI,
					lastUpMsg.Payload.GetJoinRequestPayload().DevNonce,
					append([]byte{downMsg.RawPayload[0]}, joinAcceptBytes...),
				)
			} else {
				expectedMIC, err = crypto.ComputeLegacyJoinAcceptMIC(
					key,
					append([]byte{downMsg.RawPayload[0]}, joinAcceptBytes...),
				)
			}
			if err != nil {
				return err
			}
			if !bytes.Equal(downMsg.Payload.MIC, expectedMIC[:]) {
				logger.Warnf("Expected MIC %x but got %x", expectedMIC, downMsg.Payload.MIC)
			}

			dev.DevAddr, dev.Session.DevAddr = &joinAcceptPayload.DevAddr, joinAcceptPayload.DevAddr

			if dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 && joinAcceptPayload.OptNeg {
				var appKey types.AES128Key
				appKey = *dev.GetRootKeys().GetAppKey().Key
				appSKey := crypto.DeriveAppSKey(appKey, joinAcceptPayload.JoinNonce, joinEUI, devNonce)
				dev.Session.SessionKeys.AppSKey = &ttnpb.KeyEnvelope{Key: &appSKey}
				logger.Infof("Derived AppSKey %X (%s)", appSKey[:], base64.StdEncoding.EncodeToString(appSKey[:]))

				var nwkKey types.AES128Key
				nwkKey = *dev.GetRootKeys().GetNwkKey().Key
				fNwkSIntKey := crypto.DeriveFNwkSIntKey(nwkKey, joinAcceptPayload.JoinNonce, joinEUI, devNonce)
				dev.Session.SessionKeys.FNwkSIntKey = &ttnpb.KeyEnvelope{Key: &fNwkSIntKey}
				logger.Infof("Derived FNwkSIntKey %X (%s)", fNwkSIntKey[:], base64.StdEncoding.EncodeToString(fNwkSIntKey[:]))
				sNwkSIntKey := crypto.DeriveSNwkSIntKey(nwkKey, joinAcceptPayload.JoinNonce, joinEUI, devNonce)
				dev.Session.SessionKeys.SNwkSIntKey = &ttnpb.KeyEnvelope{Key: &sNwkSIntKey}
				logger.Infof("Derived SNwkSIntKey %X (%s)", sNwkSIntKey[:], base64.StdEncoding.EncodeToString(sNwkSIntKey[:]))
				nwkSEncKey := crypto.DeriveNwkSEncKey(nwkKey, joinAcceptPayload.JoinNonce, joinEUI, devNonce)
				dev.Session.SessionKeys.NwkSEncKey = &ttnpb.KeyEnvelope{Key: &nwkSEncKey}
				logger.Infof("Derived NwkSEncKey %X (%s)", nwkSEncKey[:], base64.StdEncoding.EncodeToString(nwkSEncKey[:]))
			} else {
				appSKey := crypto.DeriveLegacyAppSKey(key, joinAcceptPayload.JoinNonce, joinAcceptPayload.NetID, devNonce)
				dev.Session.SessionKeys.AppSKey = &ttnpb.KeyEnvelope{Key: &appSKey}
				logger.Infof("Derived AppSKey %X (%s)", appSKey[:], base64.StdEncoding.EncodeToString(appSKey[:]))
				nwkSKey := crypto.DeriveLegacyNwkSKey(key, joinAcceptPayload.JoinNonce, joinAcceptPayload.NetID, devNonce)
				dev.Session.SessionKeys.FNwkSIntKey = &ttnpb.KeyEnvelope{Key: &nwkSKey}
				dev.Session.SessionKeys.SNwkSIntKey = &ttnpb.KeyEnvelope{Key: &nwkSKey}
				dev.Session.SessionKeys.NwkSEncKey = &ttnpb.KeyEnvelope{Key: &nwkSKey}
				logger.Infof("Derived NwkSKey %X (%s)", nwkSKey[:], base64.StdEncoding.EncodeToString(nwkSKey[:]))
			}
		case ttnpb.MType_UNCONFIRMED_DOWN, ttnpb.MType_CONFIRMED_DOWN:
			macPayload := downMsg.Payload.GetMACPayload()

			var expectedMIC [4]byte
			if dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
				var key types.AES128Key
				key = *dev.Session.SessionKeys.GetFNwkSIntKey().Key
				expectedMIC, err = crypto.ComputeLegacyDownlinkMIC(key, macPayload.DevAddr, macPayload.FCnt, downMsg.RawPayload[:len(downMsg.RawPayload)-4])
			} else {
				var key types.AES128Key
				key = *dev.Session.SessionKeys.GetSNwkSIntKey().Key
				expectedMIC, err = crypto.ComputeDownlinkMIC(key, macPayload.DevAddr, lastUpMsg.GetPayload().GetMACPayload().FCnt, macPayload.FCnt, downMsg.RawPayload[:len(downMsg.RawPayload)-4])
			}
			if err != nil {
				return err
			}
			if !bytes.Equal(downMsg.Payload.MIC, expectedMIC[:]) {
				logger.Warnf("Expected MIC %x but got %x", expectedMIC, downMsg.Payload.MIC)
			}

			var key types.AES128Key
			if macPayload.FPort == 0 {
				key = *dev.Session.SessionKeys.GetNwkSEncKey().Key
			} else {
				key = *dev.Session.SessionKeys.GetAppSKey().Key
			}

			macPayload.FRMPayload, err = crypto.DecryptDownlink(key, macPayload.DevAddr, macPayload.FCnt, macPayload.FRMPayload)
			if err != nil {
				return err
			}

			mac := macPayload.FOpts
			if macPayload.FPort == 0 {
				mac = macPayload.FRMPayload
			}
			var cmds []*ttnpb.MACCommand
			for r := bytes.NewReader(mac); r.Len() > 0; {
				cmd := &ttnpb.MACCommand{}
				if err := lorawan.DefaultMACCommands.ReadDownlink(phy, r, cmd); err != nil {
					logger.WithFields(log.Fields(
						"bytes_left", r.Len(),
						"mac_count", len(cmds),
					)).WithError(err).Warn("Failed to unmarshal MAC command")
					break
				}
				logger.WithField("cid", cmd.CID).WithField("payload", cmd.GetPayload()).Info("Read MAC command")
				cmds = append(cmds, cmd)
			}
		}
		return nil
	}
}

var (
	simulateCommand = &cobra.Command{
		Use:     "simulate",
		Aliases: []string{"sim"},
		Short:   "Simulation commands (EXPERIMENTAL)",
		Hidden:  true,
	}
	simulateJoinRequestCommand = &cobra.Command{
		Use:   "join-request",
		Short: "Simulate a join request (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			var uplinkParams simulateMetadataParams
			if err := util.SetFields(&uplinkParams, simulateUplinkFlags); err != nil {
				return err
			}
			uplinkParams.setDefaults()
			var joinParams simulateJoinRequestParams
			if err := util.SetFields(&joinParams, simulateJoinRequestFlags); err != nil {
				return err
			}

			processDownlink := processDownlink(&ttnpb.EndDevice{
				LoRaWANVersion:    uplinkParams.LoRaWANVersion,
				LoRaWANPHYVersion: uplinkParams.LoRaWANPHYVersion,
				FrequencyPlanID:   uplinkParams.BandID,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					JoinEUI: &joinParams.JoinEUI,
					DevEUI:  &joinParams.DevEUI,
				},
				RootKeys: &ttnpb.RootKeys{
					NwkKey: &ttnpb.KeyEnvelope{Key: &joinParams.NwkKey},
					AppKey: &ttnpb.KeyEnvelope{Key: &joinParams.AppKey},
				},
				Session: &ttnpb.Session{},
			})

			var joinRequest *ttnpb.UplinkMessage

			return simulate(cmd,
				func(upMsg *ttnpb.UplinkMessage) error {
					joinRequest = upMsg

					upMsg.Payload = &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_JOIN_REQUEST,
							Major: ttnpb.Major_LORAWAN_R1,
						},
						Payload: &ttnpb.Message_JoinRequestPayload{
							JoinRequestPayload: &ttnpb.JoinRequestPayload{
								JoinEUI:  joinParams.JoinEUI,
								DevEUI:   joinParams.DevEUI,
								DevNonce: joinParams.DevNonce,
							},
						},
					}

					buf, err := lorawan.MarshalMessage(*upMsg.Payload)
					if err != nil {
						return err
					}
					var key types.AES128Key
					if uplinkParams.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
						key = joinParams.NwkKey
					} else {
						key = joinParams.AppKey
					}
					mic, err := crypto.ComputeJoinRequestMIC(key, buf)
					if err != nil {
						return err
					}
					upMsg.Payload.MIC = mic[:]
					upMsg.RawPayload = append(buf, upMsg.Payload.MIC...)

					return nil
				},
				func(downMsg *ttnpb.DownlinkMessage) error {
					if err := processDownlink(joinRequest, downMsg); err != nil {
						return err
					}
					// Here we can update a persistent end-device.
					return nil
				},
			)
		},
	}
	simulateDataUplinkCommand = &cobra.Command{
		Use:   "uplink",
		Short: "Simulate a data uplink (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			var uplinkParams simulateMetadataParams
			if err := util.SetFields(&uplinkParams, simulateUplinkFlags); err != nil {
				return err
			}
			uplinkParams.setDefaults()
			var dataUplinkParams simulateDataUplinkParams
			if err := util.SetFields(&dataUplinkParams, simulateDataUplinkFlags); err != nil {
				return err
			}

			processDownlink := processDownlink(&ttnpb.EndDevice{
				LoRaWANVersion:    uplinkParams.LoRaWANVersion,
				LoRaWANPHYVersion: uplinkParams.LoRaWANPHYVersion,
				FrequencyPlanID:   uplinkParams.BandID,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						FNwkSIntKey: &ttnpb.KeyEnvelope{Key: &dataUplinkParams.FNwkSIntKey},
						SNwkSIntKey: &ttnpb.KeyEnvelope{Key: &dataUplinkParams.SNwkSIntKey},
						NwkSEncKey:  &ttnpb.KeyEnvelope{Key: &dataUplinkParams.NwkSEncKey},
						AppSKey:     &ttnpb.KeyEnvelope{Key: &dataUplinkParams.AppSKey},
					},
				},
			})

			var dataUplink *ttnpb.UplinkMessage

			return simulate(cmd,
				func(upMsg *ttnpb.UplinkMessage) error {
					dataUplink = upMsg

					macPayload := &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: dataUplinkParams.DevAddr,
							FCtrl: ttnpb.FCtrl{
								ADR:       dataUplinkParams.ADR,
								ADRAckReq: dataUplinkParams.ADRAckReq,
								Ack:       dataUplinkParams.Ack,
							},
							FCnt: dataUplinkParams.FCnt,
						},
						FPort: dataUplinkParams.FPort,
					}

					key := dataUplinkParams.AppSKey
					if dataUplinkParams.FPort == 0 {
						key = dataUplinkParams.NwkSEncKey
					}
					buf, err := crypto.EncryptUplink(
						key,
						dataUplinkParams.DevAddr,
						dataUplinkParams.FCnt,
						dataUplinkParams.FRMPayload,
					)
					if err != nil {
						return err
					}
					macPayload.FRMPayload = buf

					upMsg.Payload = &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
							Major: ttnpb.Major_LORAWAN_R1,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: macPayload},
					}
					if dataUplinkParams.Confirmed {
						upMsg.Payload.MType = ttnpb.MType_CONFIRMED_UP
					}
					buf, err = lorawan.MarshalMessage(*upMsg.Payload)
					if err != nil {
						return err
					}

					var mic [4]byte
					if uplinkParams.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
						mic, err = crypto.ComputeUplinkMIC(
							dataUplinkParams.SNwkSIntKey,
							dataUplinkParams.FNwkSIntKey,
							dataUplinkParams.ConfFCnt,
							uint8(dataUplinkParams.TxDRIdx),
							uint8(dataUplinkParams.TxChIdx),
							dataUplinkParams.DevAddr,
							dataUplinkParams.FCnt,
							buf,
						)
					} else {
						mic, err = crypto.ComputeLegacyUplinkMIC(
							dataUplinkParams.FNwkSIntKey,
							dataUplinkParams.DevAddr,
							dataUplinkParams.FCnt,
							buf,
						)
					}
					upMsg.Payload.MIC = mic[:]
					upMsg.RawPayload = append(buf, upMsg.Payload.MIC...)

					return nil
				},
				func(downMsg *ttnpb.DownlinkMessage) error {
					if err := processDownlink(dataUplink, downMsg); err != nil {
						return err
					}
					// Here we can update a persistent end-device.
					return nil
				},
			)
		},
	}
)

func init() {
	simulateJoinRequestCommand.Flags().AddFlagSet(gatewayIDFlags())
	simulateJoinRequestCommand.Flags().AddFlagSet(simulateUplinkFlags)
	simulateJoinRequestCommand.Flags().AddFlagSet(simulateDownlinkFlags())
	simulateJoinRequestCommand.Flags().AddFlagSet(simulateJoinRequestFlags)

	simulateCommand.AddCommand(simulateJoinRequestCommand)

	simulateDataUplinkCommand.Flags().AddFlagSet(gatewayIDFlags())
	simulateDataUplinkCommand.Flags().AddFlagSet(simulateUplinkFlags)
	simulateDataUplinkCommand.Flags().AddFlagSet(simulateDownlinkFlags())
	simulateDataUplinkCommand.Flags().AddFlagSet(simulateDataUplinkFlags)

	simulateCommand.AddCommand(simulateDataUplinkCommand)

	Root.AddCommand(simulateCommand)
}
