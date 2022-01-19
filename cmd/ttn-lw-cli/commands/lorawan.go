// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"fmt"
	"os"

	"github.com/mohae/deepcopy"
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

type lorawanDecodedFrame struct {
	Message     ttnpb.Message       `json:"message"`
	MACCommands []*ttnpb.MACCommand `json:"mac_commands,omitempty"`
}

type lorawanConfig struct {
	Band       band.Band
	MACVersion ttnpb.MACVersion
	PHYVersion ttnpb.PHYVersion

	AppKey,
	NwkKey,
	AppSKey,
	NwkSKey,
	NwkSEncKey,
	FNwkSIntKey,
	SNwkSIntKey types.AES128Key
}

func (c *lorawanConfig) getJoinAcceptDecodeKey() (types.AES128Key, string) {
	if c.MACVersion.Compare(ttnpb.MACVersion_MAC_V1_1) < 0 {
		return c.AppKey, "AppKey"
	} else {
		return c.NwkKey, "NwkKey"
	}
}

func getMacBuffer(p *ttnpb.MACPayload) []byte {
	if p.FPort == 0 && len(p.FrmPayload) > 0 {
		return p.FrmPayload
	}
	return p.FHdr.FOpts
}

func setMacBuffer(p *ttnpb.MACPayload, buf []byte) {
	if p.FPort == 0 && len(p.FrmPayload) > 0 {
		p.FrmPayload = buf
	} else {
		p.FHdr.FOpts = buf
	}
}

func decodeJoinRequest(msg ttnpb.Message, config lorawanConfig) (*lorawanDecodedFrame, error) {
	return &lorawanDecodedFrame{
		Message: msg,
	}, nil
}

func decodeUplink(msg ttnpb.Message, config lorawanConfig) (*lorawanDecodedFrame, error) {
	pld := msg.GetMacPayload()
	macBuf := getMacBuffer(pld)
	if len(macBuf) > 0 && (len(pld.FHdr.FOpts) == 0 || config.MACVersion.EncryptFOpts()) {
		if config.NwkSEncKey.IsZero() {
			logger.Warn("No NwkSEncKey provided, skipping decryption of MAC buffer")
		} else {
			logger.Debug("Decrypting MAC buffer")
			for msb := uint32(0); msb < 0xff; msb++ {
				fCnt := msb<<8 | pld.FHdr.FCnt
				macBuf, err := crypto.DecryptUplink(config.NwkSEncKey, pld.FHdr.DevAddr, fCnt, macBuf, pld.FPort != 0)
				if err == nil {
					setMacBuffer(pld, macBuf)
					break
				}
				logger.WithField("f_cnt", fCnt).Debug("Failed attempt to decrypt MAC buffer")
			}
		}
	}
	var macCommands []*ttnpb.MACCommand
	for r := bytes.NewReader(macBuf); r.Len() > 0; {
		cmd := &ttnpb.MACCommand{}
		if err := lorawan.DefaultMACCommands.ReadUplink(config.Band, r, cmd); err != nil {
			logger.WithError(err).Warn("Failed to read MAC command")
		} else {
			macCommands = append(macCommands, cmd)
		}
	}
	if pld.FPort > 0 {
		if config.AppSKey.IsZero() {
			logger.Warn("No AppSKey provided, skipping application payload decryption")
		} else {
			logger.Debug("Decrypting application payload")
			buf, err := crypto.DecryptUplink(config.AppSKey, pld.FHdr.DevAddr, pld.FHdr.FCnt, pld.FrmPayload, false)
			if err != nil {
				logger.WithField("f_cnt", pld.FHdr.FCnt).Debug("Failed attempt to decrypt FrmPayload")
			} else {
				pld.FrmPayload = buf
			}
		}
	}
	msg.Payload = &ttnpb.Message_MacPayload{
		MacPayload: pld,
	}
	if !config.FNwkSIntKey.IsZero() {
		logger.Debug("Verification of the uplink message MIC is not implemented yet")
	}
	return &lorawanDecodedFrame{
		Message:     msg,
		MACCommands: macCommands,
	}, nil
}

func decodeJoinAccept(msg ttnpb.Message, config lorawanConfig) (*lorawanDecodedFrame, error) {
	pld := msg.GetJoinAcceptPayload()
	key, keyName := config.getJoinAcceptDecodeKey()
	if key.IsZero() {
		logger.Warnf("No %s provided, skipping join accept decryption", keyName)
	} else {
		buf, err := crypto.DecryptJoinAccept(key, pld.Encrypted)
		if err != nil {
			return nil, err
		}
		n := len(buf)
		if n < 4 {
			logger.WithFields(log.Fields("length", n, "minimum", 4)).Warn("Invalid Join Accept message length")
			return &lorawanDecodedFrame{Message: msg}, nil
		}
		buf, mic := buf[:n-4], buf[n-4:]
		decBuf := deepcopy.Copy(pld).(*ttnpb.JoinAcceptPayload)
		if err := lorawan.UnmarshalJoinAcceptPayload(buf, decBuf); err != nil {
			logger.WithError(err).Warn("Failed to unmarshal join accept payload")
			return &lorawanDecodedFrame{Message: msg}, nil
		}

		msg.Mic = mic
		msg.Payload = &ttnpb.Message_JoinAcceptPayload{
			JoinAcceptPayload: decBuf,
		}
	}
	return &lorawanDecodedFrame{Message: msg}, nil
}

