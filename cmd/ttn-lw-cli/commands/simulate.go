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
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/simulate"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var (
	simulateUplinkFlags      = &pflag.FlagSet{}
	simulateJoinRequestFlags = &pflag.FlagSet{}
	simulateDataUplinkFlags  = &pflag.FlagSet{}

	applicationUplinkFlags = &pflag.FlagSet{}

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
	flagSet.Uint32("a_f_cnt_down", 0, "AFCntDown value for FOpts decryption of LoRaWAN 1.1+ frames")
	return flagSet
}

func startSimulation(
	cmd *cobra.Command, forUp func(*ttnpb.UplinkMessage) error, forDown func(*ttnpb.DownlinkMessage) error,
) error {
	gtwID, err := getGatewayID(cmd.Flags(), nil, true)
	if err != nil {
		return err
	}

	var uplinkParams ttnpb.SimulateMetadataParams
	if _, err := uplinkParams.SetFromFlags(simulateUplinkFlags, ""); err != nil {
		return err
	}
	if err := simulate.SetDefaults(&uplinkParams); err != nil {
		return err
	}

	upMsg := &ttnpb.UplinkMessage{
		Settings: &ttnpb.TxSettings{
			DataRate: &ttnpb.DataRate{
				Modulation: &ttnpb.DataRate_Lora{
					Lora: &ttnpb.LoRaDataRate{
						Bandwidth:       uplinkParams.Bandwidth,
						SpreadingFactor: uplinkParams.SpreadingFactor,
						CodingRate:      band.Cr4_5,
					},
				},
			},
			Frequency: uplinkParams.Frequency,
			Timestamp: uplinkParams.Timestamp,
			Time:      uplinkParams.Time,
		},
		RxMetadata: []*ttnpb.RxMetadata{
			{
				GatewayIds:  gtwID,
				Time:        uplinkParams.Time,
				Timestamp:   uplinkParams.Timestamp,
				Rssi:        uplinkParams.Rssi,
				ChannelRssi: uplinkParams.Rssi,
				Snr:         uplinkParams.Snr,
			},
		},
	}

	if err = forUp(upMsg); err != nil {
		return err
	}

	if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
		return io.Write(os.Stdout, config.OutputFormat, upMsg)
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

//nolint:gocyclo
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
			devEUI = types.MustEUI64(joinReq.DevEui).OrZero()
			joinEUI = types.MustEUI64(joinReq.JoinEui).OrZero()
			devNonce = types.MustDevNonce(joinReq.DevNonce).OrZero()
		} else if rejoinReq := lastUpMsg.GetRejoinRequestPayload(); rejoinReq != nil {
			devEUI = types.MustEUI64(rejoinReq.DevEui).OrZero()
			joinEUI = types.MustEUI64(rejoinReq.JoinEui).OrZero()
			devNonce = types.DevNonce{byte(rejoinReq.RejoinCnt), byte(rejoinReq.RejoinCnt >> 8)}
		}

		appKey := types.MustAES128Key(dev.GetRootKeys().GetAppKey().GetKey()).OrZero()
		nwkKey := types.MustAES128Key(dev.GetRootKeys().GetNwkKey().GetKey()).OrZero()

		var key types.AES128Key
		if macspec.UseNwkKey(dev.LorawanVersion) {
			key = nwkKey
		} else {
			key = appKey
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
		if macspec.UseNwkKey(dev.LorawanVersion) && joinAcceptPayload.DlSettings.OptNeg {
			jsIntKey := crypto.DeriveJSIntKey(key, devEUI)
			devNonce := types.MustDevNonce(lastUpMsg.GetJoinRequestPayload().DevNonce).OrZero()
			// TODO: Support RejoinRequest (https://github.com/TheThingsNetwork/lorawan-stack/issues/536)
			expectedMIC, err = crypto.ComputeJoinAcceptMIC(
				jsIntKey,
				0xFF,
				types.MustEUI64(dev.Ids.JoinEui).OrZero(),
				devNonce,
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

		devAddr := types.MustDevAddr(joinAcceptPayload.DevAddr).OrZero()
		dev.Ids.DevAddr, dev.Session.DevAddr = devAddr.Bytes(), devAddr.Bytes()
		dev.Session.Keys = &ttnpb.SessionKeys{}

		joinNonce := types.MustJoinNonce(joinAcceptPayload.JoinNonce).OrZero()
		if macspec.UseNwkKey(dev.LorawanVersion) && joinAcceptPayload.DlSettings.OptNeg {
			appSKey := crypto.DeriveAppSKey(appKey, joinNonce, joinEUI, devNonce)
			dev.Session.Keys.AppSKey = &ttnpb.KeyEnvelope{Key: appSKey.Bytes()}
			logger.Infof("Derived AppSKey %X (%s)", appSKey[:], base64.StdEncoding.EncodeToString(appSKey[:]))

			fNwkSIntKey := crypto.DeriveFNwkSIntKey(nwkKey, joinNonce, joinEUI, devNonce)
			dev.Session.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{Key: fNwkSIntKey.Bytes()}
			logger.Infof("Derived FNwkSIntKey %X (%s)", fNwkSIntKey[:], base64.StdEncoding.EncodeToString(fNwkSIntKey[:]))

			sNwkSIntKey := crypto.DeriveSNwkSIntKey(nwkKey, joinNonce, joinEUI, devNonce)
			dev.Session.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{Key: sNwkSIntKey.Bytes()}
			logger.Infof("Derived SNwkSIntKey %X (%s)", sNwkSIntKey[:], base64.StdEncoding.EncodeToString(sNwkSIntKey[:]))

			nwkSEncKey := crypto.DeriveNwkSEncKey(nwkKey, joinNonce, joinEUI, devNonce)
			dev.Session.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{Key: nwkSEncKey.Bytes()}
			logger.Infof("Derived NwkSEncKey %X (%s)", nwkSEncKey[:], base64.StdEncoding.EncodeToString(nwkSEncKey[:]))
		} else {
			netID := types.MustNetID(joinAcceptPayload.NetId).OrZero()
			appSKey := crypto.DeriveLegacyAppSKey(key, joinNonce, netID, devNonce)
			dev.Session.Keys.AppSKey = &ttnpb.KeyEnvelope{Key: appSKey.Bytes()}
			logger.Infof("Derived AppSKey %X (%s)", appSKey[:], base64.StdEncoding.EncodeToString(appSKey[:]))

			nwkSKey := crypto.DeriveLegacyNwkSKey(key, joinNonce, netID, devNonce)
			dev.Session.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{Key: nwkSKey.Bytes()}
			dev.Session.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{Key: nwkSKey.Bytes()}
			dev.Session.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{Key: nwkSKey.Bytes()}
			logger.Infof("Derived NwkSKey %X (%s)", nwkSKey[:], base64.StdEncoding.EncodeToString(nwkSKey[:]))
		}
	case ttnpb.MType_UNCONFIRMED_DOWN, ttnpb.MType_CONFIRMED_DOWN:
		macPayload := downMsg.Payload.GetMacPayload()
		devAddr := types.MustDevAddr(macPayload.FHdr.DevAddr).OrZero()

		var expectedMIC [4]byte
		if macspec.UseLegacyMIC(dev.LorawanVersion) {
			expectedMIC, err = crypto.ComputeLegacyDownlinkMIC(
				types.MustAES128Key(dev.Session.Keys.GetFNwkSIntKey().GetKey()).OrZero(),
				devAddr,
				macPayload.FHdr.FCnt,
				downMsg.RawPayload[:len(downMsg.RawPayload)-4],
			)
		} else {
			var confFCnt uint32
			if lastUpMsg.MHdr.MType == ttnpb.MType_CONFIRMED_UP {
				confFCnt = lastUpMsg.GetMacPayload().FHdr.FCnt
			}
			expectedMIC, err = crypto.ComputeDownlinkMIC(
				types.MustAES128Key(dev.Session.Keys.GetSNwkSIntKey().GetKey()).OrZero(),
				devAddr,
				confFCnt,
				macPayload.FHdr.FCnt,
				downMsg.RawPayload[:len(downMsg.RawPayload)-4],
			)
		}
		if err != nil {
			return err
		}
		if !bytes.Equal(downMsg.Payload.Mic, expectedMIC[:]) {
			logger.Warnf("Expected MIC %x but got %x", expectedMIC, downMsg.Payload.Mic)
		}

		var payloadKey types.AES128Key
		if macPayload.FPort == 0 {
			payloadKey = types.MustAES128Key(dev.Session.Keys.GetNwkSEncKey().GetKey()).OrZero()
		} else {
			payloadKey = types.MustAES128Key(dev.Session.Keys.GetAppSKey().GetKey()).OrZero()
			if cmdsInFOpts := len(macPayload.FHdr.FOpts) > 0; cmdsInFOpts && macspec.EncryptFOpts(dev.LorawanVersion) {
				fCnt := dev.Session.LastNFCntDown
				if macPayload.FPort != 0 {
					fCnt = dev.Session.LastAFCntDown
				}
				encOpts := macspec.EncryptionOptions(dev.LorawanVersion, macspec.DownlinkFrame, macPayload.FPort, cmdsInFOpts)
				fOpts, err := crypto.DecryptDownlink(
					types.MustAES128Key(dev.Session.Keys.GetNwkSEncKey().GetKey()).OrZero(),
					devAddr,
					fCnt,
					macPayload.FHdr.FOpts,
					encOpts...,
				)
				if err != nil {
					return err
				}
				macPayload.FHdr.FOpts = fOpts
			}
		}
		macPayload.FrmPayload, err = crypto.DecryptDownlink(
			payloadKey,
			devAddr,
			macPayload.FHdr.FCnt,
			macPayload.FrmPayload,
		)
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			var (
				uplinkParams ttnpb.SimulateMetadataParams
				joinParams   ttnpb.SimulateJoinRequestParams
				joinRequest  *ttnpb.Message
			)
			if _, err := uplinkParams.SetFromFlags(simulateUplinkFlags, ""); err != nil {
				return err
			}
			if err := simulate.SetDefaults(&uplinkParams); err != nil {
				return err
			}
			if _, err := joinParams.SetFromFlags(simulateJoinRequestFlags, ""); err != nil {
				return err
			}
			if err := uplinkParams.LorawanVersion.Validate(); err != nil {
				return errInvalidMACVersion.WithCause(err)
			}
			if err := uplinkParams.LorawanPhyVersion.Validate(); err != nil {
				return errInvalidPHYVersion.WithCause(err)
			}

			return startSimulation(cmd,
				func(upMsg *ttnpb.UplinkMessage) error {
					joinRequest = &ttnpb.Message{
						MHdr: &ttnpb.MHDR{
							MType: ttnpb.MType_JOIN_REQUEST,
							Major: ttnpb.Major_LORAWAN_R1,
						},
						Payload: &ttnpb.Message_JoinRequestPayload{
							JoinRequestPayload: &ttnpb.JoinRequestPayload{
								JoinEui:  joinParams.JoinEui,
								DevEui:   joinParams.DevEui,
								DevNonce: joinParams.DevNonce,
							},
						},
					}

					buf, err := lorawan.MarshalMessage(joinRequest)
					if err != nil {
						return err
					}
					var key *ttnpb.KeyEnvelope
					if macspec.UseNwkKey(uplinkParams.LorawanVersion) {
						key = joinParams.NwkKey
					} else {
						key = joinParams.AppKey
					}

					nwkKey, err := types.GetAES128Key(key.EncryptedKey)
					if err != nil {
						return err
					}

					mic, err := crypto.ComputeJoinRequestMIC(*nwkKey, buf)
					if err != nil {
						return err
					}
					joinRequest.Mic = mic[:]
					upMsg.RawPayload = append(buf, joinRequest.Mic...)
					return nil
				},
				func(downMsg *ttnpb.DownlinkMessage) error {
					if err := processDownlink(&ttnpb.EndDevice{
						LorawanVersion:    uplinkParams.LorawanVersion,
						LorawanPhyVersion: uplinkParams.LorawanPhyVersion,
						FrequencyPlanId:   uplinkParams.BandId,
						Ids: &ttnpb.EndDeviceIdentifiers{
							JoinEui: joinParams.JoinEui,
							DevEui:  joinParams.DevEui,
						},
						RootKeys: &ttnpb.RootKeys{
							NwkKey: joinParams.NwkKey,
							AppKey: joinParams.AppKey,
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			var (
				uplinkParams     ttnpb.SimulateMetadataParams
				dataUplinkParams ttnpb.SimulateDataUplinkParams
				dataUplink       *ttnpb.Message
			)
			if _, err := uplinkParams.SetFromFlags(simulateUplinkFlags, ""); err != nil {
				return err
			}
			if err := simulate.SetDefaults(&uplinkParams); err != nil {
				return err
			}
			if _, err := dataUplinkParams.SetFromFlags(simulateDataUplinkFlags, ""); err != nil {
				return err
			}
			if err := uplinkParams.LorawanVersion.Validate(); err != nil {
				return errInvalidMACVersion.WithCause(err)
			}
			if err := uplinkParams.LorawanPhyVersion.Validate(); err != nil {
				return errInvalidPHYVersion.WithCause(err)
			}

			return startSimulation(cmd,
				func(upMsg *ttnpb.UplinkMessage) error {
					fOpts := dataUplinkParams.FOpts
					if len(fOpts) > 0 && macspec.EncryptFOpts(uplinkParams.LorawanVersion) {
						encOpts := macspec.EncryptionOptions(
							uplinkParams.LorawanVersion,
							macspec.UplinkFrame,
							dataUplinkParams.FPort,
							true,
						)
						nwkSEncKey, err := types.GetAES128Key(dataUplinkParams.NwkSEncKey.EncryptedKey)
						if err != nil {
							return err
						}
						devAddr, err := types.GetDevAddr(dataUplinkParams.DevAddr)
						if err != nil {
							return err
						}
						buf, err := crypto.EncryptUplink(
							*nwkSEncKey,
							*devAddr,
							dataUplinkParams.FCnt,
							fOpts,
							encOpts...,
						)
						if err != nil {
							return err
						}
						fOpts = buf
					}

					var key *ttnpb.KeyEnvelope
					if dataUplinkParams.FPort == 0 {
						key = dataUplinkParams.NwkSEncKey
					} else {
						key = dataUplinkParams.AppSKey
					}

					nwkKey, err := types.GetAES128Key(key.EncryptedKey)
					if err != nil {
						return err
					}

					devAddr, err := types.GetDevAddr(dataUplinkParams.DevAddr)
					if err != nil {
						return err
					}

					frmPayload, err := crypto.EncryptUplink(
						*nwkKey,
						*devAddr,
						dataUplinkParams.FCnt,
						dataUplinkParams.FrmPayload,
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
										Adr:       dataUplinkParams.Adr,
										AdrAckReq: dataUplinkParams.AdrAckReq,
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

					buf, err := lorawan.MarshalMessage(dataUplink)
					if err != nil {
						return err
					}
					var (
						mic         [4]byte
						fNwkSIntKey *types.AES128Key
						sNwkSIntKey *types.AES128Key
					)

					fNwkSIntKey, err = types.GetAES128Key(dataUplinkParams.FNwkSIntKey.EncryptedKey)
					if err != nil {
						return err
					}
					if macspec.UseLegacyMIC(uplinkParams.LorawanVersion) {
						mic, err = crypto.ComputeLegacyUplinkMIC(
							*fNwkSIntKey,
							*devAddr,
							dataUplinkParams.FCnt,
							buf,
						)
					} else {
						sNwkSIntKey, err = types.GetAES128Key(dataUplinkParams.SNwkSIntKey.EncryptedKey)
						if err != nil {
							return err
						}
						mic, err = crypto.ComputeUplinkMIC(
							*sNwkSIntKey,
							*fNwkSIntKey,
							dataUplinkParams.ConfFCnt,
							uint8(dataUplinkParams.TxDrIdx),
							uint8(dataUplinkParams.TxChIdx),
							*devAddr,
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
					lastAFCntDown, _ := cmd.Flags().GetUint32("a_f_cnt_down")
					if err := processDownlink(&ttnpb.EndDevice{
						LorawanVersion:    uplinkParams.LorawanVersion,
						LorawanPhyVersion: uplinkParams.LorawanPhyVersion,
						FrequencyPlanId:   uplinkParams.BandId,
						Session: &ttnpb.Session{
							LastNFCntDown: lastNFCntDown,
							LastFCntUp:    lastAFCntDown,
							Keys: &ttnpb.SessionKeys{
								FNwkSIntKey: dataUplinkParams.FNwkSIntKey,
								SNwkSIntKey: dataUplinkParams.SNwkSIntKey,
								NwkSEncKey:  dataUplinkParams.NwkSEncKey,
								AppSKey:     dataUplinkParams.AppSKey,
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
			if _, err := uplinkMessage.SetFromFlags(applicationUplinkFlags, ""); err != nil {
				return err
			}

			up := &ttnpb.ApplicationUp{
				EndDeviceIds: devID,
				Up: &ttnpb.ApplicationUp_UplinkMessage{
					UplinkMessage: uplinkMessage,
				},
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
	ttnpb.AddSetFlagsForSimulateMetadataParams(simulateUplinkFlags, "", false)
	ttnpb.AddSetFlagsForSimulateJoinRequestParams(simulateJoinRequestFlags, "", false)
	ttnpb.AddSetFlagsForSimulateDataUplinkParams(simulateDataUplinkFlags, "", false)
	ttnpb.AddSetFlagsForApplicationUplink(applicationUplinkFlags, "", false)

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
