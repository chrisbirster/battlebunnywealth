package httpserver

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/game"
	"github.com/chrisbirster/battlebunnywealth/internal/identity"
	"github.com/chrisbirster/battlebunnywealth/internal/proofofplay"
)

type proofOfPlay interface {
	Status(time.Time) proofofplay.NetworkStatus
	Head() proofofplay.Block
	CurrentEpoch(time.Time) uint64
}

const sessionCookieName = "bbw_session"

func New(logger *slog.Logger, protocol proofOfPlay, authority *proofofplay.AuthorityService, games *game.Registry, auth *identity.Service, spa http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/healthz", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"status":"ok","service":"battle-bunny-wealth","version":"dev"}) })
	mux.HandleFunc("GET /api/v1/proof-of-play", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, protocol.Status(time.Now().UTC())) })
	mux.HandleFunc("GET /api/v1/proof-of-play/blocks/head", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, protocol.Head()) })
	mux.HandleFunc("GET /api/v1/proof-of-play/me", func(w http.ResponseWriter, r *http.Request) {
		account, err := currentAccount(auth, r)
		if err != nil { writeIdentityError(w, err); return }
		snapshot, err := authority.Snapshot(account.ID)
		if err != nil { writeProofError(w, err); return }
		writeJSON(w, http.StatusOK, snapshot)
	})
	mux.HandleFunc("POST /api/v1/proof-of-play/missions", func(w http.ResponseWriter, r *http.Request) {
		account, err := currentAccount(auth, r)
		if err != nil { writeIdentityError(w, err); return }
		var body struct { DeviceID string `json:"deviceId"` }
		if err := decodeJSON(w, r, &body); err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error":err.Error()}); return }
		device, err := activeDevice(account, body.DeviceID)
		if err != nil { writeIdentityError(w, err); return }
		now := time.Now().UTC()
		snapshot, err := authority.IssueMission(account.ID, device.ID, device.PublicKeySPKI, device.AttestationStatus, protocol.Head().Hash, protocol.CurrentEpoch(now))
		if err != nil { writeProofError(w, err); return }
		writeJSON(w, http.StatusOK, snapshot)
	})
	mux.HandleFunc("POST /api/v1/proof-of-play/missions/{id}/complete", func(w http.ResponseWriter, r *http.Request) {
		account, err := currentAccount(auth, r)
		if err != nil { writeIdentityError(w, err); return }
		var body struct { DeviceID string `json:"deviceId"`; Signature string `json:"signature"` }
		if err := decodeJSON(w, r, &body); err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error":err.Error()}); return }
		device, err := activeDevice(account, body.DeviceID)
		if err != nil { writeIdentityError(w, err); return }
		snapshot, err := authority.CompleteMission(account.ID, device.ID, device.PublicKeySPKI, r.PathValue("id"), body.Signature)
		if err != nil { writeProofError(w, err); return }
		writeJSON(w, http.StatusOK, snapshot)
	})

	mux.HandleFunc("GET /api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) { account, err := currentAccount(auth, r); if err != nil { writeIdentityError(w, err); return }; writeJSON(w, http.StatusOK, account) })
	mux.HandleFunc("POST /api/v1/auth/passkey/register/begin", func(w http.ResponseWriter, r *http.Request) { var body struct { DisplayName string `json:"displayName"` }; if err := decodeJSON(w,r,&body); err != nil { writeJSON(w,http.StatusBadRequest,map[string]string{"error":err.Error()}); return }; accountID := ""; if account, err := currentAccount(auth,r); err == nil { accountID = account.ID }; result, err := auth.BeginRegistration(requestOrigin(r), body.DisplayName, accountID); if err != nil { writeIdentityError(w,err); return }; writeJSON(w,http.StatusOK,result) })
	mux.HandleFunc("POST /api/v1/auth/passkey/register/finish", func(w http.ResponseWriter, r *http.Request) { var body struct { Ceremony string `json:"ceremony"`; Credential identity.RegistrationCredential `json:"credential"` }; if err:=decodeJSON(w,r,&body); err!=nil { writeJSON(w,http.StatusBadRequest,map[string]string{"error":err.Error()}); return }; account, token, err := auth.FinishRegistration(body.Ceremony,body.Credential); if err!=nil { writeIdentityError(w,err); return }; setSessionCookie(w,r,token,30*24*time.Hour); writeJSON(w,http.StatusOK,account) })
	mux.HandleFunc("POST /api/v1/auth/passkey/login/begin", func(w http.ResponseWriter, r *http.Request) { result,err:=auth.BeginLogin(requestOrigin(r)); if err!=nil {writeIdentityError(w,err);return}; writeJSON(w,http.StatusOK,result) })
	mux.HandleFunc("POST /api/v1/auth/passkey/login/finish", func(w http.ResponseWriter, r *http.Request) { var body struct { Ceremony string `json:"ceremony"`; Credential identity.AssertionCredential `json:"credential"` }; if err:=decodeJSON(w,r,&body); err!=nil {writeJSON(w,http.StatusBadRequest,map[string]string{"error":err.Error()});return}; account,token,err:=auth.FinishLogin(body.Ceremony,body.Credential); if err!=nil {writeIdentityError(w,err);return}; setSessionCookie(w,r,token,30*24*time.Hour); writeJSON(w,http.StatusOK,account) })
	mux.HandleFunc("POST /api/v1/auth/logout", func(w http.ResponseWriter, r *http.Request) { token:=sessionToken(r); _=auth.Logout(token); clearSessionCookie(w,r); writeJSON(w,http.StatusOK,map[string]bool{"ok":true}) })
	mux.HandleFunc("PUT /api/v1/account/atproto", func(w http.ResponseWriter, r *http.Request) { account,err:=currentAccount(auth,r); if err!=nil{writeIdentityError(w,err);return}; var body struct{DID string `json:"did"`}; if err:=decodeJSON(w,r,&body);err!=nil{writeJSON(w,http.StatusBadRequest,map[string]string{"error":err.Error()});return}; updated,err:=auth.LinkATProto(r.Context(),account.ID,body.DID);if err!=nil{writeIdentityError(w,err);return};writeJSON(w,http.StatusOK,updated) })
	mux.HandleFunc("DELETE /api/v1/account/atproto", func(w http.ResponseWriter, r *http.Request) { account,err:=currentAccount(auth,r);if err!=nil{writeIdentityError(w,err);return};updated,err:=auth.UnlinkATProto(account.ID);if err!=nil{writeIdentityError(w,err);return};writeJSON(w,http.StatusOK,updated) })
	mux.HandleFunc("POST /api/v1/account/devices", func(w http.ResponseWriter, r *http.Request) { account,err:=currentAccount(auth,r);if err!=nil{writeIdentityError(w,err);return};var body struct{Name string `json:"name"`;Platform string `json:"platform"`;PublicKeySPKI string `json:"publicKeySpki"`};if err:=decodeJSON(w,r,&body);err!=nil{writeJSON(w,http.StatusBadRequest,map[string]string{"error":err.Error()});return};updated,err:=auth.EnrollDevice(account.ID,body.Name,body.Platform,body.PublicKeySPKI);if err!=nil{writeIdentityError(w,err);return};writeJSON(w,http.StatusOK,updated) })
	mux.HandleFunc("POST /api/v1/account/devices/{id}/revoke", func(w http.ResponseWriter, r *http.Request) { account,err:=currentAccount(auth,r);if err!=nil{writeIdentityError(w,err);return};updated,err:=auth.RevokeDevice(account.ID,r.PathValue("id"));if err!=nil{writeIdentityError(w,err);return};writeJSON(w,http.StatusOK,updated) })

	withGame := func(fn func(http.ResponseWriter,*http.Request,*game.Service)) http.HandlerFunc { return func(w http.ResponseWriter,r *http.Request){ account,err:=currentAccount(auth,r);if err!=nil{writeIdentityError(w,err);return};service,err:=games.ForAccount(account.ID);if err!=nil{writeGameError(w,err);return};fn(w,r,service) } }
	mux.HandleFunc("GET /api/v1/game/state", withGame(func(w http.ResponseWriter,_ *http.Request,s *game.Service){ snapshot,err:=s.Snapshot();if err!=nil{writeGameError(w,err);return};writeJSON(w,http.StatusOK,snapshot) }))
	mux.HandleFunc("PUT /api/v1/game/profile", withGame(func(w http.ResponseWriter,r *http.Request,s *game.Service){var profile game.PlayerProfile;if err:=decodeJSON(w,r,&profile);err!=nil{writeJSON(w,http.StatusBadRequest,map[string]string{"error":err.Error()});return};snapshot,err:=s.UpdateProfile(profile);if err!=nil{writeGameError(w,err);return};writeJSON(w,http.StatusOK,snapshot)}))
	mux.HandleFunc("POST /api/v1/game/businesses/{id}/upgrade", withGame(func(w http.ResponseWriter,r *http.Request,s *game.Service){snapshot,err:=s.UpgradeBusiness(r.PathValue("id"));if err!=nil{writeGameError(w,err);return};writeJSON(w,http.StatusOK,snapshot)}))
	mux.HandleFunc("POST /api/v1/game/onboarding/advance", withGame(func(w http.ResponseWriter,_ *http.Request,s *game.Service){snapshot,err:=s.AdvanceOnboarding();if err!=nil{writeGameError(w,err);return};writeJSON(w,http.StatusOK,snapshot)}))
	mux.HandleFunc("POST /api/v1/game/story/advance", withGame(func(w http.ResponseWriter,_ *http.Request,s *game.Service){snapshot,err:=s.AdvanceStory();if err!=nil{writeGameError(w,err);return};writeJSON(w,http.StatusOK,snapshot)}))
	mux.HandleFunc("PUT /api/v1/game/warren", withGame(func(w http.ResponseWriter,r *http.Request,s *game.Service){var body struct{Theme string `json:"theme"`};if err:=decodeJSON(w,r,&body);err!=nil{writeJSON(w,http.StatusBadRequest,map[string]string{"error":err.Error()});return};snapshot,err:=s.UpdateWarren(body.Theme);if err!=nil{writeGameError(w,err);return};writeJSON(w,http.StatusOK,snapshot)}))
	mux.HandleFunc("POST /api/v1/game/season/turn-in", withGame(func(w http.ResponseWriter,_ *http.Request,s *game.Service){snapshot,err:=s.TurnInSeason();if err!=nil{writeGameError(w,err);return};writeJSON(w,http.StatusOK,snapshot)}))
	mux.HandleFunc("GET /api/v1/game/standings", func(w http.ResponseWriter,r *http.Request){if _,err:=currentAccount(auth,r);err!=nil{writeIdentityError(w,err);return};standings,err:=games.Standings();if err!=nil{writeGameError(w,err);return};writeJSON(w,http.StatusOK,standings)})
	mux.HandleFunc("POST /api/v1/game/dev/season/advance", withGame(func(w http.ResponseWriter,_ *http.Request,s *game.Service){if os.Getenv("BBWEALTH_DEV_CONTROLS")!="1"{http.Error(w,"not found",http.StatusNotFound);return};snapshot,err:=s.AdvanceSeasonForDevelopment();if err!=nil{writeGameError(w,err);return};writeJSON(w,http.StatusOK,snapshot)}))

	mux.Handle("/", spa)
	return requestLog(logger, securityHeaders(originGuard(auth,mux)))
}

