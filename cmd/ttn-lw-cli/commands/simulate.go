// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
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

var (
	errDataRate  = errors.DefineInvalidArgument("data_rate", "data rate is invalid")
	errFrequency = errors.DefineInvalidArgument("frequency", "frequency is invalid")
)

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
	if m.LoRaWANPHYVersion == ttnpb.PHYVersion_PHY_UNKNOWN {
		m.LoRaWANPHYVersion = ttnpb.PHYVersion_RP001_V1_0_2_REV_B
	}
	phy, err := band.Get(m.BandID, m.LoRaWANPHYVersion)
	if err != nil {
		return err
	}
	if m.Frequency == 0 {
		m.Frequency = phy.UplinkChannels[int(m.ChannelIndex)].Frequency
	} else if m.ChannelIndex == 0 {
		chIdx, err := func() (uint32, error) {
			for i, ch := range phy.UplinkChannels {
				if ch.Frequency == m.Frequency {
					return uint32(i), nil
				}
			}
			return 0, errFrequency.New()
		}()
		if err != nil {
			return err
		}
		m.ChannelIndex = chIdx
	}
	if m.Bandwidth == 0 || m.SpreadingFactor == 0 {
		drIdx := ttnpb.DataRateIndex(m.DataRateIndex)
		if drIdx < phy.UplinkChannels[m.ChannelIndex].MinDataRate || drIdx > phy.UplinkChannels[m.ChannelIndex].MaxDataRate {
			drIdx = phy.UplinkChannels[m.ChannelIndex].MaxDataRate
		}
		dr, ok := phy.DataRates[drIdx]
		if !ok {
			return errInvalidDataRateIndex.New()
		}
		lora := dr.Rate.GetLora()
		m.SpreadingFactor, m.Bandwidth = lora.SpreadingFactor, lora.Bandwidth
	} else if m.DataRateIndex == 0 {
		drIdx, err := func() (uint32, error) {
			for i, dr := range phy.DataRates {
				if lora := dr.Rate.GetLora(); lora != nil && lora.SpreadingFactor == m.SpreadingFactor && lora.Bandwidth == m.Bandwidth {
					return uint32(i), nil
				}
			}
			return 0, errDataRate.New()
		}()
		if err != nil {
			return err
		}
		m.DataRateIndex = drIdx
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
	FOpts       []byte          `protobuf:"bytes,16,opt,name=f_opts,json=fOpts,proto3" json:"f_opts,omitempty"`
}

var (
	simulateUplinkFlags      = util.FieldFlags(&simulateMetadataParams{})
	simulateJoinRequestFlags = util.FieldFlags(&simulateJoinRequestParams{})
	simulateDataUplinkFlags  = util.FieldFlags(&simulateDataUplinkParams{})

	applicationUplinkFlags = util.FieldFlags(&ttnpb.ApplicationUplink{})

	errApplicationServerDisabled = errors.DefineFailedPrecondition("application_server_disabled", "Application Server is disabled")
)

func simulateFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("gateway-api-key", "", "API key used for linking the gateway (optional when using user authentication)")
	flagSet.Bool("dry-run", false, "print the message instead of sending it")
	return flagSet
}

func simulateDownlinkFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Duration("timeout", 20*time.Second, "how long to wait for downlinks")
	flagSet.Int("downlinks", 1, "how many downlinks to expect")
	return flagSet
}

func simulateDataDownlinkFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Uint32("n_f_cnt_down", 0, "NFCntDown value for FOpts decryption of LoRaWAN 1.1+ frames")
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

	if err := uplinkParams.setDefaults(); err != nil {
		return err
	}

	upMsg := &ttnpb.UplinkMessage{
		Settings: &ttnpb.TxSettings{
			DataRate: &ttnpb.DataRate{
				Modulation: &ttnpb.DataRate_Lora{
					Lora: &ttnpb.LoRaDataRate{
						Bandwidth:       uplinkParams.Bandwidth,
						SpreadingFactor: uplinkParams.SpreadingFactor,
					},
				},
			},
			CodingRate: "4/5",
			Frequency:  uplinkParams.Frequency,
			Timestamp:  uplinkParams.Timestamp,
			Time:       ttnpb.ProtoTime(uplinkParams.Time),
		},
		RxMetadata: []*ttnpb.RxMetadata{
			{
				GatewayIds:  gtwID,
				Time:        ttnpb.ProtoTime(uplinkParams.Time),
				Timestamp:   uplinkParams.Timestamp,
				Rssi:        uplinkParams.RSSI,
				ChannelRssi: uplinkParams.RSSI,
				Snr:         uplinkParams.SNR,
			},
		},
	}

	if err = forUp(upMsg); err != nil {
		return err
	}

	if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
		if err = io.Write(os.Stdout, config.OutputFormat, upMsg); err != nil {
			return err
		}
		return nil
	}

	gs, err := api.Dial(ctx, config.GatewayServerGRPCAddress)
	if err != nil {
		return err
	}
	timeout, _ := cmd.Flags().GetDuration("timeout")
	linkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	md := rpcmetadata.MD{
		ID: gtwID.GatewayId,
	}
	if apiKey, _ := cmd.Flags().GetString("gateway-api-key"); apiKey != "" {
		md.AuthType = "Bearer"
		md.AuthValue = apiKey
	}
	linkCtx = md.ToOutgoingContext(linkCtx)
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

