/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		SPDX-License-Identifier: Apache-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package sw

import (
	"crypto/rand"
	"fmt"
	"github.com/fabric-creed/fabric-hub/pkg/common/sw/sig"

	"github.com/fabric-creed/cryptogm/sm2"
)

func signSM2(k *sm2.PrivateKey, digest []byte, opts SignerOpts) (signature []byte, err error) {
	r, s, err := sm2.Sign(rand.Reader, k, digest)
	if err != nil {
		return nil, err
	}
	return sig.MarshalSM2Signature(r, s)
}

func verifySM2(k *sm2.PublicKey, signature, digest []byte, opts SignerOpts) (valid bool, err error) {
	r, s, err := sig.UnmarshalSM2Signature(signature)
	if err != nil {
		return false, fmt.Errorf("Failed unmashalling signature [%s]", err)
	}
	return sm2.Verify(k, digest, r, s), nil
}

type sm2Signer struct{}

func (s *sm2Signer) Sign(k Key, digest []byte, opts SignerOpts) ([]byte, error) {
	return signSM2(k.(*sm2PrivateKey).privKey, digest, opts)
}

type sm2PrivateKeyVerifier struct{}

func (v *sm2PrivateKeyVerifier) Verify(k Key, signature, digest []byte, opts SignerOpts) (bool, error) {
	return verifySM2(&(k.(*sm2PrivateKey).privKey.PublicKey), signature, digest, opts)
}

type sm2PublicKeyKeyVerifier struct{}

func (v *sm2PublicKeyKeyVerifier) Verify(k Key, signature, digest []byte, opts SignerOpts) (bool, error) {
	return verifySM2(k.(*sm2PublicKey).pubKey, signature, digest, opts)
}
