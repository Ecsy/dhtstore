package network

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/anacrolix/torrent/util"
	"github.com/mh-cbon/dht/bootstrap"
	"github.com/mh-cbon/dht/dht"
	"github.com/mh-cbon/dht/ed25519"
	"github.com/mh-cbon/dht/security"
	"github.com/pkg/errors"
)

var (
	// publicDHTNodes lists public addresses to bootstrap.
	publicDHTNodes = []string{
		"router.bittorrent.com:6881",
		"router.utorrent.com:6881",
	}

	ErrValueNotFound = errors.New("value not found")
)

// DHT network helper functions.
type DHT struct {
	public     *dht.DHT
	log        *log.Logger
	storeCache map[string][]*net.UDPAddr
	storeMx    *sync.RWMutex
}

// NewDHT DHT instance.
func NewDHT(public *dht.DHT, log *log.Logger) *DHT {
	return &DHT{
		public:     public,
		log:        log,
		storeCache: make(map[string][]*net.UDPAddr, 0),
		storeMx:    &sync.RWMutex{},
	}
}

// Bootstrap creates an initial DHT network table containing IP addresses and ports.
func (d *DHT) Bootstrap(filename string) (bNodes []string, err error) {
	d.log.Println("DHT bootstrap...")
	var publicIP *util.CompactPeer
	var selfID = d.public.GetID()
	bNodes = publicDHTNodes

	if filename != "" {
		data, err := bootstrap.Get(filename)
		if err == nil && len(data.Nodes) > 0 {
			bNodes = data.Nodes
			publicIP = data.OldIP
			d.log.Println("loaded bootstrap nodes:", len(bNodes))
		}
	}

	recommendedIP := publicIP
	if rIP, bootErr := d.public.Bootstrap(selfID, publicIP, bNodes); bootErr != nil && len(bNodes) > 0 {
		err = bootErr
		return
	} else if rIP != nil {
		recommendedIP = rIP
		var hostname string
		hostname, err = os.Hostname()
		if err != nil {
			return
		}
		selfID = security.GenerateSecureNodeID(hostname, d.public.GetAddr(), &rIP.IP)
		d.log.Printf("after bootstrap a new recommended ip was provided %v\n", rIP)
		if rIP2, bootErr := d.public.Bootstrap(selfID, rIP, bNodes); bootErr != nil {
			err = bootErr
			return
		} else if rIP2 != nil {
			d.log.Printf("after bootstrap a new recommended ip was provided %v\n", rIP2)
			err = fmt.Errorf("stopping now as something is going wrong")
			return
		}
	}
	if len(bNodes) > 0 {
		bNodes := d.public.BootstrapExport()
		d.log.Println("bootstrap nodes:", len(bNodes))
		if filename != "" {
			_ = bootstrap.Save(filename, publicIP, bNodes)
		}
	}
	d.log.Printf("id after bootstrap %x\n", d.public.ID())
	if recommendedIP != nil {
		d.log.Printf("public IP after bootstrap %v:%v\n", recommendedIP.IP, recommendedIP.Port)
	}
	return
}

// ClosesStoresForHash returns closest peers for the given hash in the DHT network.
// These addresses can be used to store or retrieve mutable values from the DHT network.
func (d *DHT) ClosestStoresForHash(hash string) (addr []*net.UDPAddr, err error) {
	//log.Println("LookupStores targetHash:", targetHash)
	err = d.public.LookupStores(hash, nil)
	if err != nil {
		return
	}
	//log.Println("ClosestStores targetHash:", targetHash)
	contacts, err := d.public.ClosestStores(hash, 64)
	if err != nil {
		return
	}
	//log.Println("ClosestStores length:", len(contacts), contacts[0].GetAddr().String())

	for _, c := range contacts {
		//log.Printf("c: %s\n", c.GetAddr().String())
		addr = append(addr, c.GetAddr())
	}
	return
}

// closestStoresForHash cached version of ClosestStoresForHash.
func (d *DHT) closestStoresForHash(hash string) ([]*net.UDPAddr, error) {
	d.storeMx.RLock()
	if addr, ok := d.storeCache[hash]; ok {
		d.storeMx.RUnlock()
		return addr, nil
	}
	d.storeMx.RUnlock()

	addr, err := d.ClosestStoresForHash(hash)
	if err != nil {
		return nil, err
	}
	d.storeMx.Lock()
	d.storeCache[hash] = addr
	d.storeMx.Unlock()

	return addr, nil
}

// Get mutable value from DHT network.
// ErrValueNotFound error is returned in case of hash not found.
func (d *DHT) Get(hash string, publicKey []byte, seq int, salt string) (string, error) {
	addr, err := d.closestStoresForHash(hash)
	if err != nil {
		return "", errors.Wrap(err, "finding peers for get failed")
	}
	ret, err := d.public.MGetAll(hash, publicKey, seq, salt, addr...)
	if err != nil {
		if strings.Contains(err.Error(), "value not found for id") {
			return "", ErrValueNotFound
		}
		return "", errors.Wrap(err, "retrieving value from the DHT network failed")
	}
	return ret, nil
}

// Put mutable value to DHT network.
func (d *DHT) Put(val *dht.MutablePut) error {
	addr, err := d.closestStoresForHash(val.Target)
	if err != nil {
		return errors.Wrap(err, "finding peers for put failed")
	}
	err = d.public.MPutAll(val, addr...)
	if err != nil {
		return errors.Wrap(err, "storing value in the DHT network failed")
	}
	return nil
}

// MutableTarget creates mutable struct for private key.
func MutableTarget(privateKey ed25519.PrivateKey, val string, seq int, salt string) (*dht.MutablePut, error) {
	m, err := dht.PutFromPvk(val, salt, privateKey, seq, seq-1)
	if err != nil {
		return &dht.MutablePut{}, errors.Wrap(err, "failed to create mutable target")
	}
	return m, nil
}