func decodeDownlink(msg ttnpb.Message, config lorawanConfig) (*lorawanDecodedFrame, error) {
	pld := msg.GetMacPayload()

	macBuf := getMacBuffer(pld)
	if len(macBuf) > 0 && (len(pld.FHdr.FOpts) == 0 || config.MACVersion.EncryptFOpts()) && !config.NwkSKey.IsZero() {
		logger.Debug("Decrypting MAC buffer")
		for msb := uint32(0); msb < 0xffff; msb++ {
			fCnt := msb<<16 | pld.FHdr.FCnt
			macBuf, err := crypto.DecryptDownlink(config.NwkSKey, pld.FHdr.DevAddr, fCnt, macBuf, pld.FPort != 0)
			if err == nil {
				setMacBuffer(pld, macBuf)
				break
			}
			logger.WithField("f_cnt", fCnt).Debug("Failed attempt to decrypt MAC buffer")
		}
	}
	var macCommands []*ttnpb.MACCommand
	for r := bytes.NewReader(macBuf); r.Len() > 0; {
		cmd := &ttnpb.MACCommand{}
		if err := lorawan.DefaultMACCommands.ReadDownlink(config.Band, r, cmd); err != nil {
			logger.WithError(err).Warn("Failed to read MAC command")
		} else {
			macCommands = append(macCommands, cmd)
		}
	}
	if pld.FPort > 0 {
		if config.AppSKey.IsZero() {
			logger.Warn("No AppSKey provided, skipping application payload decryption")
		} else {
			logger.Debug("Decrypting application payload")
			buf, err := crypto.DecryptDownlink(config.AppSKey, pld.FHdr.DevAddr, pld.FHdr.FCnt, pld.FrmPayload, false)
			if err != nil {
				logger.WithField("f_cnt", pld.FHdr.FCnt).Debug("Failed attempt to decrypt FrmPayload")
			} else {
				pld.FrmPayload = buf
			}
		}
	}

	msg.Payload = &ttnpb.Message_MacPayload{
		MacPayload: pld,
	}

	if !config.FNwkSIntKey.IsZero() {
		logger.Debug("Verification of the uplink message MIC is not implemented yet")
	}
	return &lorawanDecodedFrame{
		Message:     msg,
		MACCommands: macCommands,
	}, nil
}

func decodeFrame(msg ttnpb.Message, config lorawanConfig) (*lorawanDecodedFrame, error) {
	switch msg.MHdr.MType {
	case ttnpb.MType_JOIN_REQUEST:
		return decodeJoinRequest(msg, config)
	case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
		return decodeUplink(msg, config)
	case ttnpb.MType_JOIN_ACCEPT:
		return decodeJoinAccept(msg, config)
	case ttnpb.MType_CONFIRMED_DOWN, ttnpb.MType_UNCONFIRMED_DOWN:
		return decodeDownlink(msg, config)
	}

	return nil, fmt.Errorf("not implemented")
}

