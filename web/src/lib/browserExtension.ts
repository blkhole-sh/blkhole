export const WEBAPP_MESSAGE_SOURCE = "blkhole-webapp";
export const EXTENSION_MESSAGE_SOURCE = "blkhole-extension";

const DETECTION_TIMEOUT_MS = 1200;
const PAIRING_TIMEOUT_MS = 10000;

interface ExtensionStatus {
	version: string;
	pairedDeviceId: string | null;
}

interface PairingResult {
	success: boolean;
	error?: string;
}

interface BridgeMessage {
	source: string;
	type: string;
	nonce: string;
	[key: string]: unknown;
}

export interface ExtensionInstallTarget {
	name: string;
	url?: string;
	browser: "chromium" | "firefox" | "safari" | "unsupported";
}

function nonce() {
	return crypto.randomUUID();
}

function isBridgeMessage(value: unknown): value is BridgeMessage {
	if (!value || typeof value !== "object") return false;
	const message = value as Record<string, unknown>;
	return (
		typeof message.source === "string" &&
		typeof message.type === "string" &&
		typeof message.nonce === "string"
	);
}

function sendBridgeRequest(
	request: BridgeMessage,
	responseType: string,
	timeoutMs: number,
): Promise<BridgeMessage | null> {
	return new Promise((resolve) => {
		let settled = false;
		const finish = (message: BridgeMessage | null) => {
			if (settled) return;
			settled = true;
			window.removeEventListener("message", handleMessage);
			window.clearTimeout(timeout);
			resolve(message);
		};
		const handleMessage = (event: MessageEvent) => {
			if (event.source !== window || event.origin !== window.location.origin)
				return;
			if (!isBridgeMessage(event.data)) return;
			if (
				event.data.source !== EXTENSION_MESSAGE_SOURCE ||
				event.data.type !== responseType ||
				event.data.nonce !== request.nonce
			)
				return;
			finish(event.data);
		};
		const timeout = window.setTimeout(() => finish(null), timeoutMs);
		window.addEventListener("message", handleMessage);
		window.postMessage(request, window.location.origin);
	});
}

export async function detectBrowserExtension(): Promise<ExtensionStatus | null> {
	const response = await sendBridgeRequest(
		{
			source: WEBAPP_MESSAGE_SOURCE,
			type: "BLKHOLE_EXTENSION_PING",
			nonce: nonce(),
		},
		"BLKHOLE_EXTENSION_PONG",
		DETECTION_TIMEOUT_MS,
	);
	if (!response || typeof response.version !== "string") return null;
	if (
		response.pairedDeviceId !== undefined &&
		response.pairedDeviceId !== null &&
		typeof response.pairedDeviceId !== "string"
	)
		return null;
	return {
		version: response.version,
		pairedDeviceId:
			typeof response.pairedDeviceId === "string"
				? response.pairedDeviceId
				: null,
	};
}

export async function pairBrowserExtension(
	pairingToken: string,
	apiOrigin: string,
): Promise<PairingResult> {
	const response = await sendBridgeRequest(
		{
			source: WEBAPP_MESSAGE_SOURCE,
			type: "BLKHOLE_EXTENSION_PAIR",
			nonce: nonce(),
			pairingToken,
			apiOrigin,
		},
		"BLKHOLE_EXTENSION_PAIR_RESULT",
		PAIRING_TIMEOUT_MS,
	);
	if (!response || typeof response.success !== "boolean") {
		return { success: false, error: "The extension did not respond." };
	}
	return {
		success: response.success,
		error: typeof response.error === "string" ? response.error : undefined,
	};
}

export function getExtensionInstallTarget(
	userAgent = navigator.userAgent,
): ExtensionInstallTarget {
	if (
		/iPhone|iPad|iPod/i.test(userAgent) &&
		/FxiOS|CriOS|EdgiOS|OPiOS/i.test(userAgent)
	) {
		return { browser: "unsupported", name: "Safari on iOS" };
	}
	if (/Firefox|FxiOS/i.test(userAgent)) {
		return {
			browser: "firefox",
			name: "Firefox Add-ons",
			url: import.meta.env.VITE_BLKHOLE_EXTENSION_FIREFOX_URL,
		};
	}
	if (
		/Safari/i.test(userAgent) &&
		!/Chrome|Chromium|CriOS|Edg|OPR|FxiOS/i.test(userAgent)
	) {
		return {
			browser: "safari",
			name: "App Store",
			url: import.meta.env.VITE_BLKHOLE_EXTENSION_SAFARI_URL,
		};
	}
	if (/Chrome|Chromium|CriOS|Edg|OPR/i.test(userAgent)) {
		return {
			browser: "chromium",
			name: "Chrome Web Store",
			url: import.meta.env.VITE_BLKHOLE_EXTENSION_CHROMIUM_URL,
		};
	}
	return { browser: "unsupported", name: "browser extension store" };
}