func processDownlink(dev *ttnpb.EndDevice, lastUpMsg *ttnpb.Message, downMsg *ttnpb.DownlinkMessage) error {
	phy, err := band.Get(dev.FrequencyPlanId, dev.LorawanPhyVersion)
	if err != nil {
		return err
	}

	downMsg.Payload = &ttnpb.Message{}
	if err = lorawan.UnmarshalMessage(downMsg.RawPayload, downMsg.Payload); err != nil {
		return err
	}
	switch downMsg.Payload.MHdr.MType {
	case ttnpb.MType_JOIN_ACCEPT:
		joinAcceptPayload := downMsg.Payload.GetJoinAcceptPayload()

		var devEUI, joinEUI types.EUI64
		var devNonce types.DevNonce
		if joinReq := lastUpMsg.GetJoinRequestPayload(); joinReq != nil {
			devEUI, joinEUI = joinReq.DevEui, joinReq.JoinEui
			devNonce = joinReq.DevNonce
		} else if rejoinReq := lastUpMsg.GetRejoinRequestPayload(); rejoinReq != nil {
			devEUI, joinEUI = rejoinReq.DevEui, rejoinReq.JoinEui
			devNonce = types.DevNonce{byte(rejoinReq.RejoinCnt), byte(rejoinReq.RejoinCnt >> 8)}
		}

		var key types.AES128Key
		if dev.LorawanVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
			key = *dev.GetRootKeys().GetNwkKey().Key
		} else {
			key = *dev.GetRootKeys().GetAppKey().Key
		}

		payload, err := crypto.DecryptJoinAccept(key, joinAcceptPayload.GetEncrypted())
		if err != nil {
			return err
		}

		joinAcceptBytes := payload[:len(payload)-4]
		downMsg.Payload.Mic = payload[len(payload)-4:]

		if err = lorawan.UnmarshalJoinAcceptPayload(joinAcceptBytes, joinAcceptPayload); err != nil {
			return err
		}

		var expectedMIC [4]byte
		if dev.LorawanVersion.Compare(ttnpb.MAC_V1_1) >= 0 && joinAcceptPayload.DlSettings.OptNeg {
			jsIntKey := crypto.DeriveJSIntKey(key, devEUI)
			// TODO: Support RejoinRequest (https://github.com/TheThingsNetwork/lorawan-stack/issues/536)
			expectedMIC, err = crypto.ComputeJoinAcceptMIC(
				jsIntKey,
				0xFF,
				*dev.Ids.JoinEui,
				lastUpMsg.GetJoinRequestPayload().DevNonce,
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
		if !bytes.Equal(downMsg.Payload.Mic, expectedMIC[:]) {
			logger.Warnf("Expected MIC %x but got %x", expectedMIC, downMsg.Payload.Mic)
		}

		dev.Ids.DevAddr, dev.Session.DevAddr = &joinAcceptPayload.DevAddr, joinAcceptPayload.DevAddr
		dev.Session.Keys = &ttnpb.SessionKeys{}

		if dev.LorawanVersion.Compare(ttnpb.MAC_V1_1) >= 0 && joinAcceptPayload.DlSettings.OptNeg {
			appSKey := crypto.DeriveAppSKey(*dev.GetRootKeys().GetAppKey().Key, joinAcceptPayload.JoinNonce, joinEUI, devNonce)
			dev.Session.Keys.AppSKey = &ttnpb.KeyEnvelope{Key: &appSKey}
			logger.Infof("Derived AppSKey %X (%s)", appSKey[:], base64.StdEncoding.EncodeToString(appSKey[:]))

			fNwkSIntKey := crypto.DeriveFNwkSIntKey(*dev.GetRootKeys().GetNwkKey().Key, joinAcceptPayload.JoinNonce, joinEUI, devNonce)
			dev.Session.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{Key: &fNwkSIntKey}
			logger.Infof("Derived FNwkSIntKey %X (%s)", fNwkSIntKey[:], base64.StdEncoding.EncodeToString(fNwkSIntKey[:]))

			sNwkSIntKey := crypto.DeriveSNwkSIntKey(*dev.GetRootKeys().GetNwkKey().Key, joinAcceptPayload.JoinNonce, joinEUI, devNonce)
			dev.Session.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{Key: &sNwkSIntKey}
			logger.Infof("Derived SNwkSIntKey %X (%s)", sNwkSIntKey[:], base64.StdEncoding.EncodeToString(sNwkSIntKey[:]))

			nwkSEncKey := crypto.DeriveNwkSEncKey(*dev.GetRootKeys().GetNwkKey().Key, joinAcceptPayload.JoinNonce, joinEUI, devNonce)
			dev.Session.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{Key: &nwkSEncKey}
			logger.Infof("Derived NwkSEncKey %X (%s)", nwkSEncKey[:], base64.StdEncoding.EncodeToString(nwkSEncKey[:]))
		} else {
			appSKey := crypto.DeriveLegacyAppSKey(key, joinAcceptPayload.JoinNonce, joinAcceptPayload.NetId, devNonce)
			dev.Session.Keys.AppSKey = &ttnpb.KeyEnvelope{Key: &appSKey}
			logger.Infof("Derived AppSKey %X (%s)", appSKey[:], base64.StdEncoding.EncodeToString(appSKey[:]))

			nwkSKey := crypto.DeriveLegacyNwkSKey(key, joinAcceptPayload.JoinNonce, joinAcceptPayload.NetId, devNonce)
			dev.Session.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{Key: &nwkSKey}
			dev.Session.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{Key: &nwkSKey}
			dev.Session.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{Key: &nwkSKey}
			logger.Infof("Derived NwkSKey %X (%s)", nwkSKey[:], base64.StdEncoding.EncodeToString(nwkSKey[:]))
		}
	case ttnpb.MType_UNCONFIRMED_DOWN, ttnpb.MType_CONFIRMED_DOWN:
		macPayload := downMsg.Payload.GetMacPayload()

		var expectedMIC [4]byte
		if dev.LorawanVersion.Compare(ttnpb.MAC_V1_1) < 0 {
			expectedMIC, err = crypto.ComputeLegacyDownlinkMIC(*dev.Session.Keys.GetFNwkSIntKey().Key, macPayload.FHdr.DevAddr, macPayload.FHdr.FCnt, downMsg.RawPayload[:len(downMsg.RawPayload)-4])
		} else {
			var confFCnt uint32
			if lastUpMsg.MHdr.MType == ttnpb.MType_CONFIRMED_UP {
				confFCnt = lastUpMsg.GetMacPayload().FHdr.FCnt
			}
			expectedMIC, err = crypto.ComputeDownlinkMIC(*dev.Session.Keys.GetSNwkSIntKey().Key, macPayload.FHdr.DevAddr, confFCnt, macPayload.FHdr.FCnt, downMsg.RawPayload[:len(downMsg.RawPayload)-4])
		}
		if err != nil {
			return err
		}
		if !bytes.Equal(downMsg.Payload.Mic, expectedMIC[:]) {
			logger.Warnf("Expected MIC %x but got %x", expectedMIC, downMsg.Payload.Mic)
		}

		var payloadKey types.AES128Key
		if macPayload.FPort == 0 {
			payloadKey = *dev.Session.Keys.GetNwkSEncKey().Key
		} else {
			payloadKey = *dev.Session.Keys.GetAppSKey().Key
			if len(macPayload.FHdr.FOpts) > 0 && dev.LorawanVersion.EncryptFOpts() {
				fOpts, err := crypto.DecryptDownlink(*dev.Session.Keys.GetNwkSEncKey().Key, macPayload.FHdr.DevAddr, dev.Session.LastNFCntDown, macPayload.FHdr.FOpts, true)
				if err != nil {
					return err
				}
				macPayload.FHdr.FOpts = fOpts
			}
		}
		macPayload.FrmPayload, err = crypto.DecryptDownlink(payloadKey, macPayload.FHdr.DevAddr, macPayload.FHdr.FCnt, macPayload.FrmPayload, false)
		if err != nil {
			return err
		}

		cmdBuf := macPayload.FHdr.FOpts
		if macPayload.FPort == 0 && len(macPayload.FrmPayload) > 0 {
			cmdBuf = macPayload.FrmPayload
		}
		var cmds []*ttnpb.MACCommand
		for r := bytes.NewReader(cmdBuf); r.Len() > 0; {
			cmd := &ttnpb.MACCommand{}
			if err := lorawan.DefaultMACCommands.ReadDownlink(phy, r, cmd); err != nil {
				logger.WithFields(log.Fields(
					"bytes_left", r.Len(),
					"mac_count", len(cmds),
				)).WithError(err).Warn("Failed to unmarshal MAC command")
				break
			}
			logger.WithField("cid", cmd.Cid).WithField("payload", cmd.GetPayload()).Info("Read MAC command")
			cmds = append(cmds, cmd)
		}
	}
	return nil
}

var (
	simulateCommand = &cobra.Command{
		Use:     "simulate",
		Aliases: []string{"sim"},
		Short:   "Simulation commands",
	}
	simulateJoinRequestCommand = &cobra.Command{
		Use:    "gateway-join-request",
		Short:  "Simulates a join request from an end device, sent through a simulated gateway connection (EXPERIMENTAL)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var uplinkParams simulateMetadataParams
			if err := util.SetFields(&uplinkParams, simulateUplinkFlags); err != nil {
				return err
			}
			if err := uplinkParams.setDefaults(); err != nil {
				return err
			}
			var joinParams simulateJoinRequestParams
			if err := util.SetFields(&joinParams, simulateJoinRequestFlags); err != nil {
				return err
			}

			if err := uplinkParams.LoRaWANVersion.Validate(); err != nil {
				return errInvalidMACVersion.WithCause(err)
			}

			if err := uplinkParams.LoRaWANPHYVersion.Validate(); err != nil {
				return errInvalidPHYVersion.WithCause(err)
			}

			var joinRequest *ttnpb.Message
			return simulate(cmd,
				func(upMsg *ttnpb.UplinkMessage) error {
					joinRequest = &ttnpb.Message{
						MHdr: &ttnpb.MHDR{
							MType: ttnpb.MType_JOIN_REQUEST,
							Major: ttnpb.Major_LORAWAN_R1,
						},
						Payload: &ttnpb.Message_JoinRequestPayload{
							JoinRequestPayload: &ttnpb.JoinRequestPayload{
								JoinEui:  joinParams.JoinEUI,
								DevEui:   joinParams.DevEUI,
								DevNonce: joinParams.DevNonce,
							},
						},
					}

					buf, err := lorawan.MarshalMessage(*joinRequest)
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
					joinRequest.Mic = mic[:]
					upMsg.RawPayload = append(buf, joinRequest.Mic...)
					return nil
				},
				func(downMsg *ttnpb.DownlinkMessage) error {
					if err := processDownlink(&ttnpb.EndDevice{
						LorawanVersion:    uplinkParams.LoRaWANVersion,
						LorawanPhyVersion: uplinkParams.LoRaWANPHYVersion,
						FrequencyPlanId:   uplinkParams.BandID,
						Ids: &ttnpb.EndDeviceIdentifiers{
							JoinEui: &joinParams.JoinEUI,
							DevEui:  &joinParams.DevEUI,
						},
						RootKeys: &ttnpb.RootKeys{
							NwkKey: &ttnpb.KeyEnvelope{Key: &joinParams.NwkKey},
							AppKey: &ttnpb.KeyEnvelope{Key: &joinParams.AppKey},
						},
						Session: &ttnpb.Session{},
					}, joinRequest, downMsg); err != nil {
						return err
					}
					// Here we can update a persistent end-device.
					return nil
				},
			)
		},
	}

	simulateDataUplinkCommand = &cobra.Command{
		Use:    "gateway-uplink",
		Short:  "Simulate an uplink message from an end device, sent through a simulated gateway connection (EXPERIMENTAL)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var uplinkParams simulateMetadataParams
			if err := util.SetFields(&uplinkParams, simulateUplinkFlags); err != nil {
				return err
			}
			if err := uplinkParams.setDefaults(); err != nil {
				return err
			}
			var dataUplinkParams simulateDataUplinkParams
			if err := util.SetFields(&dataUplinkParams, simulateDataUplinkFlags); err != nil {
				return err
			}

			if err := uplinkParams.LoRaWANVersion.Validate(); err != nil {
				return errInvalidMACVersion.WithCause(err)
			}

			if err := uplinkParams.LoRaWANPHYVersion.Validate(); err != nil {
				return errInvalidPHYVersion.WithCause(err)
			}

			var dataUplink *ttnpb.Message
			return simulate(cmd,
				func(upMsg *ttnpb.UplinkMessage) error {
					fOpts := dataUplinkParams.FOpts
					if len(fOpts) > 0 && uplinkParams.LoRaWANVersion.EncryptFOpts() {
						buf, err := crypto.EncryptUplink(
							dataUplinkParams.NwkSEncKey,
							dataUplinkParams.DevAddr,
							dataUplinkParams.FCnt,
							fOpts,
							true,
						)
						if err != nil {
							return err
						}
						fOpts = buf
					}

					var key types.AES128Key
					if dataUplinkParams.FPort == 0 {
						key = dataUplinkParams.NwkSEncKey
					} else {
						key = dataUplinkParams.AppSKey
					}
					frmPayload, err := crypto.EncryptUplink(
						key,
						dataUplinkParams.DevAddr,
						dataUplinkParams.FCnt,
						dataUplinkParams.FRMPayload,
						false,
					)
					if err != nil {
						return err
					}

					mType := ttnpb.MType_UNCONFIRMED_UP
					if dataUplinkParams.Confirmed {
						mType = ttnpb.MType_CONFIRMED_UP
					}
					dataUplink = &ttnpb.Message{
						MHdr: &ttnpb.MHDR{
							MType: mType,
							Major: ttnpb.Major_LORAWAN_R1,
						},
						Payload: &ttnpb.Message_MacPayload{
							MacPayload: &ttnpb.MACPayload{
								FHdr: &ttnpb.FHDR{
									DevAddr: dataUplinkParams.DevAddr,
									FCtrl: &ttnpb.FCtrl{
										Adr:       dataUplinkParams.ADR,
										AdrAckReq: dataUplinkParams.ADRAckReq,
										Ack:       dataUplinkParams.Ack,
									},
									FCnt:  dataUplinkParams.FCnt,
									FOpts: fOpts,
								},
								FPort:      dataUplinkParams.FPort,
								FrmPayload: frmPayload,
							},
						},
					}

					buf, err := lorawan.MarshalMessage(*dataUplink)
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
					if err != nil {
						return err
					}
					dataUplink.Mic = mic[:]
					upMsg.RawPayload = append(buf, dataUplink.Mic...)
					return nil
				},
				func(downMsg *ttnpb.DownlinkMessage) error {
					lastNFCntDown, _ := cmd.Flags().GetUint32("n_f_cnt_down")
					if err := processDownlink(&ttnpb.EndDevice{
						LorawanVersion:    uplinkParams.LoRaWANVersion,
						LorawanPhyVersion: uplinkParams.LoRaWANPHYVersion,
						FrequencyPlanId:   uplinkParams.BandID,
						Session: &ttnpb.Session{
							LastNFCntDown: lastNFCntDown,
							Keys: &ttnpb.SessionKeys{
								FNwkSIntKey: &ttnpb.KeyEnvelope{Key: &dataUplinkParams.FNwkSIntKey},
								SNwkSIntKey: &ttnpb.KeyEnvelope{Key: &dataUplinkParams.SNwkSIntKey},
								NwkSEncKey:  &ttnpb.KeyEnvelope{Key: &dataUplinkParams.NwkSEncKey},
								AppSKey:     &ttnpb.KeyEnvelope{Key: &dataUplinkParams.AppSKey},
							},
						},
					}, dataUplink, downMsg); err != nil {
						return err
					}
					// Here we can update a persistent end-device.
					return nil
				},
			)
		},
	}
	simulateApplicationUplinkCommand = &cobra.Command{
		Use:   "application-uplink [application-id] [device-id]",
		Short: "Simulate an application-layer uplink message from an end device, sent directly to the Application Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			if !config.ApplicationServerEnabled {
				return errApplicationServerDisabled.New()
			}
			uplinkMessage := &ttnpb.ApplicationUplink{}
			up := &ttnpb.ApplicationUp{
				EndDeviceIds: devID,
				Up: &ttnpb.ApplicationUp_UplinkMessage{
					UplinkMessage: uplinkMessage,
				},
			}
			if err := util.SetFields(uplinkMessage, applicationUplinkFlags); err != nil {
				return err
			}
			cc, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			if uplinkMessage.ReceivedAt == nil {
				uplinkMessage.ReceivedAt = ttnpb.ProtoTimePtr(time.Now())
			}
			_, err = ttnpb.NewAppAsClient(cc).SimulateUplink(ctx, up)
			return err
		},
	}
)

