//go:build js && wasm

package app

import (
	"fmt"
	"os"
	"time"

	"crypto/elliptic"

	enc "github.com/named-data/ndnd/std/encoding"
	"github.com/named-data/ndnd/std/security"
	"github.com/named-data/ndnd/std/security/ndncert"
	spec_ndncert "github.com/named-data/ndnd/std/security/ndncert/tlv"
	sig "github.com/named-data/ndnd/std/security/signer"
)

func (a *App) NdncertEmail(email string, CodeCb func(status string) string) (err error) {
	// Connect to the testbed
	if err := a.WaitForConnectivity(time.Second * 5); err != nil {
		return err
	}

	// Create NDNCERT client
	certClient, err := ndncert.NewClient(a.engine, testbedRootCert)
	if err != nil {
		return err
	}

	// Request a certificate from NDNCERT
	certRes, err := certClient.RequestCert(ndncert.RequestCertArgs{
		Challenge: &ndncert.ChallengeEmail{
			Email:        email,
			CodeCallback: CodeCb,
		},
		OnProfile: func(profile *spec_ndncert.CaProfile) error {
			fmt.Fprintf(os.Stderr, "NDNCERT CA: %s\n", profile.CaInfo)
			return nil
		},
		OnProbeParam: func(key string) ([]byte, error) {
			switch key {
			case ndncert.KwEmail:
				return []byte(email), nil
			default:
				return nil, fmt.Errorf("unknown probe key: %s", key)
			}
		},
		OnChooseKey: func(suggestions []enc.Name) int {
			return 0 // choose the first key
		},
		OnKeyChosen: func(keyName enc.Name) error {
			fmt.Fprintf(os.Stderr, "Certifying key: %s\n", keyName)
			return nil
		},
	})
	if err != nil {
		return err
	}

	// Verfiy the received certificate and fetch the chain
	_, err = a.verifyTestbedCert(certRes.CertWire, true)
	if err != nil {
		return fmt.Errorf("failed to verify issued certificate: %w", err)
	}

	// Store the certificate and the signer key
	if err = a.keychain.InsertKey(certRes.Signer); err != nil {
		return err
	}
	if err = a.keychain.InsertCert(certRes.CertWire.Join()); err != nil {
		return err
	}

	return nil
}

func (a *App) NdncertDns(domain string, ConfirmCb func(recordName, expectedValue, status string) string) (err error) {
	if err := a.WaitForConnectivity(time.Second * 5); err != nil {
		return err
	}

	certClient, err := ndncert.NewClient(a.engine, testbedRootCert)
	if err != nil {
		return err
	}

	caPrefix := certClient.CaPrefix()
	if len(caPrefix) == 0 {
		return fmt.Errorf("ca prefix unavailable")
	}

	identity := caPrefix.Append(enc.NewGenericComponent(domain))
	keyName := security.MakeKeyName(identity)
	signer, err := sig.KeygenEcc(keyName, elliptic.P256())
	if err != nil {
		return fmt.Errorf("failed to generate dns challenge key: %w", err)
	}
	certClient.SetSigner(signer)
	certRes, err := certClient.RequestCert(ndncert.RequestCertArgs{
		Challenge: &ndncert.ChallengeDns{
			DomainCallback: func(status string) string {
				return domain
			},
			ConfirmationCallback: func(recordName, expectedValue, status string) string {
				if ConfirmCb == nil {
					return ""
				}
				return ConfirmCb(recordName, expectedValue, status)
			},
		},
		DisableProbe: true,
		OnProfile: func(profile *spec_ndncert.CaProfile) error {
			return nil
		},
		OnProbeParam: func(key string) ([]byte, error) {
			switch key {
			case ndncert.KwDomain:
				return []byte(domain), nil
			case ndncert.KwEmail:
				return nil, nil
			default:
				return nil, nil
			}
		},
		OnChooseKey: func(suggestions []enc.Name) int {
			return 0
		},
		OnKeyChosen: func(keyName enc.Name) error {
			fmt.Fprintf(os.Stderr, "Certifying key: %s\n", keyName)
			return nil
		},
	})
	if err != nil {
		return err
	}

	if _, err = a.verifyTestbedCert(certRes.CertWire, true); err != nil {
		return fmt.Errorf("failed to verify issued certificate: %w", err)
	}

	if err = a.keychain.InsertKey(certRes.Signer); err != nil {
		return err
	}
	if err = a.keychain.InsertCert(certRes.CertWire.Join()); err != nil {
		return err
	}

	return nil
}
