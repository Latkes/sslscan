// Package sslscan provides methods and structs for working with SSLLabs public API
package sslscan

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2019 ESSENTIAL KAOS                         //
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
	API_URL_INFO     = "https://api.ssllabs.com/api/v3/info"
	API_URL_ANALYZE  = "https://api.ssllabs.com/api/v3/analyze"
	API_URL_DETAILED = "https://api.ssllabs.com/api/v3/getEndpointData"
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
	TICKETBLEED_STATUS_FAILED         = -1
	TICKETBLEED_STATUS_UNKNOWN        = 0
	TICKETBLEED_STATUS_NOT_VULNERABLE = 1
	TICKETBLEED_STATUS_VULNERABLE     = 2
)

const (
	BLEICHENBACHER_STATUS_FAILED               = -1
	BLEICHENBACHER_STATUS_UNKNOWN              = 0
	BLEICHENBACHER_STATUS_NOT_VULNERABLE       = 1
	BLEICHENBACHER_STATUS_VULNERABLE_WEAK      = 2
	BLEICHENBACHER_STATUS_VULNERABLE_STRONG    = 3
	BLEICHENBACHER_STATUS_INCONSISTENT_RESULTS = 4
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
	SPKP_STATUS_UNKNOWN    = "unknown"
	SPKP_STATUS_ABSENT     = "absent"
	SPKP_STATUS_INVALID    = "invalid"
	SPKP_STATUS_INCOMPLETE = "incomplete"
	SPKP_STATUS_PARTIAL    = "partial"
	SPKP_STATUS_FORBIDDEN  = "forbidden"
	SPKP_STATUS_VALID      = "valid"
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

const (
	CERT_CHAIN_ISSUE_UNUSED = 1 << iota
	CERT_CHAIN_ISSUE_INCOMPLETE
	CERT_CHAIN_ISSUE_DUPLICATE
	CERT_CHAIN_ISSUE_INCORRECT_ORDER
	CERT_CHAIN_ISSUE_SELF_SIGNED_ROOT
	CERT_CHAIN_ISSUE_CANT_VALIDATE
)

const (
	PROTOCOL_SSL2  = 512
	PROTOCOL_SSL3  = 768
	PROTOCOL_TLS10 = 769
	PROTOCOL_TLS11 = 770
	PROTOCOL_TLS12 = 771
	PROTOCOL_TLS13 = 772
)

// VERSION is current package version
const VERSION = "12.0.0"

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
	IgnoreMismatch bool
}

type AnalyzeProgress struct {
	host       string
	prevStatus string

	maxAge int

	api *API
}

// DOCS: https://github.com/ssllabs/ssllabs-scan/blob/master/ssllabs-api-docs-v3.md

type Info struct {
	EngineVersion        string   `json:"engineVersion"`        // SSL Labs software version as a string (e.g., "1.11.14")
	CriteriaVersion      string   `json:"criteriaVersion"`      // rating criteria version as a string (e.g., "2009f")
	MaxAssessments       int      `json:"maxAssessments"`       // the maximum number of concurrent assessments the client is allowed to initiate
	CurrentAssessments   int      `json:"currentAssessments"`   // the number of ongoing assessments submitted by this client
	NewAssessmentCoolOff int      `json:"newAssessmentCoolOff"` // he cool-off period after each new assessment; you're not allowed to submit a new assessment before the cool-off expires, otherwise you'll get a 429
	Messages             []string `json:"messages"`             // a list of messages (strings). Messages can be public (sent to everyone) and private (sent only to the invoking client). Private messages are prefixed with "[Private]".
}

