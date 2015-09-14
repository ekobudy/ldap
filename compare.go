package ldap

import (
	"github.com/go-asn1-ber/asn1-ber"
)

/*
CompareRequest ::= [APPLICATION 14] SEQUENCE {
    entry           LDAPDN,
    ava             AttributeValueAssertion }

AttributeValueAssertion ::= SEQUENCE {
    attributeDesc   AttributeDescription,
    assertionValue  AssertionValue }
*/

type CompareRequest struct {
	DN       string
	Name     string
	Value    string
	Controls []Control
}

func (l *Connection) Compare(req *CompareRequest) (bool, error) {
	messageID, ok := l.nextMessageID()
	if !ok {
		return false, newError(ErrorClosing, "MessageID channel is closed.")
	}

	encodedCompare, err := encodeCompareRequest(req)
	if err != nil {
		return false, err
	}

	packet, err := requestBuildPacket(messageID, encodedCompare, req.Controls)
	if err != nil {
		return false, err
	}

	// CompareTrue = 6, CompareFalse = 5
	// returns an "Error"
	err = l.sendReqRespPacket(messageID, packet)
	if lerr, ok := err.(*Error); ok {
		return lerr.ResultCode == ResultCompareTrue, nil
	} else {
		return false, err
	}
	//return l.sendReqRespPacket(messageID, packet)
}

func encodeCompareRequest(req *CompareRequest) (*ber.Packet, error) {
	p := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ber.Tag(ApplicationCompareRequest), nil, ApplicationCompareRequest.String())
	p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, req.DN, "LDAP DN"))
	ava, err := encodeItem([]string{req.Name, "=", req.Value})
	if err != nil {
		return nil, err
	}
	p.AppendChild(ava)
	return p, nil
}

func NewCompareRequest(dn, name, value string) (req *CompareRequest) {
	req = &CompareRequest{DN: dn, Name: name, Value: value, Controls: make([]Control, 0)}
	return
}
