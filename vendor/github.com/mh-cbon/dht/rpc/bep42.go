package rpc

import (
	"fmt"
	"net"

	"github.com/mh-cbon/dht/kmsg"
	"github.com/mh-cbon/dht/security"
	"github.com/mh-cbon/dht/socket"
)

// SecuredResponseOnly blank responded tokens when their node id is invalid.
// bep42: Once enforced, responses to get_peers/get requests whose node ID does not match its external IP
// should be considered to not contain a token
// and thus not be eligible as storage target.
func SecuredResponseOnly(remote *net.UDPAddr, f func(kmsg.Msg)) func(kmsg.Msg) {
	return func(res kmsg.Msg) {
		if res.R != nil && security.NodeIDSecure(res.R.ID, remote.IP) == false {
			res.R.Token = ""
		}
		f(res)
	}
}

// SecuredQueryOnly checks that incomig queries are secured with bep42.
// bep42: Nodes that enforce the node-ID will respond with an error message ("y": "e", "e": { ... }),
// whereas a node that supports this extension
// but without enforcing it will respond with a normal reply ("y": "r", "r": { ... }).
func SecuredQueryOnly(k *KRPC, f socket.QueryHandler) socket.QueryHandler {
	return func(msg kmsg.Msg, remote *net.UDPAddr) error {
		if msg.A == nil {
			return fmt.Errorf("Invalid get_peers packet: mising Arguments")
		}
		q := msg.Q
		if (q == kmsg.QGet || q == kmsg.QGetPeers) && security.NodeIDSecure(msg.A.ID, remote.IP) == false {
			//todo: check about that with peerstat store -> // k.addBadNode(remote)
			k.lookupTableForPeers.RemoveNode(remote)
			k.lookupTableForStores.RemoveNode(remote)
			return k.Error(remote, msg.T, kmsg.ErrorInsecureNodeID)
		}
		return f(msg, remote)
	}
}