type AnalyzeInfo struct {
	Host            string          `json:"host"`            // assessment host, which can be a hostname or an IP address
	Port            int             `json:"port"`            // assessment port (e.g., 443)
	Protocol        string          `json:"protocol"`        // protocol (e.g., HTTP)
	IsPublic        bool            `json:"isPublic"`        // true if this assessment publicly available (listed on the SSL Labs assessment boards)
	Status          string          `json:"status"`          // assessment status; possible values: DNS, ERROR, IN_PROGRESS, and READY
	StatusMessage   string          `json:"statusMessage"`   // status message in English. When status is ERROR, this field will contain an error message
	StartTime       int64           `json:"startTime"`       // assessment starting time, in milliseconds since 1970
	TestTime        int64           `json:"testTime"`        // assessment completion time, in milliseconds since 1970
	EngineVersion   string          `json:"engineVersion"`   // assessment engine version (e.g., "1.0.120")
	CriteriaVersion string          `json:"criteriaVersion"` // grading criteria version (e.g., "2009")
	CacheExpiryTime int64           `json:"cacheExpiryTime"` // when will the assessment results expire from the cache
	CertHostnames   []string        `json:"certHostnames"`   // the list of certificate hostnames collected from the certificates seen during assessment
	Endpoints       []*EndpointInfo `json:"endpoints"`       // list of Endpoint structs
	Certs           []*Cert         `json:"certs"`           // a list of Cert structs, representing the chain certificates in the order in which they were retrieved from the server
}

type EndpointInfo struct {
	IPAdress             string           `json:"ipAddress"`            // endpoint IP address, in IPv4 or IPv6 format
	ServerName           string           `json:"serverName"`           // server name retrieved via reverse DNS
	StatusMessage        string           `json:"statusMessage"`        // assessment status message
	StatusDetails        string           `json:"statusDetails"`        // code of the operation currently in progress
	StatusDetailsMessage string           `json:"statusDetailsMessage"` // description of the operation currently in progress
	Grade                string           `json:"grade"`                // possible values: A+, A-, A-F, T (no trust) and M (certificate name mismatch)
	GradeTrustIgnored    string           `json:"gradeTrustIgnored"`    // grade (as above), if trust issues are ignored
	FutureGrade          string           `json:"futureGrade"`          // next grade because of upcoming grading criteria changes
	HasWarnings          bool             `json:"hasWarnings"`          // if this endpoint has warnings that might affect the score (e.g., get A- instead of A).
	IsExceptional        bool             `json:"isExceptional"`        // this flag will be raised when an exceptional configuration is encountered. The SSL Labs test will give such sites an A+
	Progress             int              `json:"progress"`             // assessment progress, which is a value from 0 to 100, and -1 if the assessment has not yet started
	Duration             int              `json:"duration"`             // assessment duration, in milliseconds
	ETA                  int              `json:"eta"`                  // estimated time, in seconds, until the completion of the assessment
	Delegation           int              `json:"delegation"`           // indicates domain name delegation with and without the www prefix
	Details              *EndpointDetails `json:"details"`              // this field contains an EndpointDetails struct. It's not present by default, but can be enabled by using the "all" paramerer to the analyze API call
}

