package ed25519

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ed25519"
)

// mustDecodeStrHex decodes and hex string to its byte strings or panics.
func mustDecodeStrHex(in string) []byte {
	x, e := hex.DecodeString(in)
	if e != nil {
		panic(e)
	}
	return x
}

// mustDecodeHex decodes a []byte(hexadecimal) or panics
func mustDecodeHex(in []byte) []byte {
	out := make([]byte, len(in))
	n, dErr := hex.Decode(out, in)
	if dErr != nil {
		panic(dErr)
	}
	return out[:n]
}

var bep44testPvk = "e06d3183d14159228433ed599221b80bd0a5ce8352e4bdf0262f76786ef1c74db7e7a9fea2c0eb269d61e3b38e450a22e754941ac78479d6c54e1faf6037881d"

// PvkFromDir creates/reads a pvk from given directory.
// it handles "bep44test" name as static value (see the bep tests).
func PvkFromDir(dir, name string) (PrivateKey, ed25519.PublicKey, error) {
	if name == "bep44test" {
		pvk := ed25519.PrivateKey(mustDecodeStrHex(bep44testPvk))
		return PrivateKey(pvk), PublicKeyFromPvk(pvk), nil
	}
	os.MkdirAll(dir, os.ModePerm)
	fileName := name + ".key"
	file := filepath.Join(dir, fileName)
	if _, statErr := os.Stat(file); !os.IsNotExist(statErr) {
		b, readErr := ioutil.ReadFile(file)
		if readErr != nil {
			return nil, nil, readErr
		}
		b = bytes.TrimRight(b, "\n")
		pvk := ed25519.PrivateKey(mustDecodeHex(b))
		pvk = pvk[:64]
		return PrivateKey(pvk), PublicKeyFromPvk(pvk), nil
	}
	_, pvk, err := ed25519.GenerateKey(nil)
	pvk = pvk[:64]
	if err != nil {
		return nil, nil, err
	}
	return PrivateKey(pvk), PublicKeyFromPvk(pvk), ioutil.WriteFile(file, []byte(hex.EncodeToString(pvk)), os.ModePerm)
}

// PvkFromHex returns a tuple of (PrivateKey, ed25519.PublicKey) given an hex representation.
func PvkFromHex(pvkHex string) (PrivateKey, ed25519.PublicKey, error) {
	k, err := hex.DecodeString(pvkHex)
	if err != nil {
		return nil, nil, err
	}

	pvk := PrivateKey(k)
	return pvk, PublicKeyFromPvk(pvk), nil
}

// PbkFromHex returns an ed25519.PublicKey given an hex representation.
func PbkFromHex(pbk string) (ed25519.PublicKey, error) {
	k, err := hex.DecodeString(pbk)
	if err != nil {
		return nil, err
	}
	return ed25519.PublicKey(k), nil
}
