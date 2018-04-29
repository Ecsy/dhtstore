package main

import (
	"encoding/hex"
	"flag"
	"log"
	"net"
	"os"

	"github.com/ecsy/thesis/src/network"
	"github.com/mh-cbon/dht/dht"
	"github.com/mh-cbon/dht/ed25519"
	"github.com/mh-cbon/dht/security"
	"github.com/mh-cbon/dht/socket"
	cryptoed25519 "golang.org/x/crypto/ed25519"
)

const (
	keyName = "dht"
)

var (
	// Program version information.
	version string
	// Program build date.
	buildDate string
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile | log.LUTC)
	log.Printf("program: %s, version: %s, build date: %s\n", os.Args[0], version, buildDate)

	var (
		action = flag.String("action", "sign", "program action. valid options are: sign, get, put")
		value  = flag.String("value", "", "")
		seq    = flag.Int("seq", 1, "")
		salt   = flag.String("salt", "", "")
		target = flag.String("target", "", "")
	)

	flag.Parse()

	privateKey, publicKey, err := getKeys()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Public Key: %x\n", publicKey)
	log.Printf("Private Key: %x...%x\n", privateKey[:4], privateKey[60:])
	_, _ = privateKey, publicKey

	var readyFn func(public *dht.DHT) error

	switch *action {
	case "sign":
		m, err := network.MutableTarget(privateKey, *value, *seq, *salt)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("hash: %s\n", m.Target)
		log.Printf("signature: %x\n", m.Sign)
		return

	case "get":
		if len(*target) != 40 {
			log.Fatal("please specifiy a valid target")
		}
		readyFn = get(publicKey, *target, *seq, *salt)

	case "put":
		readyFn = put(privateKey, *value, *seq, *salt)

	default:
		log.Fatal("Invalid program action")
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Panic(err)
	}

	var pip *net.IP
	i := security.GenerateSecureNodeID(hostname, nil, pip)
	socket := socket.NewConcurrent(24)
	//socket.AddLogger(logger.Text(log.Printf))

	opts := make([]dht.Opt, 0)
	opts = append(opts, dht.Opts.WithRPCSocket(socket))
	opts = append(opts, dht.Opts.WithAddr(""))
	opts = append(opts, dht.Opts.ID(string(i)))
	opts = append(opts, dht.Opts.WithConcurrency(8))
	opts = append(opts, dht.Opts.WithK(20))

	node := dht.New(opts...)
	if err := node.ListenAndServe(dht.StdQueryHandler(node), readyFn); err != nil {
		log.Fatal(err)
	}
}

func get(publicKey cryptoed25519.PublicKey, target string, seq int, salt string) func(*dht.DHT) error {
	return func(public *dht.DHT) error {
		// DHT bootstrap
		n := network.NewDHT(public, log.New(os.Stderr, "", log.Flags()))
		_, err := n.Bootstrap("bootstrap.json")
		if err != nil {
			return err
		}

		for {
			val, err := n.Get(target, publicKey, seq, salt)
			if err != nil {
				if err == network.ErrValueNotFound {
					log.Println(err)
					continue
				}
				log.Fatal(err)
			}
			log.Printf("seq: %d, val: %s\n", seq, val)
		}

		return nil
	}
}

func put(privateKey ed25519.PrivateKey, value string, seq int, salt string) func(*dht.DHT) error {
	return func(public *dht.DHT) error {
		// DHT bootstrap
		n := network.NewDHT(public, log.New(os.Stderr, "", log.Flags()))
		_, err := n.Bootstrap("bootstrap.json")
		if err != nil {
			return err
		}

		m, err := network.MutableTarget(privateKey, value, seq, salt)
		if err != nil {
			return err
		}
		log.Printf("put target hash: %v\n", m.Target)
		err = n.Put(m)
		if err != nil {
			return err
		}

		log.Println("put done")
		return nil
	}
}

func getKeys() (ed25519.PrivateKey, cryptoed25519.PublicKey, error) {
	b, _ := hex.DecodeString("5b3246f7fdfd21518648e4247af5e506b94660c4a93f5a4ca95dfaaf6e73955c36b6958c9b335c5852ecf06499a515c186fbbefb12cf93a94c5dd329b27ac743")
	private := ed25519.PrivateKey(b)
	b, _ = hex.DecodeString("c5b985b2e22eb9c13a0c9dd788c2b1d6d539d1c096c43eed3c5df34b775d4668")
	public := cryptoed25519.PublicKey(b)
	var err error

	//// Private and public key, generate if not found.
	//private, public, err := ed25519.PvkFromDir(".", keyName)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//// Fix ed25519.PvkFromDir returning longer private key than expected.
	//private = private[:64]
	return private, public, err
}
