---
title: "Cisco Wireless Gateway for LoRaWAN"
description: ""
weight: 1
---

The **Cisco Wireless Gateway for LoRaWAN** technical specifications can be found in [Cisco's official documentation](https://www.cisco.com/c/en/us/products/routers/wireless-gateway-lorawan/). This page guides you to connecting the gateway to {{% tts %}}.

![Cisco LoRaWAN Gateway](cisco.jpg)

## Prerequisites

1. User account on {{% tts %}} with rights to create gateways.
2. Cisco Wireless Gateway for LoRaWAN with latest firmware (version `2.0.32` or higher), connected to the internet (or your local network) via ethernet.
3. [Console cable from USB to RJ45](https://www.cablesandkits.com/accessories/console-cables/usb-rj45-6ft/pro-9900/).

## Registration

Create a gateway by following the instructions for the [Console]({{< ref "/getting-started/console#create-gateway" >}}) or the [CLI]({{< ref "/getting-started/cli#create-gateway" >}}).

The **EUI** is derived from the **MAC_ADDRESS** that can be found on the back panel of the gateway. To get the EUI from the MAC_ADDRESS insert `FFFE` **after the first 6 characters** to make it a 64bit EUI.

> Note: If your **MAC_ADDRESS** is `5B:A0:CB:80:04:2B` then the **EUI** is `5B A0 CB FF FE 80 04 2B`.

The **Gateway Server Address** is the same as what you use instead of `thethings.example.com` in the [Getting Started guide]({{< ref "/getting-started" >}}).

## Configuration

Plug the RJ45 end of the cable in the **Console** port at the side of the gateway, and the USB port to your computer.

If you are using MacOS or Linux, connect to the Gateway by opening a terminal and a executing the following commands:

```bash
$ ls /dev/tty.usb*
```

> Note: This displays the list of available USB serial devices.

Once you have found the one matching the Cisco console, connect using the following command:

```bash
$ screen /dev/tty.usbserial-AO001X6M 115200
```

Use PuTTy if you are using Windows.

You are now in the gateway's shell, called `standalone mode`.

### System setup

First you need to enable the privileged mode.

```
Gateway> enable
```

#### Network

To configure your Cisco Gateway to your network, type the following commands:

```
Gateway# configure terminal 
Gateway(config)# interface FastEthernet 0/1
```

If your local network has a **DHCP server** attributing IPs:

```
Gateway(config-if)# ip address dhcp
```

Otherwise, if you know the **static IP address** of your gateway:

```
Gateway(config-if)# ip address <ip-address> <subnet-mask>
```

Next, type the following to save the network configuration of your gateway:

```
Gateway(config-if)# description Ethernet
Gateway(config)# exit
Gateway# exit
Gateway# copy running-config startup-config
```

You can test your Internet configuration with the `ping` command, for example ping Google's DNS server

```
Gateway# ping ip 8.8.8.8
```

> Note: To see more information about the gateway's IP and the network, you can use 
> `show interfaces FastEthernet 0/1`
> `show ip interfaces FastEthernet 0/1` or
> `show ip route`.

#### Date and Time

To configure your system's date and time, you can use **ntp**:

```
Gateway# configure terminal
Gateway(config)# ntp server address <NTP server address>
Gateway(config)# exit
```

or

```
Gateway# configure terminal
Gateway(config)# ntp server ip <NTP server IP>
Gateway(config)# exit
```

If you don't have production-grade ntp servers available, you can use [pool.ntp.org](http://www.pool.ntp.org/en/use.html)'s servers.

#### FPGA

If you needed to update your gateway firmware previously, your FPGA will need ~20 minutes to update once the new firmware is installed. The packet forwarder will not work until then, so we recommend at this point waiting until the FPGA is upgraded.To show the status of the FPGA, you can use the following command:

```
Gateway# show inventory
```

When the **FPGAStatus** line indicates **Ready**, this means you can go forward with this guide.

#### GPS

If you have a GPS connected to your Cisco gateway, enable it with the following commands:

```
Gateway# configure terminal
Gateway(config)# gps ubx enable
Gateway(config)# exit
```

> Note: This command may return the message `packet-forwarder firmware is not installed`, this message can be ignored.

#### Enable Radio

As a final step before setting up the packet forwarder software, we're going to **enable the radio**. You can see radio information with the `show radio` command:

```
Gateway# show radio 
ORA_SN: FOC21028R8S
ORA_PN: 95.1602T01
ORA_SKU: 915
ORA_CALC: <NA,NA,NA,50,31,106,97,88,80,71,63,53,44,34,25,16-NA,NA,NA,54,36,109,100,91,83,74,66,57,48,39,30,21>
AL_TEMP_CELSIUS: 31
AL_TEMP_CODE_AD9361: 87
SSI_OFFSET: -204.00,-204.40
ORA_REVISION_NUM: C0
SSI_OFFSET_AUS: -203.00,-204.00

radio status:
on
```

If the radio is off, enable it with

```
Gateway# configure terminal 
Gateway(config)# no radio off
Gateway(config)# exit
```

> Note: The `show radio` command also shows you more information about the LoRa concentrator powering the gateway. For example, **LORA_SKU** indicates the base frequency of the concentrator.

#### Enable Authentication

To prevent unauthorized access to the gateway, you'll want to set up user authentication. The Cisco gateway has a **secret** system, that requires users to enter a secret to access privileged commands.

To enable this secret system, you can use the following commands:

+ `Gateway# configure terminal` to enter global configuration mode.
+ To set the secret, you can use different commands:
  `Gateway(config)# enable secret <secret>` to enter in plaintext the secret you wish to set, instead of `<secret>`. *Note*: Special characters cannot be used in plain secrets.
  `Gateway(config)# enable secret 5 <secret>` to enter the secret **md5-encrypted**.
  `Gateway(config)# enable secret 8 <secret>` to enter the secret **SHA512-encrypted**.
+ `Gateway(config)# exit` to exit global configuration mode.
+ `Gateway#copy running-config startup-config` to save the configuration.

#### Verification

Before we install the packet forwarder, let's run verification to ensure that the gateway is ready.

+ Type `show radio` to verify that the **radio is enabled**. The result should indicate **radio status: on**.
+ Type `show inventory` to verify that the **FPGAStatus is Ready**.
+ Type `show gps status` to verify that the **GPS is correctly connected**. You can get additional GPS metadata by typing `show gps info`.
+ Verify that the **network connection is working**. You can test this by pinging common ping servers with `ping ip <IP>`, if your local network does not block ping commands. For example, you can ping Google's servers with `ping ip 8.8.8.8`.

If some of those checks fail, go back to the appropriate section earlier in order to fix it.

Then save the configuration by executing

```
Gateway# copy running-config startup-config
```

### Packet Forwarder Configuration

> ⚠️ Keep in mind that the pre-installed packet forwarder is not supported by Cisco for production purposes.

To run the packet forwarder, we'll make use of the **container** that is running on the gateway at all times.

```
Gateway# request shell container-console
```

You will be requested to enter the System Password. By default this is **admin**.

Create the directory to store the Packet Forwarder configuration:

```bash
bash-3.2# mkdir /etc/pktfwd
```

Copy the packet forwarder to **/etc/pktfwd**:

```bash
bash-3.2# cp /tools/pkt_forwarder /etc/pktfwd/pkt_forwarder
```

**Retrieve configuration from the Gateway Configuration Server**
The Gateway Configuration Server can be used to retrieve a proper `global_conf.json` configuration file for your gateway.

You will need a Gateway API key with the `View gateway information` right enabled. Instructions can be found in the relevant sections of the [Console]({{< ref "/getting-started/console#create-gateway" >}}) or the [CLI]({{< ref "/getting-started/cli#create-gateway" >}}) getting started guides.


Make sure to replace `thethings.example.com`, `GATEWAY_ID` and `GTW_API_KEY` with your **server address**, **Gateway ID** and **Gateway API Key ID** which can be found using your Console or CLI.

```
bash-3.2# curl -k XGET "https://thethings.example.com/api/v3/gcs/gateways/GATEWAY_ID/semtechudp/global_conf.json" -H "Authorization: Bearer GTW_API_KEY" > ~/config.json
```

Copy the configuration template as **config.json** to **/etc/pktfwd** 

```bash
bash-3.2# cp config.json /etc/pktfwd/config.json
```

You can now test the packet forwarder by executing:

```bash
bash-3.2# /etc/pktfwd/pkt_forwarder -c /etc/pktfwd/config.json -g/dev/ttyS1
```

Your gateway will connect to {{% tts %}} after a couple of minutes.


Now that we know the packet forwarder is running, let's make it run automatically. Use this command:

```bash
bash-3.2# vi /etc/init.d/S60pkt_forwarder
```

>Note: Press the `i` key on your keyboard to start insert mode. Once finished editing, press `ESC` and enter `:wq` to write the file and quit.

Then copy paste the code below.

> Note: Replace `things.example.com` with the name of your network after `nslookup`.

```bash
SCRIPT_DIR=/etc/pktfwd
SCRIPT=$SCRIPT_DIR/pkt_forwarder
CONFIG=$SCRIPT_DIR/config.json

PIDFILE=/var/run/pkt_forwarder.pid
LOGFILE=/var/log/pkt_forwarder.log

export NETWORKIP=$(nslookup things.example.com | grep -E -o "([0-9]{1,3}[\.]){3}[0-9]{1,3}" | tail -1)
sed -i 's/[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}/'$NETWORKIP'/g' "$CONFIG"

start() {
  echo "Starting pkt_forwarder"
  cd $SCRIPT_DIR
  start-stop-daemon \
        --start \
        --make-pidfile \
        --pidfile "$PIFDILE" \
        --background \
        --startas /bin/bash -- -c "exec $SCRIPT -- -c $CONFIG -g /dev/ttyS1 >> $LOGFILE 2>&1"
  echo $?
}

stop() {
  echo "Stopping pkt_forwarder"
  start-stop-daemon \
        --stop \
        --oknodo \
        --quiet \
        --pidfile "$PIDFILE"
}

restart() {
  stop
  sleep 1
  start
}

case "$1" in
  start)
    start
    ;;
  stop)
    stop
    ;;
  restart|reload)
    restart
    ;;
  *)
    echo "Usage: $0 {start|stop|restart}"
    exit 1
esac

exit $?

```

Then make the init script executable:

```bash
bash-3.2# chmod +x /etc/init.d/S60pkt_forwarder
```

To enable it immediately, execute 

```bash
bash-3.2# /etc/init.d/S60pkt_forwarder start
```

You can now reboot the gateway, it can take up to 4 minutes.

## Troubleshooting

If the gateway does not connect to the {{% tts %}} after a few minutes, you can check the **log file** to see if the packet forwarder started properly.

```bash
bash-3.2# tail -100 var/log/pkt_forwarder.log
```

> Note: GPS warnings may appear, this means the packet forwarder started.

If the radio failed to start, disconnect and reconnect the power supply to power-cycle the gateway.

For further information and troubleshooting, have a look at [Cisco's Configuration Guide](https://www.cisco.com/c/en/us/td/docs/routers/interface-module-lorawan/software/configuration/guide/b_lora_scg.pdf).
