package cacheutil

// Repersents a potential Reason to not cache an object.
//
// Applications may wish to ignore specific reasons, which will make them non-RFC
// compliant, but this type gives them specific cases they can choose to ignore,
// making them compliant in as many cases as they can.
type Reason int

const (

	// The request method was POST and an Expiration header was not supplied.
	ReasonRequestMethodPOST Reason = iota

	// The request method was PUT and PUTs are not cachable.
	ReasonRequestMethodPUT

	// The request method was DELETE and DELETEs are not cachable.
	ReasonRequestMethodDELETE

	// The request method was CONNECT and CONNECTs are not cachable.
	ReasonRequestMethodCONNECT

	// The request method was OPTIONS and OPTIONS are not cachable.
	ReasonRequestMethodOPTIONS

	// The request method was TRACE and TRACEs are not cachable.
	ReasonRequestMethodTRACE

	// The request method was not recognized by cachecontrol, and should not be cached.
	ReasonRequestMethodUnknown

	// The request included an Cache-Control: no-store header
	ReasonRequestNoStore

	// The request included an Authorization header without an explicit Public or Expiration time: http://tools.ietf.org/html/rfc7234#section-3.2
	ReasonRequestAuthorizationHeader

	// The response included an Cache-Control: no-store header
	ReasonResponseNoStore

	// The response included an Cache-Control: private header and this is not a Private cache
	ReasonResponsePrivate

	// The response failed to meet at least one of the conditions specified in RFC 7234 section 3: http://tools.ietf.org/html/rfc7234#section-3
	ReasonResponseUncachableByDefault
)

func (r Reason) String() string {
	switch r {
	case ReasonRequestMethodPOST:
		return "ReasonRequestMethodPOST"
	case ReasonRequestMethodPUT:
		return "ReasonRequestMethodPUT"
	case ReasonRequestMethodDELETE:
		return "ReasonRequestMethodDELETE"
	case ReasonRequestMethodCONNECT:
		return "ReasonRequestMethodCONNECT"
	case ReasonRequestMethodOPTIONS:
		return "ReasonRequestMethodOPTIONS"
	case ReasonRequestMethodTRACE:
		return "ReasonRequestMethodTRACE"
	case ReasonRequestMethodUnknown:
		return "ReasonRequestMethodUnkown"
	case ReasonRequestNoStore:
		return "ReasonRequestNoStore"
	case ReasonRequestAuthorizationHeader:
		return "ReasonRequestAuthorizationHeader"
	case ReasonResponseNoStore:
		return "ReasonResponseNoStore"
	case ReasonResponsePrivate:
		return "ReasonResponsePrivate"
	case ReasonResponseUncachableByDefault:
		return "ReasonResponseUncachableByDefault"
	}

	panic(r)
}
