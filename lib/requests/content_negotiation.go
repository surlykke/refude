// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package requests

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"regexp"
	"strings"
)

//  From RFC 7231:
//
//  Accept = #( media-range [ accept-params ] )
//  media-range    = ( "*/*"  / ( type "/" "*" )  / ( type "/" subtype )) *( OWS ";" OWS parameter )
//  accept-params  = weight *( accept-ext )
//  accept-ext = OWS ";" OWS token [ "=" ( token / quoted-string ) ]
//
//  and from RFC 7230:
//
//  token          = 1*tchar
//  tchar          = "!" / "#" / "$" / "%" / "&" / "'" / "*"  / "+" / "-" / "." / "^" / "_" / "`" / "|" / "~"
//                   / DIGIT / ALPHA
//                   ; any VCHAR, except delimiters
//  quoted-string  = DQUOTE *( qdtext / quoted-pair ) DQUOTE
//  qdtext         = HTAB / SP /%x21 / %x23-5B / %x5D-7E / obs-text
//  quoted-pair    = "\" ( HTAB / SP / VCHAR / obs-text )
//  obs-text       = %x80-FF

// Later versions of http has deprecated use of acsii > 127 (obs-text here), and to keep things simple we don't accept it either. (Probably
// a violation of the spec. But it's such a pain...)
// Also the spec allows escaping a lot of different characters inside a string with '\' but noting that the sender 'SHOULD NOT' use escaping
// for other characters than '"' and '\'. We tighten that to 'MUST NOT'. (Again, probably a violation of specs..)
//
// We do _not_ handle suffixes, so if, eg., application/myjsonformat+json is offered and application/json is acceptet, that will not match. You'll
// have to offer both application/myjsonformat+json and application/json

const tChar = "[a-zA-Z0-9!#$%&'*+\\-.^_`|~']"
const stringChar = "[\t !#-[\\]-~]" // tab, space and all visible chars except '"' and '\'
const qoutedPair = "\\[\"\\]"       // '\' followed by '"' or '\'
const ws = "[ \t]"

var mimetypePattern = regexp.MustCompile(fmt.Sprintf("^%s*(%s+/%s+)%s*", ws, tChar, tChar, ws))
var parameterPattern = regexp.MustCompile(fmt.Sprintf("^%s*(%s+)=(%s+|\"(%s|%s)*\")%s*", ws, tChar, tChar, stringChar, qoutedPair, ws))

// May not be called with zero offers
func Negotiate(r *http.Request, offers...string) (string, error) {
	if accept, ok := extractAcceptString(r); !ok {
		return offers[0], nil
	} else {
		return read(accept, offers...)
	}
}

func extractAcceptString(r *http.Request) (string, bool) {
	if accepts, ok := r.Header["Accept"]; !ok || len(accepts) == 0 {
		return "", false
	} else {
		var accept = accepts[0]
		for i := 1; i < len(accepts); i++ {
			accept = accept + "," + accepts[i]
		}
		return accept, true
	}
}

func read(accept string, offers...string) (string, error) {
	var weights = make([]int, len(offers), len(offers))

	// Read through accept
	// Each	iteration of this loop reads one media range (if any left) with parameters and weight, eg: image/png;q=0.3
	for {
		if mrMatch := mimetypePattern.FindStringSubmatch(accept); mrMatch != nil {
			accept = accept[len(mrMatch[0]):]
			var requiredRange, requiredTypeLen = mrMatch[1], strings.Index(mrMatch[1], "/")
			var weight = 1000
			var haveWeight = false
			for len(accept) > 0 && accept[0] == ';' {
				accept = accept[1:]
				if parMatch := parameterPattern.FindStringSubmatch(accept); parMatch != nil {
					// FIXME Deal with other parameters than "q"
					if !haveWeight && parMatch[1] == "q" {
						if w, ok := readWeight(parMatch[2]); !ok {
							return "", errors.New("Invalid weight: " + parMatch[2])
						} else {
							weight = w
						}
						haveWeight = true
					}
				} else {
					return "", errors.New("Invalid parameter: " + accept)
				}
			}

			for i := 0; i < len(offers); i++ {
				if match(offers[i], requiredRange, requiredTypeLen) && weights[i] < weight {
					weights[i] = weight
				}
			}

			if len(accept) == 0 || accept[0] != ',' {
				break
			}

			accept = accept[1:]
		} else {
			return "", errors.New("Invalid media-range: " + accept)
		}

	}

	var candidate, candidateWeight = -1, 0
	for i := 0; i < len(offers); i++ {
		if weights[i] > candidateWeight {
			candidate, candidateWeight = i, weights[i]
		}
	}

	if candidate < 0 {
		return "", nil
	} else {
		return offers[candidate], nil
	}
}

func match(offer string, requiredRange string, requiredTypeLen int) bool {
	if requiredRange == "*/*" {
		return true
	} else if requiredRange[requiredTypeLen + 1:] == "*" {
		return strings.HasPrefix(offer, requiredRange[0:requiredTypeLen])
	} else {
		return offer == requiredRange
	}
}

func readWeight(val string) (int, bool) {
	var weight int

	if val[0] < '0' || val[0] > '9' {
		return 0, false
	} else {
		weight = int(val[0]-'0') * 1000
	}

	var l = len(val)

	if l > 1 && val[1] != '.' {
		return 0, false
	}

	if l > 2 {
		if val[2] < '0' || val[2] > '9' {
			return 0, false
		} else {
			weight = weight + int(val[2]-'0')*100
		}
	}

	if l > 3 {
		if val[3] < '0' || val[3] > '9' {
			return 0, false
		} else {
			weight = weight + int(val[3]-'0')*10
		}
	}

	if l > 4 {
		if val[4] < '0' || val[4] > '9' {
			return 0, false
		} else {
			weight = weight + int(val[4]-'0')
		}
	}

	if l > 5 {
		return 0, false
	} else {
		return weight, true
	}
}

// Typical, from chrome: Accept: image/webp,image/apng,image/*,*/*;q=0.8
//          from firefox:        image/webp,*/*
