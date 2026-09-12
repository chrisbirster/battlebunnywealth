package com.battlebunnywealth.attestation

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import com.google.android.play.core.integrity.IntegrityManagerFactory
import com.google.android.play.core.integrity.StandardIntegrityManager
import java.security.KeyPairGenerator
import java.security.KeyStore
import java.security.Signature
import java.security.spec.ECGenParameterSpec

/** v0.7 integration spike. The host app supplies authenticated HTTP transport. */
class AttestedDeviceClient(
    context: Context,
    private val cloudProjectNumber: Long,
) {
    private val integrityManager = IntegrityManagerFactory.createStandard(context)
    private var tokenProvider: StandardIntegrityManager.StandardIntegrityTokenProvider? = null

    data class Challenge(
        val id: String,
        val deviceId: String,
        val provider: String,
        val nonce: String,
        val devicePublicKeyHash: String,
        val bindingPayload: String,
        val requestHash: String,
    )

    data class Evidence(
        val provider: String,
        val challengeId: String,
        val deviceId: String,
        val payload: String,
    )

    fun prepare(onReady: () -> Unit, onError: (Throwable) -> Unit) {
        integrityManager.prepareIntegrityToken(
            StandardIntegrityManager.PrepareIntegrityTokenRequest.builder()
                .setCloudProjectNumber(cloudProjectNumber)
                .build()
        ).addOnSuccessListener { provider ->
            tokenProvider = provider
            onReady()
        }.addOnFailureListener(onError)
    }

    fun createHardwareBackedMissionKey(alias: String = "bbw-pop-device"): String {
        val generator = KeyPairGenerator.getInstance(KeyProperties.KEY_ALGORITHM_EC, "AndroidKeyStore")
        val builder = KeyGenParameterSpec.Builder(alias, KeyProperties.PURPOSE_SIGN or KeyProperties.PURPOSE_VERIFY)
            .setAlgorithmParameterSpec(ECGenParameterSpec("secp256r1"))
            .setDigests(KeyProperties.DIGEST_SHA256)
            .setUserAuthenticationRequired(false)
        try { builder.setIsStrongBoxBacked(true) } catch (_: Throwable) { /* StrongBox is optional. */ }
        generator.initialize(builder.build())
        val pair = generator.generateKeyPair()
        // PublicKey.encoded is X.509 SubjectPublicKeyInfo, matching the Go enrollment API.
        return base64Url(pair.public.encoded)
    }

    fun requestIntegrityToken(challenge: Challenge, onResult: (Evidence) -> Unit, onError: (Throwable) -> Unit) {
        val provider = tokenProvider ?: return onError(IllegalStateException("Play Integrity provider not prepared"))
        provider.request(
            StandardIntegrityManager.StandardIntegrityTokenRequest.builder()
                .setRequestHash(challenge.requestHash)
                .build()
        ).addOnSuccessListener { response ->
            onResult(Evidence("google-play-integrity", challenge.id, challenge.deviceId, response.token()))
        }.addOnFailureListener(onError)
    }

    fun signMissionPayload(payload: String, alias: String = "bbw-pop-device"): String {
        val store = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }
        val privateKey = store.getKey(alias, null)
        val signer = Signature.getInstance("SHA256withECDSA")
        signer.initSign(privateKey as java.security.PrivateKey)
        signer.update(payload.toByteArray(Charsets.UTF_8))
        return base64Url(signer.sign())
    }

    private fun base64Url(bytes: ByteArray): String = Base64.encodeToString(bytes, Base64.URL_SAFE or Base64.NO_WRAP or Base64.NO_PADDING)
}
