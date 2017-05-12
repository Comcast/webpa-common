package aws

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"io/ioutil"
	"net/http"
)

// base64Decode performs a base64 decode on the supplied string
func base64Decode(msg *SNSMessage) (b []byte, err error) {
	b, err = base64.StdEncoding.DecodeString(msg.Signature)
	if err != nil {
		return b, err
	}
	
	return b, err
}


// getPemFile obtains a PEM file from the passed url string
func (v *Validator) getPemFile(address string) (body []byte, err error) {
	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return
	}
	
	resp, err := v.client.Do(req)
	if err != nil {
		return
	}
	
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	
	return
}

// getCertificate return a x509 parsed PEM file certificate
func getCerticate(b []byte) (cert *x509.Certificate, err error) {
	block, _ := pem.Decode(b)
	if block == nil {
		return
	}
	
	cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return
	}
	
	return
}

// generateSignature uses message values to replicate signature
// Values are delimited with newline characters
// Name/value pairs are sorted by name in byte sort order.
func generateSignature(msg *SNSMessage) hash.Hash {
	var formated string
	if msg.Subject != "" {
		formated = fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n", 
			"Message", msg.Message, 
			"MessageId", msg.MessageId, 
			"Subject", msg.Subject, 
			"Timestamp", msg.Timestamp, 
			"TopicArn", msg.TopicArn,
			"Type", msg.Type,
		)
	} else {
		formated = fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n", 
			"Message", msg.Message, 
			"MessageId", msg.MessageId,  
			"Timestamp", msg.Timestamp, 
			"TopicArn", msg.TopicArn,
			"Type", msg.Type,
		)
	}
	
	h := sha1.New()
	h.Write([]byte(formated))
	
	return h
}

type Validator struct {
	client *http.Client
}

type SNSValidator interface {
	Validate(*SNSMessage) (bool, error)
}

func NewValidator(client *http.Client) *Validator {
	if client == nil {
		client = new(http.Client)
	}
	
	v := new(Validator)
	v.client = client
	
	return v
}

func NewSNSValidator() SNSValidator {
	return NewValidator(nil)
}

// Validator validates an Amazon SNS message signature
func (v *Validator) Validate(msg *SNSMessage) (ok bool, err error) {
	var decodedSignature []byte
	if decodedSignature, err = base64Decode(msg); err != nil {
		return
	}
	
	var p []byte
	if p, err = v.getPemFile(msg.SigningCertURL); err != nil {
		return
	}
	
	var cert *x509.Certificate
	if cert, err = getCerticate(p); err != nil {
		return
	}
	
	var pub *rsa.PublicKey
	var okay bool
	if pub, okay = cert.PublicKey.(*rsa.PublicKey); !okay {
		return okay, errors.New("unknown type of public key")
	}
	
	h := generateSignature(msg)
	
	if err = rsa.VerifyPKCS1v15(pub, crypto.SHA1, h.Sum(nil), decodedSignature); err != nil {
		// signature verification failed
		return
	}
	
	// valid signature
	ok = true

	return
}