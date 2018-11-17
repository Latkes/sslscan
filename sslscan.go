// Package sslscan provides methods and structs for working with SSLLabs public API
package sslscan

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2018 ESSENTIAL KAOS                         //
//      Apache License, Version 2.0 <http://www.apache.org/licenses/LICENSE-2.0>      //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/valyala/fasthttp"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	API_URL_INFO     = "https://api.ssllabs.com/api/v2/info"
	API_URL_ANALYZE  = "https://api.ssllabs.com/api/v2/analyze"
	API_URL_DETAILED = "https://api.ssllabs.com/api/v2/getEndpointData"
)

const (
	STATUS_IN_PROGRESS = "IN_PROGRESS"
	STATUS_DNS         = "DNS"
	STATUS_READY       = "READY"
	STATUS_ERROR       = "ERROR"
)

const (
	SSLCSC_STATUS_FAILED              = -1
	SSLCSC_STATUS_UNKNOWN             = 0
	SSLCSC_STATUS_NOT_VULNERABLE      = 1
	SSLCSC_STATUS_POSSIBLE_VULNERABLE = 2
	SSLCSC_STATUS_VULNERABLE          = 3
)

const (
	LUCKY_MINUS_STATUS_FAILED         = -1
	LUCKY_MINUS_STATUS_UNKNOWN        = 0
	LUCKY_MINUS_STATUS_NOT_VULNERABLE = 1
	LUCKY_MINUS_STATUS_VULNERABLE     = 2
)

const (
	POODLE_STATUS_TIMEOUT           = -3
	POODLE_STATUS_TLS_NOT_SUPPORTED = -2
	POODLE_STATUS_FAILED            = -1
	POODLE_STATUS_UNKNOWN           = 0
	POODLE_STATUS_NOT_VULNERABLE    = 1
	POODLE_STATUS_VULNERABLE        = 2
)

const (
	REVOCATION_STATUS_NOT_CHECKED            = 0
	REVOCATION_STATUS_REVOKED                = 1
	REVOCATION_STATUS_NOT_REVOKED            = 2
	REVOCATION_STATUS_REVOCATION_CHECK_ERROR = 3
	REVOCATION_STATUS_NO_REVOCATION_INFO     = 4
	REVOCATION_STATUS_INTERNAL_INFO          = 5
)

const (
	HSTS_STATUS_UNKNOWN  = "unknown"
	HSTS_STATUS_ABSENT   = "absent"
	HSTS_STATUS_PRESENT  = "present"
	HSTS_STATUS_INVALID  = "invalid"
	HSTS_STATUS_DISABLED = "disabled"
	HSTS_STATUS_ERROR    = "error"
)

const (
	HPKP_STATUS_UNKNOWN    = "unknown"
	HPKP_STATUS_ABSENT     = "absent"
	HPKP_STATUS_INVALID    = "invalid"
	HPKP_STATUS_DISABLED   = "disabled"
	HPKP_STATUS_INCOMPLETE = "incomplete"
	HPKP_STATUS_VALID      = "valid"
	HPKP_STATUS_ERROR      = "error"
)

const (
	DROWN_STATUS_ERROR                 = "error"
	DROWN_STATUS_UNKNOWN               = "unknown"
	DROWN_STATUS_NOT_CHECKED           = "not_checked"
	DROWN_STATUS_NOT_CHECKED_SAME_HOST = "not_checked_same_host"
	DROWN_STATUS_HANDSHAKE_FAILURE     = "handshake_failure"
	DROWN_STATUS_SSLV2                 = "sslv2"
	DROWN_STATUS_KEY_MATCH             = "key_match"
	DROWN_STATUS_HOSTNAME_MATCH        = "hostname_match"
)

const (
	PROTOCOL_INTOLERANCE_TLS1_0 = 1 << iota
	PROTOCOL_INTOLERANCE_TLS1_1
	PROTOCOL_INTOLERANCE_TLS1_2
	PROTOCOL_INTOLERANCE_TLS1_3
	PROTOCOL_INTOLERANCE_TLS1_152
	PROTOCOL_INTOLERANCE_TLS2_152
)

// Package version
const VERSION = "9.0.1"

// ////////////////////////////////////////////////////////////////////////////////// //

type API struct {
	Info   *Info
	Client *fasthttp.Client
}

type AnalyzeParams struct {
	Public         bool
	StartNew       bool
	FromCache      bool
	MaxAge         int
	All            bool
	IgnoreMismatch bool
}