func init() {
	simulateJoinRequestCommand.Flags().AddFlagSet(gatewayIDFlags())
	simulateJoinRequestCommand.Flags().AddFlagSet(simulateUplinkFlags)
	simulateJoinRequestCommand.Flags().AddFlagSet(simulateDownlinkFlags())
	simulateJoinRequestCommand.Flags().AddFlagSet(simulateJoinRequestFlags)
	simulateJoinRequestCommand.Flags().AddFlagSet(simulateFlags())

	simulateCommand.AddCommand(simulateJoinRequestCommand)

	simulateDataUplinkCommand.Flags().AddFlagSet(gatewayIDFlags())
	simulateDataUplinkCommand.Flags().AddFlagSet(simulateUplinkFlags)
	simulateDataUplinkCommand.Flags().AddFlagSet(simulateDownlinkFlags())
	simulateDataUplinkCommand.Flags().AddFlagSet(simulateDataUplinkFlags)
	simulateDataUplinkCommand.Flags().AddFlagSet(simulateDataDownlinkFlags())
	simulateDataUplinkCommand.Flags().AddFlagSet(simulateFlags())

	simulateCommand.AddCommand(simulateDataUplinkCommand)

	simulateApplicationUplinkCommand.Flags().AddFlagSet(endDeviceIDFlags())
	simulateApplicationUplinkCommand.Flags().AddFlagSet(applicationUplinkFlags)

	simulateCommand.AddCommand(simulateApplicationUplinkCommand)

	Root.AddCommand(simulateCommand)
}
