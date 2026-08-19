import { normalizeApiOrigin } from "./config";
import type { BridgeRequest } from "./types";

function validNonce(value: unknown): value is string {
	return typeof value === "string" && value.length > 0 && value.length <= 200;
}

export function validateBridgeRequest(
	value: unknown,
	pageOrigin: string,
	configuredOrigin: string,
): BridgeRequest | undefined {
	if (typeof value !== "object" || value === null) return;
	const message = value as Record<string, unknown>;
	if (
		message.source !== "blkhole-webapp" ||
		!validNonce(message.nonce) ||
		normalizeApiOrigin(pageOrigin) !== normalizeApiOrigin(configuredOrigin)
	) {
		return;
	}
	if (message.type === "BLKHOLE_EXTENSION_PING")
		return message as unknown as BridgeRequest;
	if (message.type !== "BLKHOLE_EXTENSION_PAIR") return;
	if (
		typeof message.pairingToken !== "string" ||
		message.pairingToken.length < 1 ||
		message.pairingToken.length > 2_048 ||
		typeof message.apiOrigin !== "string" ||
		normalizeApiOrigin(message.apiOrigin) !== normalizeApiOrigin(pageOrigin)
	) {
		return;
	}
	return message as unknown as BridgeRequest;
}