func activeDevice(account identity.AccountView, id string) (identity.Device, error) {
	for _, device := range account.Devices {
		if device.ID == id && device.Status == identity.DeviceStatusActive {
			return device, nil
		}
	}
	return identity.Device{}, identity.ErrUnknownDevice
}

func currentAccount(auth *identity.Service, r *http.Request) (identity.AccountView,error) { return auth.Authenticate(sessionToken(r)) }
func sessionToken(r *http.Request) string { cookie,err:=r.Cookie(sessionCookieName);if err!=nil{return ""};return cookie.Value }
func setSessionCookie(w http.ResponseWriter,r *http.Request,token string,ttl time.Duration){http.SetCookie(w,&http.Cookie{Name:sessionCookieName,Value:token,Path:"/",HttpOnly:true,Secure:strings.HasPrefix(requestOrigin(r),"https://"),SameSite:http.SameSiteStrictMode,MaxAge:int(ttl/time.Second)})}
func clearSessionCookie(w http.ResponseWriter,r *http.Request){http.SetCookie(w,&http.Cookie{Name:sessionCookieName,Value:"",Path:"/",HttpOnly:true,Secure:strings.HasPrefix(requestOrigin(r),"https://"),SameSite:http.SameSiteStrictMode,MaxAge:-1})}
func requestOrigin(r *http.Request) string { if origin:=strings.TrimRight(r.Header.Get("Origin"),"/");origin!=""{return origin};scheme:="http";if r.TLS!=nil{scheme="https"};return scheme+"://"+r.Host }
func originGuard(auth *identity.Service,next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){if r.Method!=http.MethodGet&&r.Method!=http.MethodHead&&r.Method!=http.MethodOptions{if origin:=r.Header.Get("Origin");origin!=""&&!auth.AllowedOrigin(origin){writeJSON(w,http.StatusForbidden,map[string]string{"error":"origin not allowed"});return}};next.ServeHTTP(w,r)}) }
func decodeJSON(w http.ResponseWriter,r *http.Request,value any) error{r.Body=http.MaxBytesReader(w,r.Body,64<<10);decoder:=json.NewDecoder(r.Body);decoder.DisallowUnknownFields();return decoder.Decode(value)}
func writeIdentityError(w http.ResponseWriter,err error){status:=http.StatusInternalServerError;switch{case errors.Is(err,identity.ErrUnauthorized):status=http.StatusUnauthorized;case errors.Is(err,identity.ErrInvalidCeremony),errors.Is(err,identity.ErrInvalidCredential),errors.Is(err,identity.ErrInvalidDisplayName),errors.Is(err,identity.ErrInvalidDevice),errors.Is(err,identity.ErrInvalidDID),errors.Is(err,identity.ErrUnsupportedDIDMethod):status=http.StatusBadRequest;case errors.Is(err,identity.ErrCredentialExists),errors.Is(err,identity.ErrCounterRollback):status=http.StatusConflict;case errors.Is(err,identity.ErrUnknownDevice),errors.Is(err,identity.ErrUnknownCredential):status=http.StatusNotFound};message:=err.Error();if status==http.StatusInternalServerError{message="internal server error"};writeJSON(w,status,map[string]string{"error":message})}
func writeProofError(w http.ResponseWriter,err error){status:=http.StatusInternalServerError;switch{case errors.Is(err,proofofplay.ErrMissionLimit):status=http.StatusTooManyRequests;case errors.Is(err,proofofplay.ErrMissionNotFound):status=http.StatusNotFound;case errors.Is(err,proofofplay.ErrMissionExpired):status=http.StatusGone;case errors.Is(err,proofofplay.ErrMissionCompleted):status=http.StatusConflict;case errors.Is(err,proofofplay.ErrMissionBinding),errors.Is(err,proofofplay.ErrInvalidMissionSignature),errors.Is(err,proofofplay.ErrInvalidMissionDevice):status=http.StatusBadRequest};message:=err.Error();if status==http.StatusInternalServerError{message="internal server error"};writeJSON(w,status,map[string]string{"error":message})}
func writeGameError(w http.ResponseWriter,err error){status:=http.StatusInternalServerError;switch{case errors.Is(err,game.ErrInvalidProfile),errors.Is(err,game.ErrInvalidWarren):status=http.StatusBadRequest;case errors.Is(err,game.ErrUnknownBusiness):status=http.StatusNotFound;case errors.Is(err,game.ErrInsufficientFunds),errors.Is(err,game.ErrTurnInLocked),errors.Is(err,game.ErrStoryComplete):status=http.StatusConflict};message:=err.Error();if status==http.StatusInternalServerError{message="internal server error"};writeJSON(w,status,map[string]string{"error":message})}
func writeJSON(w http.ResponseWriter,status int,value any){w.Header().Set("Content-Type","application/json; charset=utf-8");w.WriteHeader(status);_ = json.NewEncoder(w).Encode(value)}
func securityHeaders(next http.Handler) http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.Header().Set("X-Content-Type-Options","nosniff");w.Header().Set("Referrer-Policy","strict-origin-when-cross-origin");w.Header().Set("X-Frame-Options","DENY");next.ServeHTTP(w,r)})}
func requestLog(logger *slog.Logger,next http.Handler) http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){started:=time.Now();next.ServeHTTP(w,r);logger.Info("http request","method",r.Method,"path",r.URL.Path,"duration",time.Since(started))})}
