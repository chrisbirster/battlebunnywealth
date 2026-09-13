package testnet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HTTPNode struct {
	Engine *Engine
	Guard *PeerGuard
	Store *Store
	Relayer *Relayer
	DraftOperations func() []Operation
	OnFinalized func([]Operation)
	cfg ProtocolConfig
	mu sync.Mutex
	inFlight map[string]int
	rejected uint64
	accepted uint64
}
func NewHTTPNode(engine *Engine,guard *PeerGuard,store *Store)*HTTPNode{return &HTTPNode{Engine:engine,Guard:guard,Store:store,cfg:engine.genesis.Config,inFlight:map[string]int{}}}
func(n *HTTPNode)ServeHTTP(w http.ResponseWriter,r *http.Request){if !n.enter(r){http.Error(w,"too many requests",http.StatusTooManyRequests);return};defer n.leave(r);w.Header().Set("X-Content-Type-Options","nosniff");switch{case r.Method==http.MethodGet&&r.URL.Path=="/v1/node/status":n.writeJSON(w,http.StatusOK,n.Engine.Status());case r.Method==http.MethodGet&&r.URL.Path=="/v1/node/genesis":n.writeJSON(w,http.StatusOK,n.Engine.Genesis());case r.Method==http.MethodGet&&r.URL.Path=="/v1/node/blocks":n.handleBlocks(w,r);case r.Method==http.MethodGet&&r.URL.Path=="/v1/node/metrics":n.handleMetrics(w);case r.Method==http.MethodGet&&r.URL.Path=="/v1/committee/draft":n.handleDraft(w,r);case r.Method==http.MethodPost&&r.URL.Path=="/v1/committee/proposals":n.handleDirectProposal(w,r);case r.Method==http.MethodPost&&r.URL.Path=="/v1/committee/votes":n.handleDirectVote(w,r);case r.Method==http.MethodPost&&r.URL.Path=="/v1/peer/proposals":n.handlePeer(w,r,"proposal");case r.Method==http.MethodPost&&r.URL.Path=="/v1/peer/votes":n.handlePeer(w,r,"vote");default:http.NotFound(w,r)}}
func(n *HTTPNode)enter(r *http.Request)bool{host,_,_:=net.SplitHostPort(r.RemoteAddr);if host==""{host=r.RemoteAddr};n.mu.Lock();defer n.mu.Unlock();if n.inFlight[host]>=n.cfg.MaxConcurrentRequestsPerIP{return false};n.inFlight[host]++;return true}
func(n *HTTPNode)leave(r *http.Request){host,_,_:=net.SplitHostPort(r.RemoteAddr);if host==""{host=r.RemoteAddr};n.mu.Lock();defer n.mu.Unlock();n.inFlight[host]--;if n.inFlight[host]<=0{delete(n.inFlight,host)}}
func(n *HTTPNode)decode(w http.ResponseWriter,r *http.Request,target any)bool{if ct:=r.Header.Get("Content-Type");!strings.HasPrefix(ct,"application/json"){http.Error(w,"content-type must be application/json",http.StatusUnsupportedMediaType);return false};r.Body=http.MaxBytesReader(w,r.Body,n.cfg.MaxRequestBodyBytes);dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields();if err:=dec.Decode(target);err!=nil{http.Error(w,"invalid json",http.StatusBadRequest);return false};if err:=dec.Decode(&struct{}{});err!=io.EOF{http.Error(w,"multiple json values",http.StatusBadRequest);return false};return true}
func(n *HTTPNode)handleDraft(w http.ResponseWriter,r *http.Request){validator:=r.URL.Query().Get("validator");round64,_:=strconv.ParseUint(r.URL.Query().Get("round"),10,32);var ops []Operation;if n.DraftOperations!=nil{ops=n.DraftOperations()};block,err:=n.Engine.Draft(uint32(round64),validator,ops,time.Now());if err!=nil{n.error(w,err);return};n.writeJSON(w,http.StatusOK,block)}
func(n *HTTPNode)handleDirectProposal(w http.ResponseWriter,r *http.Request){var p Proposal;if !n.decode(w,r,&p){return};if err:=n.Engine.HandleProposal(p);err!=nil{n.error(w,err);return};n.bump(true);if n.Relayer!=nil{_=n.Relayer.Broadcast(r.Context(),"proposal",p)};n.writeJSON(w,http.StatusAccepted,map[string]any{"accepted":true,"blockHash":p.Block.Hash})}
func(n *HTTPNode)persistFinalized()error{if n.Store!=nil{blocks:=n.Engine.Finalized();if err:=n.Store.Append(blocks[len(blocks)-1]);err!=nil{return err}};if n.OnFinalized!=nil{blocks:=n.Engine.Finalized();n.OnFinalized(blocks[len(blocks)-1].Block.Operations)};return nil}
func(n *HTTPNode)handleDirectVote(w http.ResponseWriter,r *http.Request){var v Vote;if !n.decode(w,r,&v){return};finalized,err:=n.Engine.HandleVote(v);if err!=nil{n.error(w,err);return};if finalized{if err:=n.persistFinalized();err!=nil{http.Error(w,"persist finalized block",http.StatusInternalServerError);return}};n.bump(true);if n.Relayer!=nil{_=n.Relayer.Broadcast(r.Context(),"vote",v)};n.writeJSON(w,http.StatusAccepted,map[string]any{"accepted":true,"finalized":finalized})}
func(n *HTTPNode)handlePeer(w http.ResponseWriter,r *http.Request,kind string){if n.Guard==nil{http.Error(w,"peer ingress disabled",http.StatusForbidden);return};var e PeerEnvelope;if !n.decode(w,r,&e){return};if e.Kind!=kind{n.error(w,errors.New("peer message kind mismatch"));return};if err:=n.Guard.Verify(e,time.Now().UTC());err!=nil{n.error(w,err);return};switch kind{case"proposal":var p Proposal;if json.Unmarshal(e.Payload,&p)!=nil{n.error(w,errors.New("invalid proposal payload"));return};if err:=n.Engine.HandleProposal(p);err!=nil{n.error(w,err);return};case"vote":var v Vote;if json.Unmarshal(e.Payload,&v)!=nil{n.error(w,errors.New("invalid vote payload"));return};finalized,err:=n.Engine.HandleVote(v);if err!=nil{n.error(w,err);return};if finalized{if err:=n.persistFinalized();err!=nil{http.Error(w,"persist finalized block",http.StatusInternalServerError);return}}};n.bump(true);n.writeJSON(w,http.StatusAccepted,map[string]any{"accepted":true})}
func(n *HTTPNode)handleBlocks(w http.ResponseWriter,r *http.Request){from,_:=strconv.ParseUint(r.URL.Query().Get("from"),10,64);if from<1{from=1};all:=n.Engine.Finalized();if from>uint64(len(all))+1{http.Error(w,"from height beyond head",http.StatusBadRequest);return};if from>uint64(len(all)){n.writeJSON(w,http.StatusOK,[]FinalizedBlock{});return};n.writeJSON(w,http.StatusOK,all[from-1:])}
func(n *HTTPNode)handleMetrics(w http.ResponseWriter){metrics:=map[string]any{"status":n.Engine.Status(),"acceptedMessages":n.accepted,"rejectedMessages":n.rejected};if n.Guard!=nil{metrics["peers"]=n.Guard.Metrics()};n.writeJSON(w,http.StatusOK,metrics)}
func(n *HTTPNode)bump(ok bool){n.mu.Lock();defer n.mu.Unlock();if ok{n.accepted++}else{n.rejected++}}
func(n *HTTPNode)error(w http.ResponseWriter,err error){n.bump(false);status:=http.StatusBadRequest;if errors.Is(err,ErrUnknownPeer)||errors.Is(err,ErrInvalidSignature){status=http.StatusForbidden};http.Error(w,err.Error(),status)}
func(n *HTTPNode)writeJSON(w http.ResponseWriter,status int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_=json.NewEncoder(w).Encode(v)}

func SyncFrom(ctxURL string,client *http.Client,engine *Engine)error{if client==nil{client=&http.Client{Timeout:10*time.Second}};base:=strings.TrimRight(ctxURL,"/");resp,err:=client.Get(base+"/v1/node/genesis");if err!=nil{return err};defer resp.Body.Close();if resp.StatusCode!=http.StatusOK{return fmt.Errorf("genesis status %d",resp.StatusCode)};var remote Genesis;if err:=json.NewDecoder(io.LimitReader(resp.Body,1<<20)).Decode(&remote);err!=nil{return err};if remote.Hash!=engine.Genesis().Hash{return errors.New("peer genesis mismatch")};from:=engine.Height()+1;resp2,err:=client.Get(fmt.Sprintf("%s/v1/node/blocks?from=%d",base,from));if err!=nil{return err};defer resp2.Body.Close();if resp2.StatusCode!=http.StatusOK{return fmt.Errorf("blocks status %d",resp2.StatusCode)};var blocks []FinalizedBlock;if err:=json.NewDecoder(io.LimitReader(resp2.Body,8<<20)).Decode(&blocks);err!=nil{return err};for _,block:=range blocks{if err:=engine.ImportFinalized(block);err!=nil{return err}};return nil}