var (
	lorawanCmd = &cobra.Command{
		Use:     "lorawan",
		Aliases: []string{"lw"},
		Short:   "LoRaWAN commands",
	}
	lorawanDecodeCmd = &cobra.Command{
		Use:     "decode",
		Aliases: []string{"d"},
		Short:   "Decode LoRaWAN frames",
		Example: `
  Join Request:
    $ echo 'AFP6A9B+1bNwFgIcAAujBABERDaumME=' | ttn-lw-cli lorawan decode --input-format base64
    $ echo '0053fa03d07ed5b37016021c000ba30400444436ae98c1' | ttn-lw-cli lorawan decode --input-format hex

  Join Accept:
    $ echo 'IAUNJTHDK7t2zM+eeFmGIyjAlSyqfNfAWPzZTjhcVfAg' | ttn-lw-cli lorawan decode --input-format base64
    $ echo 'IAUNJTHDK7t2zM+eeFmGIyjAlSyqfNfAWPzZTjhcVfAg' | ttn-lw-cli lorawan decode --input-format base64 --app-key 5CF2BD4810FD92E9271050D2541A0F2B
    $ echo 'IAUNJTHDK7t2zM+eeFmGIyjAlSyqfNfAWPzZTjhcVfAg' | ttn-lw-cli lorawan decode --input-format base64 --lorawan-version 1.1 --nwk-key 5CF2BD4810FD92E9271050D2541A0F2B

  Example Network Uplink:
    $ echo 'QL8AACeFAQADBwb/CP6z6aY=' | ttn-lw-cli lorawan decode --input-format base64

  Example Data Uplink:
    $ echo 'QD7U3QEAEgABK7VS98g=' | ttn-lw-cli lorawan decode --input-format base64
    $ echo 'QD7U3QEAEgABK7VS98g=' | ttn-lw-cli lorawan decode --input-format base64 --app-s-key CAE4B67DA7EA96144AFD687CD1EF1F23
		`,
		PersistentPreRunE: preRun(),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch config.InputFormat {
			case "hex", "base64":
			default:
				return fmt.Errorf("command supports only hex and base64 input formats")
			}
			lorawanConfig := lorawanConfig{}
			lorawanVersionStr, _ := cmd.Flags().GetString("lorawan-version")
			if err := lorawanConfig.MACVersion.UnmarshalText([]byte(lorawanVersionStr)); err != nil {
				return err
			}
			lorawanPhyVersionStr, _ := cmd.Flags().GetString("lorawan-phy-version")
			if err := lorawanConfig.PHYVersion.UnmarshalText([]byte(lorawanPhyVersionStr)); err != nil {
				return err
			}
			for _, key := range []struct {
				flag string
				t    *types.AES128Key
			}{
				{flag: "app-key", t: &lorawanConfig.AppKey},
				{flag: "nwk-key", t: &lorawanConfig.NwkKey},
				{flag: "app-s-key", t: &lorawanConfig.AppSKey},
				{flag: "nwk-s-key", t: &lorawanConfig.NwkSKey},
				{flag: "f-nwk-s-int-key", t: &lorawanConfig.FNwkSIntKey},
				{flag: "s-nwk-s-int-key", t: &lorawanConfig.SNwkSIntKey},
				{flag: "nwk-s-enc-key", t: &lorawanConfig.NwkSEncKey},
			} {
				flagStr, _ := cmd.Flags().GetString(key.flag)
				if flagStr != "" {
					if err := key.t.UnmarshalText([]byte(flagStr)); err != nil {
						return err
					}
				}
			}

			if lorawanConfig.MACVersion.Compare(ttnpb.MACVersion_MAC_V1_1) < 0 {
				for _, key := range []*types.AES128Key{&lorawanConfig.FNwkSIntKey, &lorawanConfig.SNwkSIntKey, &lorawanConfig.NwkSEncKey} {
					if key.IsZero() {
						*key = lorawanConfig.NwkSKey
					}
				}
			}

			bandID, _ := cmd.Flags().GetString("band")
			if bandID != "" {
				band, err := band.Get(bandID, lorawanConfig.PHYVersion)
				if err != nil {
					return err
				}
				lorawanConfig.Band = band
			}

			return asBulk(func(cmd *cobra.Command, args []string) error {
				if inputDecoder == nil {
					return nil
				}
				var input []byte
				if _, err := inputDecoder.Decode(&input); err != nil {
					return err
				}
				var frame ttnpb.Message
				if err := lorawan.UnmarshalMessage(input, &frame); err != nil {
					return fmt.Errorf("failed to decode LoRaWAN frame: %w", err)
				}

				decoded, err := decodeFrame(frame, lorawanConfig)
				if err != nil {
					return err
				}
				return io.Write(os.Stdout, config.OutputFormat, decoded)
			})(cmd, args)
		},
	}
)

func init() {
	lorawanDecodeCmd.Flags().String("lorawan-version", "1.0.2", "LoRaWAN version")
	lorawanDecodeCmd.Flags().String("lorawan-phy-version", "1.0.2-b", "LoRaWAN Regional Parameters version")
	lorawanDecodeCmd.Flags().String("band", "EU_863_870", "LoRaWAN Band ID")
	lorawanDecodeCmd.Flags().String("app-key", "", "LoRaWAN AppKey")
	lorawanDecodeCmd.Flags().String("nwk-key", "", "LoRaWAN NwkKey")
	lorawanDecodeCmd.Flags().String("app-s-key", "", "LoRaWAN AppSKey")
	lorawanDecodeCmd.Flags().String("nwk-s-key", "", "LoRaWAN NwkSKey")
	lorawanDecodeCmd.Flags().String("nwk-s-enc-key", "", "LoRaWAN NwkSEncKey (LoRaWAN 1.1+)")
	lorawanDecodeCmd.Flags().String("f-nwk-s-int-key", "", "LoRaWAN FNwkSIntKey (LoRaWAN 1.1+)")
	lorawanDecodeCmd.Flags().String("s-nwk-s-int-key", "", "LoRaWAN SNwkSIntKey (LoRaWAN 1.1+)")

	lorawanCmd.AddCommand(lorawanDecodeCmd)

	Root.AddCommand(lorawanCmd)
}