type AnalyzeProgress struct {
	host       string
	prevStatus string

	api *API
}

// DOCS: https://github.com/ssllabs/ssllabs-scan/blob/stable/ssllabs-api-docs.md

type Info struct {
	ClientMaxAssessments int      `json:"clientMaxAssessments"` // -
	CriteriaVersion      string   `json:"criteriaVersion"`      // rating criteria version as a string (e.g., "2009f")
	CurrentAssessments   int      `json:"currentAssessments"`   // the number of ongoing assessments submitted by this client
	EngineVersion        string   `json:"engineVersion"`        // SSL Labs software version as a string (e.g., "1.11.14")
	MaxAssessments       int      `json:"maxAssessments"`       // the maximum number of concurrent assessments the client is allowed to initiate
	Messages             []string `json:"messages"`             // a list of messages (strings). Messages can be public (sent to everyone) and private (sent only to the invoking client). Private messages are prefixed with "[Private]".
	NewAssessmentCoolOff int      `json:"newAssessmentCoolOff"` // he cool-off period after each new assessment; you're not allowed to submit a new assessment before the cool-off expires, otherwise you'll get a 429
}

type AnalyzeInfo struct {
	CriteriaVersion string          `json:"criteriaVersion"` // grading criteria version (e.g., "2009")
	Endpoints       []*EndpointInfo `json:"endpoints"`       // list of Endpoint objects
	EngineVersion   string          `json:"engineVersion"`   // assessment engine version (e.g., "1.0.120")
	Host            string          `json:"host"`            // assessment host, which can be a hostname or an IP address
	IsPublic        bool            `json:"isPublic"`        // true if this assessment publicly available (listed on the SSL Labs assessment boards)
	Port            int             `json:"port"`            // assessment port (e.g., 443)
	Protocol        string          `json:"protocol"`        // protocol (e.g., HTTP)
	StartTime       int64           `json:"startTime"`       // assessment starting time, in milliseconds since 1970
	Status          string          `json:"status"`          // assessment status; possible values: DNS, ERROR, IN_PROGRESS, and READY
	StatusMessage   string          `json:"statusMessage"`   // status message in English. When status is ERROR, this field will contain an error message
	TestTime        int64           `json:"testTime"`        // assessment completion time, in milliseconds since 1970
}

type EndpointInfo struct {
	Delegation           int              `json:"delegation"`           // indicates domain name delegation with and without the www prefix
	Details              *EndpointDetails `json:"details"`              // this field contains an EndpointDetails object. It's not present by default, but can be enabled by using the "all" paramerer to the analyze API call
	Duration             int              `json:"duration"`             // assessment duration, in milliseconds
	ETA                  int              `json:"eta"`                  // estimated time, in seconds, until the completion of the assessment
	Grade                string           `json:"grade"`                // possible values: A+, A-, A-F, T (no trust) and M (certificate name mismatch)
	GradeTrustIgnored    string           `json:"gradeTrustIgnored"`    // grade (as above), if trust issues are ignored
	HasWarnings          bool             `json:"hasWarnings"`          // if this endpoint has warnings that might affect the score (e.g., get A- instead of A).
	IPAdress             string           `json:"ipAddress"`            // endpoint IP address, in IPv4 or IPv6 format
	IsExceptional        bool             `json:"isExceptional"`        // this flag will be raised when an exceptional configuration is encountered. The SSL Labs test will give such sites an A+
	Progress             int              `json:"progress"`             // assessment progress, which is a value from 0 to 100, and -1 if the assessment has not yet started
	ServerName           string           `json:"serverName"`           // server name retrieved via reverse DNS
	StatusDetails        string           `json:"statusDetails"`        // code of the operation currently in progress
	StatusDetailsMessage string           `json:"statusDetailsMessage"` // description of the operation currently in progress
	StatusMessage        string           `json:"statusMessage"`        // assessment status message
}

