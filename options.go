package ddhcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

type OptionCode byte

const (
	Pad                          OptionCode = 0
	OptionSubnetMask             OptionCode = 1
	OptionTimeOffset             OptionCode = 2
	OptionRouter                 OptionCode = 3
	OptionTimeServer             OptionCode = 4
	OptionNameServer             OptionCode = 5
	OptionDomainNameServer       OptionCode = 6
	OptionLogServer              OptionCode = 7
	OptionCookieServer           OptionCode = 8
	OptionLPRServer              OptionCode = 9
	OptionImpressServer          OptionCode = 10
	OptionResourceLocationServer OptionCode = 11
	OptionHostName               OptionCode = 12
	OptionBootFileSize           OptionCode = 13
	OptionMeritDumpFile          OptionCode = 14
	OptionDomainName             OptionCode = 15
	OptionSwapServer             OptionCode = 16
	OptionRootPath               OptionCode = 17
	OptionExtensionsPath         OptionCode = 18
	End                          OptionCode = 255
	// IP Layer Parameters per Host
	OptionIPForwardingEnableDisable          OptionCode = 19
	OptionNonLocalSourceRoutingEnableDisable OptionCode = 20
	OptionPolicyFilter                       OptionCode = 21
	OptionMaximumDatagramReassemblySize      OptionCode = 22
	OptionDefaultIPTimeToLive                OptionCode = 23
	OptionPathMTUAgingTimeout                OptionCode = 24
	OptionPathMTUPlateauTable                OptionCode = 25
	// IP Layer Parameters per Interface
	OptionInterfaceMTU              OptionCode = 26
	OptionAllSubnetsAreLocal        OptionCode = 27
	OptionBroadcastAddress          OptionCode = 28
	OptionPerformMaskDiscovery      OptionCode = 29
	OptionMaskSupplier              OptionCode = 30
	OptionPerformRouterDiscovery    OptionCode = 31
	OptionRouterSolicitationAddress OptionCode = 32
	OptionStaticRoute               OptionCode = 33
	// Link Layer Parameters per Interface
	OptionLinkLayerParametersPerInterface OptionCode = 34
	OptionTrailerEncapsulation            OptionCode = 34
	OptionARPCacheTimeout                 OptionCode = 35
	OptionEthernetEncapsulation           OptionCode = 36
	// TCP Parameters
	OptionTCPDefaultTTL        OptionCode = 37
	OptionTCPKeepaliveInterval OptionCode = 38
	OptionTCPKeepaliveGarbage  OptionCode = 39
	// Application and Service Parameters
	OptionNetworkInformationServiceDomain            OptionCode = 40
	OptionNetworkInformationServers                  OptionCode = 41
	OptionNetworkTimeProtocolServers                 OptionCode = 42
	OptionVendorSpecificInformation                  OptionCode = 43
	OptionNetBIOSOverTCPIPNameServer                 OptionCode = 44
	OptionNetBIOSOverTCPIPDatagramDistributionServer OptionCode = 45
	OptionNetBIOSOverTCPIPNodeType                   OptionCode = 46
	OptionNetBIOSOverTCPIPScope                      OptionCode = 47
	OptionXWindowSystemFontServer                    OptionCode = 48
	OptionXWindowSystemDisplayManager                OptionCode = 49
	OptionNetworkInformationServicePlusDomain        OptionCode = 64
	OptionNetworkInformationServicePlusServers       OptionCode = 65
	OptionMobileIPHomeAgent                          OptionCode = 68
	OptionSimpleMailTransportProtocol                OptionCode = 69
	OptionPostOfficeProtocolServer                   OptionCode = 70
	OptionNetworkNewsTransportProtocol               OptionCode = 71
	OptionDefaultWorldWideWebServer                  OptionCode = 72
	OptionDefaultFingerServer                        OptionCode = 73
	OptionDefaultInternetRelayChatServer             OptionCode = 74
	OptionStreetTalkServer                           OptionCode = 75
	OptionStreetTalkDirectoryAssistance              OptionCode = 76

	OptionRelayAgentInformation OptionCode = 82
	// DHCP Extensions
	OptionRequestedIPAddress     OptionCode = 50
	OptionIPAddressLeaseTime     OptionCode = 51
	OptionOverload               OptionCode = 52
	OptionDHCPMessageType        OptionCode = 53
	OptionServerIdentifier       OptionCode = 54
	OptionParameterRequestList   OptionCode = 55
	OptionMessage                OptionCode = 56
	OptionMaximumDHCPMessageSize OptionCode = 57
	OptionRenewalTimeValue       OptionCode = 58
	OptionRebindingTimeValue     OptionCode = 59
	OptionVendorClassIdentifier  OptionCode = 60
	OptionClientIdentifier       OptionCode = 61

	OptionTFTPServerName OptionCode = 66
	OptionBootFileName   OptionCode = 67

	OptionTZPOSIXString    OptionCode = 100
	OptionTZDatabaseString OptionCode = 101

	OptionClasslessRouteFormat OptionCode = 121
)

type Options map[OptionCode][]byte

func NewOptionsFromData(buf []byte) (Options, error) {
	options := make(Options, 10)

	for i := 0; i < len(buf); {
		opt := OptionCode(buf[i])
		if opt == End {
			break
		}
		if opt == Pad {
			i++
			continue
		}

		if len(buf) < i+1 {
			return options, errors.New("buffer invalid size, there is no option size")
		}
		size := int(buf[i+1])
		optSize := 2 + size
		if len(buf)-i < optSize {
			return options, errors.New("buffer invalid size, buffer does not contain all data")
		}
		options[opt] = buf[i+2 : i+optSize]
		i += optSize
	}

	return options, nil
}

//func NewOptions(msgType MsgType, serverId net.IP, leaseDuration time.Duration) Options {
//	return OptionCode{OptionDHCPMessageType: []byte{msgType},
//		OptionServerIdentifier:   serverId,
//		OptionIPAddressLeaseTime: duration}
//}

func (o Options) MsgType() (MsgType, error) {
	data, exists := o[OptionDHCPMessageType]
	if !exists {
		return 0, errors.New("options doesnt exists")
	}
	if len(data) != 1 {
		return 0, errors.New("data incorrect size - OptionDHCPMessageType")
	}
	return NewMsgType(data[0])
}

func (o Options) Read(p []byte) (n int, err error) {
	readers := make([]io.Reader, 0, len(o)*2+1)
	for k, v := range o {
		md := []byte{byte(k), byte(len(v))}
		readers = append(readers, bytes.NewReader(md), bytes.NewReader(v))
	}

	readers = append(readers, bytes.NewReader([]byte{byte(End)}))
	reader := io.MultiReader(readers...)
	return reader.Read(p)
}

func (o Options) Bytes() ([]byte, error) {
	optionsSize := 1 // contain at least End entry
	for _, v := range o {
		optionsSize += 2 + len(v)
	}

	buf := make([]byte, optionsSize)
	n, err := o.Read(buf)
	if err != nil {
		return buf, err
	}
	if n != optionsSize {
		return buf, errors.New(fmt.Sprintf("unexpected size was read %d", n))
	}
	return buf, nil
}

func (o Options) SetLeaseDuration(d time.Duration) {
	rawDuration := make([]byte, 4)
	binary.BigEndian.PutUint32(rawDuration, uint32(d/time.Second))
	o[OptionIPAddressLeaseTime] = rawDuration
}

func (o Options) SetMsgType(msgType MsgType) {
	o[OptionDHCPMessageType] = []byte{byte(msgType)}
}