type EndpointDetails struct {
	HostStartTime                  int64              `json:"hostStartTime"`                  // endpoint assessment starting time, in milliseconds since 1970. This field is useful when test results are retrieved in several HTTP invocations. Then, you should check that the hostStartTime value matches the startTime value of the host
	CertChains                     []*ChainCert       `json:"certChains"`                     // server Certificate chains
	Protocols                      []*Protocol        `json:"protocols"`                      // supported protocols
	Suites                         []*ProtocolSuites  `json:"suites"`                         // supported cipher suites
	NoSNISuites                    *ProtocolSuites    `json:"noSniSuites"`                    // cipher suites observed only with client that does not support Server Name Indication (SNI)
	NamedGroups                    *NamedGroups       `json:"namedGroups"`                    // instance of NamedGroups object
	ServerSignature                string             `json:"serverSignature"`                // contents of the HTTP Server response header when known
	PrefixDelegation               bool               `json:"prefixDelegation"`               // true if this endpoint is reachable via a hostname with the www prefix
	NonPrefixDelegation            bool               `json:"nonPrefixDelegation"`            // true if this endpoint is reachable via a hostname without the www prefix
	VulnBeast                      bool               `json:"vulnBeast"`                      // true if the endpoint is vulnerable to the BEAST attack
	RenegSupport                   int                `json:"renegSupport"`                   // this is an integer value that describes the endpoint support for renegotiation
	SessionResumption              int                `json:"sessionResumption"`              // this is an integer value that describes endpoint support for session resumption
	CompressionMethods             int                `json:"compressionMethods"`             // integer value that describes supported compression methods
	SupportsNPN                    bool               `json:"supportsNpn"`                    // true if the server supports NPN
	NPNProtocols                   string             `json:"npnProtocols"`                   // space separated list of supported protocols
	SupportsALPN                   bool               `json:"supportsAlpn"`                   // true if the server supports ALPN
	ALPNProtocols                  string             `json:"alpnProtocols"`                  // space separated list of supported ALPN protocols
	SessionTickets                 int                `json:"sessionTickets"`                 // indicates support for Session Tickets
	OCSPStapling                   bool               `json:"ocspStapling"`                   // true if OCSP stapling is deployed on the server
	StaplingRevocationStatus       int                `json:"staplingRevocationStatus"`       // same as Cert.revocationStatus, but for the stapled OCSP response
	StaplingRevocationErrorMessage string             `json:"staplingRevocationErrorMessage"` // description of the problem with the stapled OCSP response, if any
	SNIRequired                    bool               `json:"sniRequired"`                    // if SNI support is required to access the web site
	HTTPStatusCode                 int                `json:"httpStatusCode"`                 // status code of the final HTTP response seen
	HTTPForwarding                 string             `json:"httpForwarding"`                 // available on a server that responded with a redirection to some other hostname
	SupportsRC4                    bool               `json:"supportsRc4"`                    // supportsRc4
	RC4WithModern                  bool               `json:"rc4WithModern"`                  // true if RC4 is used with modern clients
	RC4Only                        bool               `json:"rc4Only"`                        // true if only RC4 suites are supported
	ForwardSecrecy                 int                `json:"forwardSecrecy"`                 // indicates support for Forward Secrecy
	SupportAEAD                    bool               `json:"supportsAead"`                   // true if the server supports at least one AEAD suite
	SupportsCBC                    bool               `json:"supportsCBC"`                    // true if the server supports at least one CBC suite
	ProtocolIntolerance            int                `json:"protocolIntolerance"`            // indicates protocol version intolerance issues
	MiscIntolerance                int                `json:"miscIntolerance"`                // indicates protocol version intolerance issues
	SIMS                           *SIMS              `json:"sims"`                           // sims
	Heartbleed                     bool               `json:"heartbleed"`                     // true if the server is vulnerable to the Heartbleed attack
	Heartbeat                      bool               `json:"heartbeat"`                      // true if the server supports the Heartbeat extension
	OpenSSLCCS                     int                `json:"openSslCcs"`                     // results of the CVE-2014-0224 test
	OpenSSLLuckyMinus20            int                `json:"openSSLLuckyMinus20"`            // results of the CVE-2016-2107 test
	Ticketbleed                    int                `json:"ticketbleed"`                    // results of the ticketbleed CVE-2016-9244 test
	Bleichenbacher                 int                `json:"bleichenbacher"`                 // results of the Return Of Bleichenbacher's Oracle Threat (ROBOT) test
	ZombiePoodle                   int                `json:"zombiePoodle"`                   // -
	GoldenDoodle                   int                `json:"goldenDoodle"`                   // -
	ZeroLengthPaddingOracle        int                `json:"zeroLengthPaddingOracle"`        // -
	SleepingPoodle                 int                `json:"sleepingPoodle"`                 // -
	Poodle                         bool               `json:"poodle"`                         // true if the endpoint is vulnerable to POODLE
	PoodleTLS                      int                `json:"poodleTls"`                      // results of the POODLE TLS test
	FallbackSCSV                   bool               `json:"fallbackScsv"`                   // true if the server supports TLS_FALLBACK_SCSV, false if it doesn't
	Freak                          bool               `json:"freak"`                          // true of the server is vulnerable to the FREAK attack
	HasSCT                         int                `json:"hasSct"`                         // information about the availability of certificate transparency information (embedded SCTs)
	DHPrimes                       []string           `json:"dhPrimes"`                       // list of hex-encoded DH primes used by the server
	DHUsesKnownPrimes              int                `json:"dhUsesKnownPrimes"`              // whether the server uses known DH primes
	DHYsReuse                      bool               `json:"dhYsReuse"`                      // true if the DH ephemeral server value is reused
	ECDHParameterReuse             bool               `json:"ecdhParameterReuse"`             // true if the server reuses its ECDHE values
	Logjam                         bool               `json:"logjam"`                         // true if the server uses DH parameters weaker than 1024 bits
	ChaCha20Preference             bool               `json:"chaCha20Preference"`             // true if the server takes into account client preferences when deciding if to use ChaCha20 suites
	HSTSPolicy                     *HSTSPolicy        `json:"hstsPolicy"`                     // server's HSTS policy
	HSTSPreloads                   []HSTSPreload      `json:"hstsPreloads"`                   // information about preloaded HSTS policies
	HPKPPolicy                     *HPKPPolicy        `json:"hpkpPolicy"`                     // server's HPKP policy
	HPKPRoPolicy                   *HPKPPolicy        `json:"hpkpRoPolicy"`                   // server's HPKP RO (Report Only) policy
	StaticPKPPolicy                *SPKPPolicy        `json:"staticPkpPolicy"`                // server's SPKP policy
	HTTPTransactions               []*HTTPTransaction `json:"httpTransactions"`               // an slice of HttpTransaction structs
	DrownHosts                     []DrownHost        `json:"drownHosts"`                     // list of drown hosts
	DrownErrors                    bool               `json:"drownErrors"`                    // true if error occurred in drown test
	DrownVulnerable                bool               `json:"drownVulnerable"`                // true if server vulnerable to drown attack
	ImplementsTLS13MandatoryCS     bool               `json:"implementsTLS13MandatoryCS"`     // true if server supports mandatory TLS 1.3 cipher suite (TLS_AES_128_GCM_SHA256), null if TLS 1.3 not supported
	ZeroRTTEnabled                 int                `json:"zeroRTTEnabled"`                 // results of the 0-RTT test
}

