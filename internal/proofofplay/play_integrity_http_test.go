package proofofplay

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPlayIntegrityHTTPDecoder(t *testing.T){
	server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		if r.Header.Get("Authorization")!="Bearer token"{t.Fatalf("authorization=%q",r.Header.Get("Authorization"))}
		if r.URL.Path!="/v1/com.example.bbw:decodeIntegrityToken"{t.Fatalf("path=%s",r.URL.Path)}
		w.Header().Set("Content-Type","application/json")
		fmt.Fprint(w,`{"tokenPayloadExternal":{"requestDetails":{"requestPackageName":"com.example.bbw","requestHash":"hash"},"appIntegrity":{"appRecognitionVerdict":"PLAY_RECOGNIZED","packageName":"com.example.bbw","certificateSha256Digest":["cert"]},"deviceIntegrity":{"deviceRecognitionVerdict":["MEETS_DEVICE_INTEGRITY"],"recentDeviceActivity":{"deviceActivityLevel":"LEVEL_1"}}}}`)
	}));defer server.Close()
	decoder:=PlayIntegrityHTTPDecoder{Client:server.Client(),Tokens:StaticAccessToken("token"),BaseURL:server.URL}
	verdict,err:=decoder.DecodeIntegrityToken(context.Background(),"com.example.bbw","opaque");if err!=nil{t.Fatal(err)}
	if verdict.RequestHash!="hash"||verdict.PackageName!="com.example.bbw"||verdict.AppRecognitionVerdict!="PLAY_RECOGNIZED"||verdict.RecentDeviceActivity!="LEVEL_1"{t.Fatalf("verdict=%+v",verdict)}
}
