// Package kmsg provides the data structures required to implement the dht/bittorrent protocol.
package kmsg

import (
	"fmt"

	"github.com/anacrolix/torrent/util"
)

// Msg represents messages that nodes in the network send to each other as specified by the protocol.
// They are also referred to as the KRPC messages.
// There are three types of messages: QUERY, RESPONSE, ERROR
// The message is a dictonary that is then
// "bencoded" (serialization & compression format adopted by the BitTorrent)
// and sent via the UDP connection to peers.
//
// A KRPC message is a single dictionary with two keys common to every message and additional keys depending on the type of message.
// Every message has a key "t" with a string value representing a transaction ID.
// This transaction ID is generated by the querying node and is echoed in the response, so responses
// may be correlated with multiple queries to the same node. The transaction ID should be encoded as a short string of binary numbers, typically 2 characters are enough as they cover 2^16 outstanding queries. The other key contained in every KRPC message is "y" with a single character value describing the type of message. The value of the "y" key is one of "q" for query, "r" for response, or "e" for error.
// 3 message types:  QUERY, RESPONSE, ERROR
type Msg struct {
	Q  string           `bencode:"q,omitempty"`  // Query method (one of 4: "ping", "find_node", "get_peers", "announce_peer")
	A  *MsgArgs         `bencode:"a,omitempty"`  // named arguments sent with a query
	T  string           `bencode:"t"`            // required: transaction ID
	Y  string           `bencode:"y"`            // required: type of the message: q for QUERY, r for RESPONSE, e for ERROR
	R  *Return          `bencode:"r,omitempty"`  // RESPONSE type only
	E  *Error           `bencode:"e,omitempty"`  // ERROR type only
	IP util.CompactPeer `bencode:"ip,omitempty"` // bep42: outgoing query: requestor ip, incoming query: our ip accodring to the remote
	RO int              `bencode:"ro,omitempty"` // bep43: ro is a read only top level field
}

// MsgArgs are the query arguments.
type MsgArgs struct {
	ID          string `bencode:"id"`           // ID of the querying Node
	InfoHash    string `bencode:"info_hash"`    // InfoHash of the torrent
	Target      string `bencode:"target"`       // ID of the node sought
	Token       string `bencode:"token"`        // Token received from an earlier get_peers query
	Port        int    `bencode:"port"`         // Senders torrent port
	ImpliedPort int    `bencode:"implied_port"` // Use senders apparent DHT port

	V    string `bencode:"v,omitempty"`    // Data stored in a put message (encoded size < 1000)
	Seq  int    `bencode:"seq,omitempty"`  // Seq of a mutable msg
	Cas  int    `bencode:"cas,omitempty"`  // CAS value of the message mutation
	K    []byte `bencode:"k,omitempty"`    // ed25519 public key (32 bytes string) of a mutable msg
	Salt string `bencode:"salt,omitempty"` // <optional salt to be appended to "k" when hashing (string) a mutable msg
	Sign []byte `bencode:"sig,omitempty"`  // ed25519 signature (64 bytes string)
}

// Return are query responses.
type Return struct {
	ID     string              `bencode:"id"`               // ID of the querying node
	Nodes  CompactIPv4NodeInfo `bencode:"nodes,omitempty"`  // K closest nodes to the requested target
	Token  string              `bencode:"token,omitempty"`  // Token for future announce_peer
	Values []util.CompactPeer  `bencode:"values,omitempty"` // Torrent peers

	V    string `bencode:"v,omitempty"`   // Data stored in a put message (encoded size < 1000)
	Seq  int    `bencode:"seq,omitempty"` // Seq of a mutable msg
	K    []byte `bencode:"k,omitempty"`   // ed25519 public key (32 bytes string) of a mutable msg
	Sign []byte `bencode:"sig,omitempty"` // ed25519 signature (64 bytes string)
}

var _ fmt.Stringer = Msg{}

func (m Msg) String() string {
	return fmt.Sprintf("%#v", m)
}

// SenderID The node ID of the source of this Msg.
func (m Msg) SenderID() string {
	switch m.Y {
	case "q":
		if m.A != nil {
			return m.A.ID
		}
	case "r":
		if m.R != nil {
			return m.R.ID
		}
	}
	return ""
}

func (m Msg) Error() *Error {
	if m.Y != "e" {
		return nil
	}
	return m.E
}