type Cert struct {
	ID                     string     `json:"id"`                     // certificate ID
	Subject                string     `json:"subject"`                // certificate subject
	SerialNumber           string     `json:"serialNumber"`           // certificate serial number (hex-encoded)
	CommonNames            []string   `json:"commonNames"`            // common names extracted from the subject
	AltNames               []string   `json:"altNames"`               // alternative names
	NotBefore              int64      `json:"notBefore"`              // timestamp before which the certificate is not valid
	NotAfter               int64      `json:"notAfter"`               // timestamp after which the certificate is not valid
	IssuerSubject          string     `json:"issuerSubject"`          // issuer subject
	SigAlg                 string     `json:"sigAlg"`                 // certificate signature algorithm
	RevocationInfo         int        `json:"revocationInfo"`         // a number that represents revocation information present in the certificate
	CRLURIs                []string   `json:"crlURIs"`                // CRL URIs extracted from the certificate
	OCSPURIs               []string   `json:"ocspURIs"`               // OCSP URIs extracted from the certificate
	RevocationStatus       int        `json:"revocationStatus"`       // a number that describes the revocation status of the certificate
	CRLRevocationStatus    int        `json:"crlRevocationStatus"`    // same as revocationStatus, but only for the CRL information (if any)
	OCSPRevocationStatus   int        `json:"ocspRevocationStatus"`   // same as revocationStatus, but only for the OCSP information (if any)
	DNSCAA                 bool       `json:"dnsCaa"`                 // true if CAA is supported else false
	CAAPolicy              *CAAPolicy `json:"caaPolicy"`              // CAA Policy
	MustStaple             bool       `json:"mustStaple"`             // true if stapling is supported else false
	SGC                    int        `json:"sgc"`                    // Server Gated Cryptography support
	ValidationType         string     `json:"validationType"`         // E for Extended Validation certificates; may be nil if unable to determine
	Issues                 int        `json:"issues"`                 // list of certificate issues, one bit per issue
	SCT                    bool       `json:"sct"`                    // true if the certificate contains an embedded SCT
	SHA1Hash               string     `json:"sha1Hash"`               // SHA1 hash of the certificate
	SHA256Hash             string     `json:"sha256Hash"`             // SHA256 hash of the certificate
	PINSHA256              string     `json:"pinSha256"`              // SHA256 hash of the public key
	KeyAlg                 string     `json:"keyAlg"`                 // key algorithm
	KeySize                int        `json:"keySize"`                // key size, in bits appropriate for the key algorithm
	KeyStrength            int        `json:"keyStrength"`            // key strength, in equivalent RSA bits
	KeyKnownDebianInsecure bool       `json:"keyKnownDebianInsecure"` // true if debian flaw is found, else false
	Raw                    string     `json:"raw"`                    // PEM-encoded certificate
}

