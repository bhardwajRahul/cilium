// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package policymap

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cilium/cilium/pkg/byteorder"
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/policy/trafficdirection"
	"github.com/cilium/cilium/pkg/u8proto"
)

func TestPolicyEntriesDump_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		p    PolicyEntriesDump
		args args
		want bool
	}{
		{
			name: "Same element",
			p: PolicyEntriesDump{
				{
					Key: PolicyKey{
						Identity:         uint32(0),
						DestPortNetwork:  0,
						Nexthdr:          0,
						TrafficDirection: uint8(trafficdirection.Ingress),
					},
				},
			},
			args: args{
				i: 0,
				j: 0,
			},
			want: false,
		},
		{
			name: "Element #0 is less than #1 because identity is smaller",
			p: PolicyEntriesDump{
				{
					Key: PolicyKey{
						Identity:         uint32(0),
						DestPortNetwork:  0,
						Nexthdr:          0,
						TrafficDirection: uint8(trafficdirection.Ingress),
					},
				},
				{
					Key: PolicyKey{
						Identity:         uint32(1),
						DestPortNetwork:  0,
						Nexthdr:          0,
						TrafficDirection: uint8(trafficdirection.Ingress),
					},
				},
			},
			args: args{
				i: 0,
				j: 1,
			},
			want: true,
		},
		{
			name: "Element #0 is less than #1 because TrafficDirection is smaller",
			p: PolicyEntriesDump{
				{
					Key: PolicyKey{
						Identity:         uint32(0),
						DestPortNetwork:  0,
						Nexthdr:          0,
						TrafficDirection: uint8(trafficdirection.Ingress),
					},
				},
				{
					Key: PolicyKey{
						Identity:         uint32(1),
						DestPortNetwork:  0,
						Nexthdr:          0,
						TrafficDirection: uint8(trafficdirection.Egress),
					},
				},
			},
			args: args{
				i: 0,
				j: 1,
			},
			want: true,
		},
		{
			name: "Element #0 is not less than #1 because Identity is bigger",
			p: PolicyEntriesDump{
				{
					Key: PolicyKey{
						Identity:         uint32(1),
						DestPortNetwork:  0,
						Nexthdr:          0,
						TrafficDirection: uint8(trafficdirection.Egress),
					},
				},
				{
					Key: PolicyKey{
						Identity:         uint32(0),
						DestPortNetwork:  0,
						Nexthdr:          0,
						TrafficDirection: uint8(trafficdirection.Egress),
					},
				},
			},
			args: args{
				i: 0,
				j: 1,
			},
			want: false,
		},
		{
			name: "Element #0 is greater than #1 because it is not an allow (denies take precedence)",
			p: PolicyEntriesDump{
				{
					Key: PolicyKey{
						Identity:         uint32(1),
						DestPortNetwork:  0,
						Nexthdr:          0,
						TrafficDirection: uint8(trafficdirection.Egress),
					},
				},
				{
					Key: PolicyKey{
						Identity:         uint32(0),
						DestPortNetwork:  0,
						Nexthdr:          0,
						TrafficDirection: uint8(trafficdirection.Egress),
					},
					PolicyEntry: PolicyEntry{
						Flags: policyFlagDeny,
					},
				},
			},
			args: args{
				i: 0,
				j: 1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		got := tt.p.Less(tt.args.i, tt.args.j)
		require.Equal(t, tt.want, got, "Test Name: %s", tt.name)
	}
}

type opType int

const (
	allow opType = iota
	deny
)

const (
	ingress = trafficdirection.Ingress
	egress  = trafficdirection.Egress
)

func TestPolicyMapWildcarding(t *testing.T) {
	type args struct {
		op               opType
		id               identity.NumericIdentity
		dport            uint16
		dportPrefixLen   uint8
		proto            u8proto.U8proto
		trafficDirection trafficdirection.TrafficDirection
		authType         int
		proxyPort        uint16
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Allow, no wildcarding, no redirection",
			args: args{allow, 42, 80, 16, 6, ingress, 0, 0},
		},
		{
			name: "Allow, no wildcarding, with redirection and auth",
			args: args{allow, 42, 80, 16, 6, ingress, 1, 23767},
		},
		{
			name: "Allow, wildcarded port, no redirection",
			args: args{allow, 42, 0, 0, 6, ingress, 0, 0},
		},
		{
			name: "Allow, wildcarded protocol, no redirection",
			args: args{allow, 42, 0, 0, 0, ingress, 0, 0},
		},
		{
			name: "Deny, no wildcarding, no redirection",
			args: args{deny, 42, 80, 16, 6, ingress, 0, 0},
		},
		{
			name: "Deny, partially wildcarded port, no redirection",
			args: args{deny, 42, 80, 15, 6, ingress, 0, 0},
		},
		{
			name: "Deny, no wildcarding, no redirection",
			args: args{deny, 42, 80, 16, 6, ingress, 0, 0},
		},
		{
			name: "Deny, wildcarded port, no redirection",
			args: args{deny, 42, 0, 0, 6, ingress, 0, 0},
		},
		{
			name: "Deny, wildcarded protocol, no redirection",
			args: args{deny, 42, 0, 0, 0, ingress, 0, 0},
		},
		{
			name: "Allow, wildcarded id, no port wildcarding, no redirection",
			args: args{allow, 0, 80, 16, 6, ingress, 0, 0},
		},
		{
			name: "Allow, wildcarded id, no port wildcarding, with redirection and auth",
			args: args{allow, 0, 80, 16, 6, ingress, 1, 23767},
		},
		{
			name: "Allow, wildcarded id, wildcarded port, no redirection",
			args: args{allow, 0, 0, 0, 6, ingress, 0, 0},
		},
		{
			name: "Allow, wildcarded id, partially wildcarded port, no redirection",
			args: args{allow, 0, 80, 10, 6, ingress, 0, 0},
		},
		{
			name: "Allow, wildcarded id, wildcarded protocol, no redirection",
			args: args{allow, 0, 0, 0, 0, ingress, 0, 0},
		},
		{
			name: "Deny, wildcarded id, no port wildcarding, no redirection",
			args: args{deny, 0, 80, 16, 6, ingress, 0, 0},
		},
		{
			name: "Deny, wildcarded id, no port wildcarding, no redirection",
			args: args{deny, 0, 80, 16, 6, ingress, 0, 0},
		},
		{
			name: "Deny, wildcarded id, wildcarded port, no redirection",
			args: args{deny, 0, 0, 0, 6, ingress, 0, 0},
		},
		{
			name: "Deny, wildcarded id, wildcarded protocol, no redirection",
			args: args{deny, 0, 0, 0, 0, ingress, 0, 0},
		},
	}
	for _, tt := range tests {
		// Validate test data
		if tt.args.proto == 0 {
			require.Equal(t, uint16(0), tt.args.dport, "Test: %s data error: dport must be wildcarded when protocol is wildcarded", tt.name)
			require.Equal(t, uint8(0), tt.args.dportPrefixLen, "Test: %s data error: dport prefix length must be 0 when protocol is wildcarded", tt.name)
		}
		if tt.args.dport == 0 {
			require.Equal(t, uint8(0), tt.args.dportPrefixLen, "Test: %s data error: dport prefix length must be 0 when dport is wildcarded", tt.name)
			require.Equal(t, uint16(0), tt.args.proxyPort, "Test: %s data error: proxyPort must be zero when dport is wildcarded", tt.name)
		}
		if tt.args.op == deny {
			require.Equal(t, uint16(0), tt.args.proxyPort, "Test: %s data error: proxyPort must be zero with a deny key", tt.name)
			require.Equal(t, 0, tt.args.authType, "Test: %s data error: authType must be zero with a deny key", tt.name)
		}

		// Get key
		key := NewKey(tt.args.trafficDirection, tt.args.id, tt.args.proto, tt.args.dport, tt.args.dportPrefixLen)

		// Compure entry & validate key and entry
		var entry PolicyEntry
		switch tt.args.op {
		case allow:
			entry = newAllowEntry(key, uint8(tt.args.authType), uint16(tt.args.proxyPort))

			require.Equal(t, policyEntryFlags(0), entry.Flags&policyFlagDeny)
			require.Equal(t, uint8(tt.args.authType), entry.AuthType)
			require.Equal(t, uint16(tt.args.proxyPort), byteorder.NetworkToHost16(entry.ProxyPortNetwork))
		case deny:
			entry = newDenyEntry(key)

			require.Equal(t, policyFlagDeny, entry.Flags&policyFlagDeny)
			require.Equal(t, uint8(0), entry.AuthType)
			require.Equal(t, uint16(0), entry.ProxyPortNetwork)
		}

		require.Equal(t, uint32(tt.args.id), key.Identity)
		require.Equal(t, uint8(tt.args.proto), key.Nexthdr)

		// key and entry need to agree on the prefix length
		require.Equal(t, StaticPrefixBits+uint32(entry.LPMPrefixLength), key.Prefixlen)

		if key.Nexthdr == 0 {
			require.Equal(t, uint16(0), key.DestPortNetwork)
			require.Equal(t, StaticPrefixBits, key.Prefixlen)
			require.Equal(t, uint8(0), entry.LPMPrefixLength)
		} else {
			if key.DestPortNetwork == 0 {
				require.Equal(t, StaticPrefixBits+NexthdrBits, key.Prefixlen)
				require.Equal(t, uint8(NexthdrBits), entry.LPMPrefixLength)
			} else {
				require.Equal(t, uint16(tt.args.dport), byteorder.NetworkToHost16(key.DestPortNetwork))
				require.Equal(t, StaticPrefixBits+NexthdrBits+uint32(tt.args.dportPrefixLen), key.Prefixlen)
				require.Equal(t, uint8(NexthdrBits)+tt.args.dportPrefixLen, entry.LPMPrefixLength)
			}
		}
	}
}