type EndpointDetails struct {
	Cert                           *Cert         `json:"cert"`                           // certificate information
	ChaCha20Preference             bool          `json:"chaCha20Preference"`             // -
	Chain                          *Chain        `json:"chain"`                          // chain information
	CompressionMethods             int           `json:"compressionMethods"`             // integer value that describes supported compression methods
	DHPrimes                       []string      `json:"dhPrimes"`                       // list of hex-encoded DH primes used by the server
	DHUsesKnownPrimes              int           `json:"dhUsesKnownPrimes"`              // whether the server uses known DH primes
	DHYsReuse                      bool          `json:"dhYsReuse"`                      // true if the DH ephemeral server value is reused
	DrownErrors                    bool          `json:"drownErrors"`                    // true if error occurred in drown test
	DrownHosts                     []DrownHost   `json:"drownHosts"`                     // list of drown hosts
	DrownVulnerable                bool          `json:"drownVulnerable"`                // true if server vulnerable to drown attack
	FallbackSCSV                   bool          `json:"fallbackScsv"`                   // true if the server supports TLS_FALLBACK_SCSV, false if it doesn't
	ForwardSecrecy                 int           `json:"forwardSecrecy"`                 // indicates support for Forward Secrecy
	Freak                          bool          `json:"freak"`                          // true of the server is vulnerable to the FREAK attack
	HasSCT                         int           `json:"hasSct"`                         // information about the availability of certificate transparency information (embedded SCTs)
	Heartbeat                      bool          `json:"heartbeat"`                      // true if the server supports the Heartbeat extension
	Heartbleed                     bool          `json:"heartbleed"`                     // true if the server is vulnerable to the Heartbleed attack
	HostStartTime                  int64         `json:"hostStartTime"`                  // endpoint assessment starting time, in milliseconds since 1970. This field is useful when test results are retrieved in several HTTP invocations. Then, you should check that the hostStartTime value matches the startTime value of the host
	HPKPPolicy                     *HPKPPolicy   `json:"hpkpPolicy"`                     // server's HPKP policy
	HPKPRoPolicy                   *HPKPPolicy   `json:"hpkpRoPolicy"`                   // server's HPKP RO (Report Only) policy
	HSTSPolicy                     *HSTSPolicy   `json:"hstsPolicy"`                     // server's HSTS policy
	HSTSPreloads                   []HSTSPreload `json:"hstsPreloads"`                   // information about preloaded HSTS policies
	HTTPForwarding                 string        `json:"httpForwarding"`                 // available on a server that responded with a redirection to some other hostname
	HTTPStatusCode                 int           `json:"httpStatusCode"`                 // status code of the final HTTP response seen
	Key                            *Key          `json:"key"`                            // key information
	Logjam                         bool          `json:"logjam"`                         // true if the server uses DH parameters weaker than 1024 bits
	MiscIntolerance                int           `json:"miscIntolerance"`                // indicates protocol version intolerance issues
	NonPrefixDelegation            bool          `json:"nonPrefixDelegation"`            // true if this endpoint is reachable via a hostname without the www prefix
	NPNProtocols                   string        `json:"npnProtocols"`                   // space separated list of supported protocols
	OCSPStapling                   bool          `json:"ocspStapling"`                   // true if OCSP stapling is deployed on the server
	OpenSslCCS                     int           `json:"openSslCcs"`                     // results of the CVE-2014-0224 test
	OpenSSLLuckyMinus20            int           `json:"openSSLLuckyMinus20"`            // results of the CVE-2016-2107 test
	Poodle                         bool          `json:"poodle"`                         // true if the endpoint is vulnerable to POODLE
	PoodleTLS                      int           `json:"poodleTls"`                      // results of the POODLE TLS test
	PrefixDelegation               bool          `json:"prefixDelegation"`               // true if this endpoint is reachable via a hostname with the www prefix
	ProtocolIntolerance            int           `json:"protocolIntolerance"`            // indicates protocol version intolerance issues
	Protocols                      []*Protocol   `json:"protocols"`                      // supported protocols
	RC4Only                        bool          `json:"rc4Only"`                        // true if only RC4 suites are supported
	RC4WithModern                  bool          `json:"rc4WithModern"`                  // true if RC4 is used with modern clients
	RenegSupport                   int           `json:"renegSupport"`                   // this is an integer value that describes the endpoint support for renegotiation
	ServerSignature                string        `json:"serverSignature"`                // contents of the HTTP Server response header when known
	SessionResumption              int           `json:"sessionResumption"`              // this is an integer value that describes endpoint support for session resumption
	SessionTickets                 int           `json:"sessionTickets"`                 // indicates support for Session Tickets
	SIMS                           *SIMS         `json:"sims"`                           // sims
	SNIRequired                    bool          `json:"sniRequired"`                    // if SNI support is required to access the web site
	StaplingRevocationErrorMessage string        `json:"staplingRevocationErrorMessage"` // description of the problem with the stapled OCSP response, if any
	StaplingRevocationStatus       int           `json:"staplingRevocationStatus"`       // same as Cert.revocationStatus, but for the stapled OCSP response
	Suites                         *Suites       `json:"suites"`                         // supported cipher suites
	SupportsALPN                   bool          `json:"supportsAlpn"`                   // -
	SupportsNPN                    bool          `json:"supportsNpn"`                    // true if the server supports NPN
	SupportsRC4                    bool          `json:"supportsRc4"`                    // supportsRc4
	VulnBeast                      bool          `json:"vulnBeast"`                      // true if the endpoint is vulnerable to the BEAST attack
}