type ChainCert struct {
	ID         string       `json:"id"`         // Certificate chain ID
	CertIDs    []string     `json:"certIds"`    // list of IDs of each certificate, representing the chain certificates in the order in which they were retrieved from the server
	TrustPaths []*TrustPath `json:"trustPaths"` // trust path object
	Issues     int          `json:"issues"`     // a number of flags that describe the chain and the problems it has
	NoSNI      bool         `json:"noSni"`      // true for certificate obtained only with No Server Name Indication (SNI)
}

type TrustPath struct {
	CertIDs []string      `json:"certIds"` // list of certificate ID from leaf to root
	Trust   []*TrustStore `json:"trust"`   // trust object. This object shows info about the trusted certificate by using Mozilla trust store
}

type TrustStore struct {
	RootStore         string `json:"rootStore"`         // this field shows the Trust store being used (eg. "Mozilla")
	IsTrusted         bool   `json:"isTrusted"`         // true if trusted against above rootStore
	TrustErrorMessage string `json:"trustErrorMessage"` // shows the error message if any
}

type NamedGroups struct {
	List       []NamedGroup `json:"list"`       // an slice of NamedGroup structs
	Preference bool         `json:"preference"` // true if the server has preferred curves that it uses first
}

type NamedGroup struct {
	ID             int    `json:"id"`             // named curve ID
	Name           string `json:"name"`           // named curve name
	Bits           int    `json:"bits"`           // named curve strength in EC bits
	NamedGroupType string `json:"namedGroupType"` // -
}

type Protocol struct {
	ID               int    `json:"id"`               // protocol version number, e.g. 0x0303 for TLS 1.2
	Name             string `json:"name"`             // protocol name, i.e. SSL or TLS
	Version          string `json:"version"`          // protocol version, e.g. 1.2 (for TLS)
	V2SuitesDisabled bool   `json:"v2SuitesDisabled"` // some servers have SSLv2 protocol enabled, but with all SSLv2 cipher suites disabled
	Q                *int   `json:"q"`                // 0 if the protocol is insecure, null otherwise
}

type ProtocolSuites struct {
	Protocol   int      `json:"protocol"`   // protocol version
	List       []*Suite `json:"list"`       // list of Suite structs
	Preference bool     `json:"preference"` // true if the server actively selects cipher suites
}

type Suite struct {
	ID             int    `json:"id"`             // suite RFC ID (e.g., 5)
	Name           string `json:"name"`           // suite name (e.g., TLS_RSA_WITH_RC4_128_SHA)
	CipherStrength int    `json:"cipherStrength"` // suite strength (e.g., 128)
	KxType         string `json:"kxType"`         // key exchange type (e.g., ECDH)
	KxStrength     int    `json:"kxStrength"`     // key exchange strength, in RSA-equivalent bits
	DHBits         int    `json:"dhBits"`         // strength of DH params (e.g., 1024)
	DHG            int    `json:"dhG"`            // DH params, g component
	DHP            int    `json:"dhP"`            // DH params, p component
	DHYs           int    `json:"dhYs"`           // DH params, Ys component
	NamedGroupBits int    `json:"namedGroupBits"` // EC bits
	NamedGroupID   int    `json:"namedGroupId"`   // EC curve ID
	NamedGroupName string `json:"namedGroupName"` // EC curve name
	Q              *int   `json:"q"`              // 0 if the suite is insecure, null otherwise
}

