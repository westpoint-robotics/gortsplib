package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/westpoint-robotics/gortsplib/pkg/base"
	"github.com/westpoint-robotics/gortsplib/pkg/headers"
	"github.com/westpoint-robotics/gortsplib/pkg/url"
)

// Validator allows to validate credentials generated by a Sender.
type Validator struct {
	user       string
	userHashed bool
	pass       string
	passHashed bool
	methods    []headers.AuthMethod
	realm      string
	nonce      string
}

// NewValidator allocates a Validator.
// If methods is nil, the Basic and Digest methods are used.
func NewValidator(user string, pass string, methods []headers.AuthMethod) *Validator {
	if methods == nil {
		methods = []headers.AuthMethod{headers.AuthBasic, headers.AuthDigest}
	}

	userHashed := false
	if strings.HasPrefix(user, "sha256:") {
		user = strings.TrimPrefix(user, "sha256:")
		userHashed = true
	}

	passHashed := false
	if strings.HasPrefix(pass, "sha256:") {
		pass = strings.TrimPrefix(pass, "sha256:")
		passHashed = true
	}

	// if credentials are hashed, only basic auth is supported
	if userHashed || passHashed {
		methods = []headers.AuthMethod{headers.AuthBasic}
	}

	nonceByts := make([]byte, 16)
	rand.Read(nonceByts)
	nonce := hex.EncodeToString(nonceByts)

	return &Validator{
		user:       user,
		userHashed: userHashed,
		pass:       pass,
		passHashed: passHashed,
		methods:    methods,
		realm:      "IPCAM",
		nonce:      nonce,
	}
}

// Header generates the WWW-Authenticate header needed by a client to
// authenticate.
func (va *Validator) Header() base.HeaderValue {
	var ret base.HeaderValue
	for _, m := range va.methods {
		switch m {
		case headers.AuthBasic:
			ret = append(ret, (&headers.Authenticate{
				Method: headers.AuthBasic,
				Realm:  &va.realm,
			}).Marshal()...)

		case headers.AuthDigest:
			ret = append(ret, headers.Authenticate{
				Method: headers.AuthDigest,
				Realm:  &va.realm,
				Nonce:  &va.nonce,
			}.Marshal()...)
		}
	}
	return ret
}

// ValidateRequest validates a request sent by a client.
func (va *Validator) ValidateRequest(req *base.Request, baseURL *url.URL) error {
	var auth headers.Authorization
	err := auth.Unmarshal(req.Header["Authorization"])
	if err != nil {
		return err
	}

	switch auth.Method {
	case headers.AuthBasic:
		if !va.userHashed {
			if auth.BasicUser != va.user {
				return fmt.Errorf("wrong response")
			}
		} else {
			if sha256Base64(auth.BasicUser) != va.user {
				return fmt.Errorf("wrong response")
			}
		}

		if !va.passHashed {
			if auth.BasicPass != va.pass {
				return fmt.Errorf("wrong response")
			}
		} else {
			if sha256Base64(auth.BasicPass) != va.pass {
				return fmt.Errorf("wrong response")
			}
		}

	default: // headers.AuthDigest
		if auth.DigestValues.Realm == nil {
			return fmt.Errorf("realm is missing")
		}

		if auth.DigestValues.Nonce == nil {
			return fmt.Errorf("nonce is missing")
		}

		if auth.DigestValues.Username == nil {
			return fmt.Errorf("username is missing")
		}

		if auth.DigestValues.URI == nil {
			return fmt.Errorf("uri is missing")
		}

		if auth.DigestValues.Response == nil {
			return fmt.Errorf("response is missing")
		}

		if *auth.DigestValues.Nonce != va.nonce {
			return fmt.Errorf("wrong nonce")
		}

		if *auth.DigestValues.Realm != va.realm {
			return fmt.Errorf("wrong realm")
		}

		if *auth.DigestValues.Username != va.user {
			return fmt.Errorf("wrong username")
		}

		ur := req.URL

		if *auth.DigestValues.URI != ur.String() {
			// in SETUP requests, VLC strips the control attribute.
			// try again with the base URL.
			if baseURL != nil {
				ur = baseURL
				if *auth.DigestValues.URI != ur.String() {
					return fmt.Errorf("wrong URL")
				}
			} else {
				return fmt.Errorf("wrong URL")
			}
		}

		response := md5Hex(md5Hex(va.user+":"+va.realm+":"+va.pass) +
			":" + va.nonce + ":" + md5Hex(string(req.Method)+":"+ur.String()))

		if *auth.DigestValues.Response != response {
			return fmt.Errorf("wrong response")
		}
	}

	return nil
}
