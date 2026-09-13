import CryptoKit
import DeviceCheck
import Foundation

/// v0.9 integration spike. The host app supplies authenticated HTTP transport.
final class BattleBunnyAppAttestClient {
    private let appAttest = DCAppAttestService.shared
    private let appAttestKeyIDDefaultsKey = "BattleBunnyAppAttestKeyID"
    private var missionKey: SecureEnclave.P256.Signing.PrivateKey?

    struct Challenge: Decodable {
        let id: String
        let deviceId: String
        let provider: String
        let purpose: String?
        let assertionPurpose: String?
        let nonce: String
        let devicePublicKeyHash: String?
        let bindingPayload: String
        let requestHash: String
    }

    struct Evidence: Encodable {
        let provider: String
        let challengeId: String
        let deviceId: String
        let payload: String
        let keyId: String
    }

    func createMissionSigningKey() throws -> (privateKey: SecureEnclave.P256.Signing.PrivateKey, publicKeySPKIBase64URL: String) {
        let key = try SecureEnclave.P256.Signing.PrivateKey()
        missionKey = key
        return (key, base64URL(p256SPKI(x963: key.publicKey.x963Representation)))
    }

    func attest(challenge: Challenge) async throws -> Evidence {
        guard appAttest.isSupported else { throw AttestError.unsupported }
        let keyID = try await appAttestKeyID()
        let clientDataHash = Data(SHA256.hash(data: Data(challenge.bindingPayload.utf8)))
        let object = try await appAttest.attestKey(keyID, clientDataHash: clientDataHash)
        return Evidence(provider: "apple-app-attest", challengeId: challenge.id, deviceId: challenge.deviceId, payload: object.base64EncodedString(), keyId: keyID)
    }

    func assertAppIntegrity(challenge: Challenge) async throws -> Evidence {
        guard appAttest.isSupported else { throw AttestError.unsupported }
        guard challenge.purpose == "assertion" else { throw AttestError.invalidChallenge }
        let keyID = try await existingAppAttestKeyID()
        let clientDataHash = Data(SHA256.hash(data: Data(challenge.bindingPayload.utf8)))
        let assertion = try await appAttest.generateAssertion(keyID, clientDataHash: clientDataHash)
        return Evidence(provider: "apple-app-attest", challengeId: challenge.id, deviceId: challenge.deviceId, payload: assertion.base64EncodedString(), keyId: keyID)
    }

    func signMissionPayload(_ payload: String) throws -> String {
        guard let key = missionKey else { throw AttestError.noMissionKey }
        let signature = try key.signature(for: Data(payload.utf8))
        return base64URL(signature.derRepresentation)
    }

    private func appAttestKeyID() async throws -> String {
        if let keyID = UserDefaults.standard.string(forKey: appAttestKeyIDDefaultsKey), !keyID.isEmpty { return keyID }
        let keyID = try await appAttest.generateKey()
        UserDefaults.standard.set(keyID, forKey: appAttestKeyIDDefaultsKey)
        return keyID
    }

    private func existingAppAttestKeyID() async throws -> String {
        guard let keyID = UserDefaults.standard.string(forKey: appAttestKeyIDDefaultsKey), !keyID.isEmpty else { throw AttestError.noAppAttestKey }
        return keyID
    }

    private func p256SPKI(x963: Data) -> Data {
        let prefix: [UInt8] = [0x30,0x59,0x30,0x13,0x06,0x07,0x2a,0x86,0x48,0xce,0x3d,0x02,0x01,0x06,0x08,0x2a,0x86,0x48,0xce,0x3d,0x03,0x01,0x07,0x03,0x42,0x00]
        return Data(prefix) + x963
    }

    private func base64URL(_ data: Data) -> String {
        data.base64EncodedString().replacingOccurrences(of: "+", with: "-").replacingOccurrences(of: "/", with: "_").replacingOccurrences(of: "=", with: "")
    }

    enum AttestError: Error { case unsupported, noMissionKey, noAppAttestKey, invalidChallenge }
}