type SIMS struct {
	Results []*SIM `json:"results"`
}

type SIM struct {
	Client         *SimClient `json:"client"`         // instance of SimClient
	ErrorCode      int        `json:"errorCode"`      // zero if handshake was successful, 1 if it was not
	ErrorMessage   string     `json:"errorMessage"`   // error message if simulation has failed
	Attempts       int        `json:"attempts"`       // always 1 with the current implementation
	CertChainID    string     `json:"certChainId"`    // id of the certificate chain
	ProtocolID     int        `json:"protocolId"`     // negotiated protocol ID
	SuiteID        int        `json:"suiteId"`        // negotiated suite ID
	SuiteName      string     `json:"suiteName"`      // negotiated suite name
	KxType         string     `json:"kxType"`         // negotiated key exchange
	KxStrength     int        `json:"kxStrength"`     // negotiated key exchange strength, in RSA-equivalent bits
	DHBits         int        `json:"dhBits"`         // strength of DH params (e.g., 1024)
	DHG            int        `json:"dhG"`            // DH params, g component
	DHP            int        `json:"dhP"`            // DH params, p component
	DHYs           int        `json:"dhYs"`           // DH params, Ys component
	NamedGroupBits int        `json:"namedGroupBits"` // when ECDHE is negotiated, length of EC parameters
	NamedGroupID   int        `json:"namedGroupId"`   // when ECDHE is negotiated, EC curve ID
	NamedGroupName string     `json:"namedGroupName"` // when ECDHE is negotiated, EC curve nanme (e.g., "secp256r1")
	KeyAlg         string     `json:"keyAlg"`         // connection certificate key algorithms (e.g., "RSA")
	KeySize        int        `json:"keySize"`        // connection certificate key size (e.g., 2048)
	SigAlg         string     `json:"sigAlg"`         // connection certificate signature algorithm (e.g, "SHA256withRSA")
}

type SimClient struct {
	ID          int    `json:"id"`          // unique client ID
	Name        string `json:"name"`        // name of the client (e.g., Chrome)
	Platform    string `json:"platform"`    // name of the platform (e.g., XP SP3)
	Version     string `json:"version"`     // version of the software being simulated (e.g., 49)
	IsReference bool   `json:"isReference"` // true if the browser is considered representative of modern browsers, false otherwise
}

type HSTSPolicy struct {
	LongMaxAge        int               `json:"LONG_MAX_AGE"`      // this constant contains what SSL Labs considers to be sufficiently large max-age value
	Header            string            `json:"header"`            // the contents of the HSTS response header, if present
	Status            string            `json:"status"`            // HSTS status
	Error             string            `json:"error"`             // error message when error is encountered, null otherwise
	MaxAge            int64             `json:"maxAge"`            // the max-age value specified in the policy; null if policy is missing or invalid or on parsing error
	IncludeSubDomains bool              `json:"includeSubDomains"` // true if the includeSubDomains directive is set; null otherwise
	Preload           bool              `json:"preload"`           // true if the preload directive is set; null otherwise
	Directives        map[string]string `json:"directives"`        // list of raw policy directives
}

type HSTSPreload struct {
	Source     string `json:"source"`     // source name
	Hostname   string `json:"hostname"`   // host name
	Status     string `json:"status"`     // preload status
	Error      string `json:"error"`      // error message, when status is "error"
	SourceTime int64  `json:"sourceTime"` // time, as a Unix timestamp, when the preload database was retrieved
}