type Key struct {
	Size       int    `json:"size"`       // key size, e.g., 1024 or 2048 for RSA and DSA, or 256 bits for EC
	Alg        string `json:"alg"`        // key algorithm; possible values: RSA, DSA, and EC
	DebianFlaw bool   `json:"debianFlaw"` // true if we suspect that the key was generated using a weak random number generator (detected via a blacklist database)
	Strength   int    `json:"strength"`   // key size expressed in RSA bits
	Q          *int   `json:"q"`          // 0 if key is insecure, null otherwise
}

type Chain struct {
	Certs  []*ChainCert `json:"certs"`
	Issues int          `json:"issues"`
}

type Cert struct {
	AltNames             []string `json:"altNames"`             // alternative names
	CommonNames          []string `json:"commonNames"`          // common names extracted from the subject
	CRLRevocationStatus  int      `json:"crlRevocationStatus"`  // same as revocationStatus, but only for the CRL information (if any)
	CRLURIs              []string `json:"crlURIs"`              // CRL URIs extracted from the certificate
	IssuerLabel          string   `json:"issuerLabel"`          // issuer name
	IssuerSubject        string   `json:"issuerSubject"`        // issuer subject
	Issues               int      `json:"issues"`               // list of certificate issues, one bit per issue
	MustStaple           int      `json:"mustStaple"`           // a number that describes the must staple feature extension status
	NotAfter             int64    `json:"notAfter"`             // timestamp after which the certificate is not valid
	NotBefore            int64    `json:"notBefore"`            // timestamp before which the certificate is not valid
	OCSPRevocationStatus int      `json:"ocspRevocationStatus"` // same as revocationStatus, but only for the OCSP information (if any)
	OCSPURIs             []string `json:"ocspURIs"`             // OCSP URIs extracted from the certificate
	PINSHA256            string   `json:"pinSha256"`            // -
	RevocationInfo       int      `json:"revocationInfo"`       // a number that represents revocation information present in the certificate
	RevocationStatus     int      `json:"revocationStatus"`     // a number that describes the revocation status of the certificate
	SCT                  bool     `json:"sct"`                  // true if the certificate contains an embedded SCT
	SGC                  int      `json:"sgc"`                  // Server Gated Cryptography support
	SHA1Hash             string   `json:"sha1Hash"`             // -
	SigAlg               string   `json:"sigAlg"`               // certificate signature algorithm
	Subject              string   `json:"subject"`              // certificate subject
	ValidationType       string   `json:"validationType"`       // E for Extended Validation certificates; may be nil if unable to determine
}

type ChainCert struct {
	CRLRevocationStatus  int    `json:"crlRevocationStatus"`  // same as revocationStatus, but only for the CRL information (if any)
	IssuerLabel          string `json:"issuerLabel"`          // issuer name
	IssuerSubject        string `json:"issuerSubject"`        // issuer subject
	Issues               int    `json:"issues"`               // list of certificate issues, one bit per issue
	KeyAlg               string `json:"keyAlg"`               // key algorithm
	KeySize              int    `json:"keySize"`              // key size, in bits appropriate for the key algorithm
	KeyStrength          int    `json:"keyStrength"`          // key strength, in equivalent RSA bits
	Label                string `json:"label"`                // certificate label (user-friendly name)
	NotAfter             int64  `json:"notAfter"`             // timestamp after which the certificate is not valid
	NotBefore            int64  `json:"notBefore"`            // timestamp before which the certificate is not valid
	OCSPRevocationStatus int    `json:"ocspRevocationStatus"` // same as revocationStatus, but only for the OCSP information (if any)
	PINSHA256            string `json:"pinSha256"`            // -
	Raw                  string `json:"raw"`                  // Raw certificate data
	RevocationStatus     int    `json:"revocationStatus"`     // a number that describes the revocation status of the certificate
	SHA1Hash             string `json:"sha1Hash"`             // -
	SigAlg               string `json:"sigAlg"`               // certificate signature algorithm
	Subject              string `json:"subject"`              // certificate subject
}