func TestPortProtoString(t *testing.T) {
	type args struct {
		key *PolicyKey
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Allow all",
			args: args{
				&PolicyKey{
					Prefixlen:        StaticPrefixBits,
					Identity:         0,
					TrafficDirection: trafficdirection.Ingress.Uint8(),
					Nexthdr:          0,
					DestPortNetwork:  0,
				},
			},
			want: "ANY",
		},
		{
			name: "Fully specified port",
			args: args{
				&PolicyKey{
					Prefixlen:        StaticPrefixBits + NexthdrBits + DestPortBits,
					Identity:         0,
					TrafficDirection: trafficdirection.Ingress.Uint8(),
					Nexthdr:          0,
					DestPortNetwork:  byteorder.HostToNetwork16(8080),
				},
			},
			want: "8080/ANY",
		},
		{
			name: "Fully specified port and proto",
			args: args{
				&PolicyKey{
					Prefixlen:        StaticPrefixBits + NexthdrBits + DestPortBits,
					Identity:         0,
					TrafficDirection: trafficdirection.Ingress.Uint8(),
					Nexthdr:          6,
					DestPortNetwork:  byteorder.HostToNetwork16(8080),
				},
			},
			want: "8080/TCP",
		},
		{
			name: "Match TCP / wildcarded port",
			args: args{
				&PolicyKey{
					Prefixlen:        StaticPrefixBits + NexthdrBits,
					Identity:         0,
					TrafficDirection: trafficdirection.Ingress.Uint8(),
					Nexthdr:          6,
					DestPortNetwork:  0,
				},
			},
			want: "TCP",
		},
		{
			name: "Wildard proto / match upper 8 bits of port",
			args: args{
				&PolicyKey{
					Prefixlen:        StaticPrefixBits + NexthdrBits + DestPortBits/2,
					Identity:         0,
					TrafficDirection: trafficdirection.Ingress.Uint8(),
					Nexthdr:          0,
					DestPortNetwork:  byteorder.HostToNetwork16(0x0100), // 256 and all ports with 256 as a prefix
				},
			},
			want: "256-511/ANY",
		},
	}
	for _, tt := range tests {
		got := tt.args.key.PortProtoString()
		require.Equal(t, tt.want, got, "Test Name: %s", tt.name)
	}
}