type HPKPPolicy struct {
	Header            string      `json:"header"`            // the contents of the HPKP response header, if present
	Status            string      `json:"status"`            // HPKP status
	Error             string      `json:"error"`             // error message, when the policy is invalid
	MaxAge            int64       `json:"maxAge"`            // the max-age value from the policy
	IncludeSubDomains bool        `json:"includeSubDomains"` // true if the includeSubDomains directive is set; null otherwise
	ReportURI         string      `json:"reportUri"`         // the report-uri value from the policy
	Pins              []Pin       `json:"pins"`              // list of all pins used by the policy
	MatchedPins       []Pin       `json:"matchedPins"`       // list of pins that match the current configuration
	Directives        []Directive `json:"directives"`        // list of raw policy directives
}

type SPKPPolicy struct {
	Status               string `json:"status"`               // SPKP status
	Error                string `json:"error"`                // error message, when the policy is invalid
	IncludeSubDomains    bool   `json:"includeSubDomains"`    // true if the includeSubDomains directive is set else false
	ReportURI            string `json:"reportUri"`            // the report-uri value from the policy
	PINs                 []Pin  `json:"pins"`                 // list of all pins used by the policy
	MatchedPINs          []Pin  `json:"matchedPins"`          // list of pins that match the current configuration
	ForbiddenPINs        []Pin  `json:"forbiddenPins"`        // list of all forbidden pins used by policy
	MatchedForbiddenPINs []Pin  `json:"matchedForbiddenPins"` // list of forbidden pins that match the current configuration
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

type CAAPolicy struct {
	PolicyHostname string      `json:"policyHostname"` // hostname where policy is located
	CAARecords     []CAARecord `json:"caaRecords"`     // list of Supported CAARecords
}

type CAARecord struct {
	Tag   string `json:"tag"`   // a property of the CAA record
	Value string `json:"value"` // corresponding value of a CAA property
	Flags int    `json:"flags"` // corresponding flags of CAA property
}

type HTTPTransaction struct {
	RequestURL         string       `json:"requestUrl"`         // request URL
	StatusCode         int          `json:"statusCode"`         // response status code
	RequestLine        string       `json:"requestLine"`        // the entire request line as a single field
	RequestHeaders     []string     `json:"requestHeaders"`     // a slice of request HTTP headers
	ResponseLine       string       `json:"responseLine"`       // the entire response line as a single field
	ResponseHeadersRaw []string     `json:"responseHeadersRaw"` // all response headers as a single field (useful if the headers are malformed)
	ResponseHeaders    []HTTPHeader `json:"responseHeaders"`    // a slice of response HTTP headers
	FragileServer      bool         `json:"fragileServer"`      // true if the server crashes when inspected by SSL Labs
}

type HTTPHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
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
	progress := &AnalyzeProgress{host: host, api: api, maxAge: params.MaxAge}
	query := "host=" + host
	query += "&" + paramsToQuery(params)

	err := api.doRequest(API_URL_ANALYZE+"?"+query, nil)

	if err != nil {
		return nil, err
	}

	return progress, nil
}

// Info return short info
func (ap *AnalyzeProgress) Info(detailed, fromCache bool) (*AnalyzeInfo, error) {
	query := "host=" + ap.host

	if detailed {
		query += "&all=on"
	}

	if fromCache {
		query += "&fromCache=on"

		if ap.maxAge > 0 {
			query += "&maxAge=" + fmt.Sprintf("%d", ap.maxAge)
		}
	}

	info := &AnalyzeInfo{}
	err := ap.api.doRequest(API_URL_ANALYZE+"?"+query, info)

	if err != nil {
		return nil, err
	}

	ap.prevStatus = info.Status

	return info, nil
}

// GetEndpointInfo returns detailed endpoint info
func (ap *AnalyzeProgress) GetEndpointInfo(ip string, fromCache bool) (*EndpointInfo, error) {
	var err error

	if ap.prevStatus != STATUS_READY {
		_, err = ap.Info(false, false)

		if err != nil {
			return nil, err
		}

		if ap.prevStatus != STATUS_READY {
			return nil, fmt.Errorf("Retrieving detailed information possible only with status READY")
		}
	}

	query := "host=" + ap.host + "&s=" + ip

	if fromCache {
		query += "&fromCache=on"

		if ap.maxAge > 0 {
			query += "&maxAge=" + fmt.Sprintf("%d", ap.maxAge)
		}
	}

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