type Protocol struct {
	ID               int    `json:"id"`               // protocol version number, e.g. 0x0303 for TLS 1.2
	Name             string `json:"name"`             // protocol name, i.e. SSL or TLS
	Q                *int   `json:"q"`                // 0 if the protocol is insecure, null otherwise
	V2SuitesDisabled bool   `json:"v2SuitesDisabled"` // some servers have SSLv2 protocol enabled, but with all SSLv2 cipher suites disabled
	Version          string `json:"version"`          // protocol version, e.g. 1.2 (for TLS)
}

type Suites struct {
	List       []*Suite `json:"list"`
	Preference bool     `json:"preference"`
}

type Suite struct {
	CipherStrength int    `json:"cipherStrength"` // suite strength (e.g., 128)
	DHG            int    `json:"dhG"`            // DH params, g component
	DHP            int    `json:"dhP"`            // DH params, p component
	DHStrength     int    `json:"dhStrength"`     // strength of DH params (e.g., 1024)
	DHYs           int    `json:"dhYs"`           // DH params, Ys component
	ECDHBits       int    `json:"ecdhBits"`       // ECDH bits
	ECDHStrength   int    `json:"ecdhStrength"`   // ECDH RSA-equivalent strength
	ID             int    `json:"id"`             // suite RFC ID (e.g., 5)
	Name           string `json:"name"`           // suite name (e.g., TLS_RSA_WITH_RC4_128_SHA)
	Q              *int   `json:"q"`              // 0 if the suite is insecure, null otherwise
}

type SIMS struct {
	Results []*SIM `json:"results"`
}

type SIM struct {
	Attempts   int        `json:"attempts"`   // always 1 with the current implementation
	Client     *SimClient `json:"client"`     // instance of SimClient
	ErrorCode  int        `json:"errorCode"`  // zero if handshake was successful, 1 if it was not
	KXInfo     string     `json:"kxInfo"`     // key exchange info
	ProtocolID int        `json:"protocolId"` // Negotiated protocol ID
	SuiteID    int        `json:"suiteId"`    // Negotiated suite ID
}

type SimClient struct {
	ID          int    `json:"id"`          // unique client ID
	IsReference bool   `json:"isReference"` // true if the browser is considered representative of modern browsers, false otherwise
	Name        string `json:"name"`        // some text
	Platform    string `json:"platform"`    // some text
	Version     string `json:"version"`     // some text
}

type HSTSPolicy struct {
	Directives        map[string]string `json:"directives"`        // list of raw policy directives
	Error             string            `json:"error"`             // error message when error is encountered, null otherwise
	Header            string            `json:"header"`            // the contents of the HSTS response header, if present
	IncludeSubDomains bool              `json:"includeSubDomains"` // true if the includeSubDomains directive is set; null otherwise
	LongMaxAge        int               `json:"LONG_MAX_AGE"`      // this constant contains what SSL Labs considers to be sufficiently large max-age value
	MaxAge            int64             `json:"maxAge"`            // the max-age value specified in the policy; null if policy is missing or invalid or on parsing error
	Preload           bool              `json:"preload"`           // true if the preload directive is set; null otherwise
	Status            string            `json:"status"`            // HSTS status
}

type HSTSPreload struct {
	Hostname   string `json:"hostname"`   // host name
	Source     string `json:"source"`     // source name
	SourceTime int64  `json:"sourceTime"` // time, as a Unix timestamp, when the preload database was retrieved
	Status     string `json:"status"`     // preload status
}

type HPKPPolicy struct {
	Directives        []Directive `json:"directives"`        // list of raw policy directives
	Header            string      `json:"header"`            // the contents of the HPKP response header, if present
	IncludeSubDomains bool        `json:"includeSubDomains"` // true if the includeSubDomains directive is set; null otherwise
	MatchedPins       []Pin       `json:"matchedPins"`       // list of pins that match the current configuration
	MaxAge            int64       `json:"maxAge"`            // the max-age value from the policy
	Pins              []Pin       `json:"pins"`              // list of all pins used by the policy
	ReportURI         string      `json:"reportUri"`         // the report-uri value from the policy
	Status            string      `json:"status"`            // HPKP status
}

type Pin struct {
	HashFunction string `json:"hashFunction"`
	Value        string `json:"value"`
}

