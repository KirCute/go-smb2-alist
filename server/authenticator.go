package smb2

import (
	"encoding/asn1"

	"github.com/KirCute/go-smb2-alist/internal/ntlm"
	"github.com/KirCute/go-smb2-alist/internal/spnego"
)

type Authenticator interface {
	oid() asn1.ObjectIdentifier
	challenge(sc []byte) ([]byte, error)
	authenticate(sc []byte) (string, error)
	sum(bs []byte) []byte // GSS_getMIC
	sessionKey() []byte   // QueryContextAttributes(ctx, SECPKG_ATTR_SESSION_KEY, &out)
}

// NTLMAuthenticator implements session-setup through NTLMv2.
// It doesn't support NTLMv1. You can use Hash instead of Password.
type NTLMAuthenticator struct {
	UserPassword func(string) (string, bool)
	TargetSPN    string
	NbDomain     string
	NbName       string
	DnsName      string
	DnsDomain    string
	AllowGuest   func() bool

	ntlm   *ntlm.Server
	seqNum uint32
}

func (i *NTLMAuthenticator) oid() asn1.ObjectIdentifier {
	return spnego.NlmpOid
}

func (i *NTLMAuthenticator) challenge(sc []byte) ([]byte, error) {
	i.ntlm = ntlm.NewServer(i.TargetSPN, i.NbName, i.NbDomain, i.DnsName, i.DnsDomain)
	i.ntlm.SetAccount(i.UserPassword)
	i.ntlm.SetAllowGuest(i.AllowGuest)

	nmsg, err := i.ntlm.Challenge(sc)
	if err != nil {
		return nil, err
	}
	return nmsg, nil
}

func (i *NTLMAuthenticator) authenticate(sc []byte) (string, error) {
	err := i.ntlm.Authenticate(sc)
	if err == nil {
		return i.ntlm.Session().User(), nil
	}
	return "", err
}

func (i *NTLMAuthenticator) sum(bs []byte) []byte {
	mic, _ := i.ntlm.Session().Sum(bs, i.seqNum)
	return mic
}

func (i *NTLMAuthenticator) sessionKey() []byte {
	return i.ntlm.Session().SessionKey()
}

func (i *NTLMAuthenticator) infoMap() *ntlm.InfoMap {
	return i.ntlm.Session().InfoMap()
}
