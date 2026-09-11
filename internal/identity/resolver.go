package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidDID           = errors.New("invalid ATProto DID")
	ErrUnsupportedDIDMethod = errors.New("unsupported DID method")
)

type ResolvedDID struct { DID string; Handle string; PDS string }
type DIDResolver interface { Resolve(context.Context, string) (ResolvedDID, error) }
type PLCResolver struct { client *http.Client }

func NewPLCResolver(client *http.Client) *PLCResolver { if client == nil { client = &http.Client{Timeout:5*time.Second} }; return &PLCResolver{client:client} }
var plcPattern = regexp.MustCompile(`^did:plc:[a-z2-7]{24}$`)

func (r *PLCResolver) Resolve(ctx context.Context, did string) (ResolvedDID, error) {
	did = strings.TrimSpace(did)
	if !plcPattern.MatchString(did) { if strings.HasPrefix(did,"did:web:"){return ResolvedDID{},ErrUnsupportedDIDMethod}; return ResolvedDID{},ErrInvalidDID }
	req,err:=http.NewRequestWithContext(ctx,http.MethodGet,"https://plc.directory/"+did,nil);if err!=nil{return ResolvedDID{},err}
	resp,err:=r.client.Do(req);if err!=nil{return ResolvedDID{},fmt.Errorf("resolve DID: %w",err)};defer resp.Body.Close();if resp.StatusCode!=http.StatusOK{return ResolvedDID{},fmt.Errorf("resolve DID: HTTP %d",resp.StatusCode)}
	var doc struct{ID string `json:"id"`;AlsoKnownAs []string `json:"alsoKnownAs"`;Service []struct{Type string `json:"type"`;ServiceEndpoint string `json:"serviceEndpoint"`} `json:"service"`}
	if err:=json.NewDecoder(io.LimitReader(resp.Body,256<<10)).Decode(&doc);err!=nil{return ResolvedDID{},fmt.Errorf("decode DID: %w",err)};if doc.ID!=did{return ResolvedDID{},ErrInvalidDID}
	resolved:=ResolvedDID{DID:did};for _,aka:=range doc.AlsoKnownAs{if strings.HasPrefix(aka,"at://")&&len(aka)>5{resolved.Handle=strings.TrimPrefix(aka,"at://");break}};for _,service:=range doc.Service{if service.Type=="AtprotoPersonalDataServer"{resolved.PDS=service.ServiceEndpoint;break}};return resolved,nil
}