type Directive struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type DrownHost struct {
	Export  bool   `json:"export"`  // true if export cipher suites detected
	IP      string `json:"ip"`      // Ip address of server that shares same RSA-Key/hostname in its certificate
	Port    int    `json:"port"`    // port number of the server
	Special bool   `json:"special"` // true if vulnerable OpenSSL version detected
	SSLv2   bool   `json:"sslv2"`   // true if SSL v2 is supported
	Status  string `json:"status"`  // drown host status
}

// RequestTimeout is request timeout in seconds
var RequestTimeout = 10.0

// ////////////////////////////////////////////////////////////////////////////////// //

// NewAPI create new api struct
func NewAPI(app, version string) (*API, error) {
	if app == "" {
		return nil, fmt.Errorf("App name can't be empty")
	}

	api := &API{
		Client: &fasthttp.Client{
			Name:                getUserAgent(app, version),
			MaxIdleConnDuration: 5 * time.Second,
			ReadTimeout:         time.Duration(RequestTimeout) * time.Second,
			WriteTimeout:        time.Duration(RequestTimeout) * time.Second,
			MaxConnsPerHost:     100,
		},
	}

	info := &Info{}
	err := api.doRequest(API_URL_INFO, info)

	if err != nil {
		return nil, err
	}

	api.Info = info

	return api, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Analyze start check for host
func (api *API) Analyze(host string, params AnalyzeParams) (*AnalyzeProgress, error) {
	progress := &AnalyzeProgress{host: host, api: api}
	query := "host=" + host
	query += "&" + paramsToQuery(params)

	err := api.doRequest(API_URL_ANALYZE+"?"+query, nil)

	if err != nil {
		return nil, err
	}

	return progress, nil
}

// Info return short info
func (ap *AnalyzeProgress) Info() (*AnalyzeInfo, error) {
	query := "host=" + ap.host

	info := &AnalyzeInfo{}
	err := ap.api.doRequest(API_URL_ANALYZE+"?"+query, info)

	if err != nil {
		return nil, err
	}

	ap.prevStatus = info.Status

	return info, nil
}

// DetailedInfo return detailed endpoint info
func (ap *AnalyzeProgress) DetailedInfo(ip string) (*EndpointInfo, error) {
	var err error

	if ap.prevStatus != STATUS_READY {
		_, err = ap.Info()

		if err != nil {
			return nil, err
		}

		if ap.prevStatus != STATUS_READY {
			return nil, fmt.Errorf("Retrieving detailed information possible only with status READY")
		}
	}

	query := "host=" + ap.host + "&s=" + ip
	info := &EndpointInfo{}

	err = ap.api.doRequest(API_URL_DETAILED+"?"+query, info)

	if err != nil {
		return nil, err
	}

	return info, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// doRequest sends request through http client
func (api *API) doRequest(uri string, result interface{}) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	req.SetRequestURI(uri)

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	err := api.Client.Do(req, resp)

	if err != nil {
		return err
	}

	statusCode := resp.StatusCode()

	if statusCode != 200 {
		return fmt.Errorf("API return HTTP code %d", statusCode)
	}

	if result == nil {
		return nil
	}

	err = json.Unmarshal(resp.Body(), result)

	return err
}

// ////////////////////////////////////////////////////////////////////////////////// //

// paramsToQuery is a lightweight query encoder
func paramsToQuery(params AnalyzeParams) string {
	var result string

	if params.Public {
		result += "publish=on&"
	}

	if params.StartNew {
		result += "startNew=on&"
	}

	if params.FromCache {
		result += "fromCache=on&"
	}

	if params.MaxAge != 0 {
		result += "maxAge=" + fmt.Sprintf("%d", params.MaxAge) + "&"
	}

	if params.All {
		result += "all=on&"
	}

	if params.IgnoreMismatch {
		result += "ignoreMismatch=on&"
	}

	if len(result) != 0 {
		return result[:len(result)-1]
	}

	return ""
}

// getUserAgent generate user-agent string for client
func getUserAgent(app, version string) string {
	if app != "" && version != "" {
		return fmt.Sprintf(
			"%s/%s SSLScan/%s (go; %s; %s-%s)",
			app, version, VERSION, runtime.Version(),
			runtime.GOARCH, runtime.GOOS,
		)
	}

	return fmt.Sprintf(
		"SSLScan/%s (go; %s; %s-%s)",
		VERSION, runtime.Version(),
		runtime.GOARCH, runtime.GOOS,
	)
}
