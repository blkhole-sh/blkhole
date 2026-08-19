import { pairClient } from "./api";
import { PLATFORM } from "./config";
import { replaceDynamicRules } from "./rules";
import {
	changeApiOrigin,
	clearPairing,
	getState,
	savePairing,
} from "./storage";
import { syncRules } from "./sync";
import type { BridgeRequest } from "./types";

const SYNC_ALARM = "sync-rules";
const pendingNavigation = new Map<
	number,
	{ hostname: string; timestamp: number }
>();
const PENDING_NAVIGATION_PREFIX = "pendingNavigation:";

interface MessageSender {
	url?: string;
	tab?: { id?: number; url?: string };
}

interface RuntimeMessage {
	type?: string;
	origin?: string;
	apiOrigin?: string;
	request?: BridgeRequest;
}

interface NavigationDetails {
	frameId: number;
	tabId: number;
	url: string;
}

type PendingNavigation = { hostname: string; timestamp: number };

function pendingNavigationKey(tabID: number): string {
	return `${PENDING_NAVIGATION_PREFIX}${tabID}`;
}

function persistPendingNavigation(
	tabID: number,
	navigation: PendingNavigation,
): void {
	if (!chrome.storage.session) return;
	chrome.storage.session.set({ [pendingNavigationKey(tabID)]: navigation });
}

async function readPendingNavigation(
	tabID: number,
): Promise<PendingNavigation | undefined> {
	const current = pendingNavigation.get(tabID);
	if (current || !chrome.storage.session) return current;
	return new Promise((resolve) => {
		const key = pendingNavigationKey(tabID);
		chrome.storage.session.get(key, (result: Record<string, unknown>) => {
			resolve(result[key] as PendingNavigation | undefined);
		});
	});
}

function clearPendingNavigation(tabID: number): void {
	pendingNavigation.delete(tabID);
	chrome.storage.session?.remove(pendingNavigationKey(tabID));
}

function senderOrigin(sender: MessageSender): string | undefined {
	try {
		const url = sender.url ?? sender.tab?.url;
		return url ? new URL(url).origin : undefined;
	} catch {
		return;
	}
}

function isExtensionPage(sender: MessageSender): boolean {
	return (
		typeof sender.url === "string" &&
		sender.url.startsWith(chrome.runtime.getURL(""))
	);
}

async function detectBrowser(): Promise<string> {
	const userAgent = navigator.userAgent;
	if (userAgent.includes("Firefox/")) return "firefox";
	if (userAgent.includes("Edg/")) return "edge";
	const brave = (
		navigator as Navigator & { brave?: { isBrave(): Promise<boolean> } }
	).brave;
	if (brave && (await brave.isBrave())) return "brave";
	if (userAgent.includes("Safari/") && !userAgent.includes("Chrome/"))
		return "safari";
	return PLATFORM === "chromium" ? "chromium" : PLATFORM;
}

async function status() {
	const state = await getState();
	return {
		installed: true,
		paired: Boolean(state.accessToken),
		pairedDeviceId:
			typeof state.device?.id === "number" ? String(state.device.id) : null,
		apiOrigin: state.apiOrigin,
		version: chrome.runtime.getManifest().version,
		device: state.device,
		domainCount: state.domainCount ?? 0,
		lastSyncedAt: state.lastSyncedAt,
	};
}

async function handleBridge(
	message: RuntimeMessage,
	sender: MessageSender,
): Promise<unknown> {
	const state = await getState();
	const origin = senderOrigin(sender);
	if (!origin || message.origin !== origin || origin !== state.apiOrigin) {
		throw new Error("Origin is not allowed");
	}
	const request = message.request;
	if (!request) throw new Error("Invalid bridge request");
	if (message.type === "bridge:ping") {
		return {
			source: "blkhole-extension",
			type: "BLKHOLE_EXTENSION_PONG",
			nonce: request.nonce,
			...(await status()),
		};
	}
	if (
		message.type !== "bridge:pair" ||
		request.type !== "BLKHOLE_EXTENSION_PAIR" ||
		request.apiOrigin !== origin
	) {
		throw new Error("Invalid pairing request");
	}

	try {
		const browser = await detectBrowser();
		await replaceDynamicRules([]);
		await clearPairing();
		const response = await pairClient(
			fetch,
			state.apiOrigin,
			request.pairingToken,
			`${browser} extension`,
			browser,
		);
		await savePairing(response);
		syncInBackground();
		return {
			source: "blkhole-extension",
			type: "BLKHOLE_EXTENSION_PAIR_RESULT",
			nonce: request.nonce,
			success: true,
			paired: true,
			device: response.device,
		};
	} catch (error) {
		return {
			source: "blkhole-extension",
			type: "BLKHOLE_EXTENSION_PAIR_RESULT",
			nonce: request.nonce,
			success: false,
			error: error instanceof Error ? error.message : "Pairing failed",
		};
	}
}

async function handleMessage(
	message: RuntimeMessage,
	sender: MessageSender,
): Promise<unknown> {
	if (message?.type === "bridge:get-config") {
		const state = await getState();
		return { apiOrigin: state.apiOrigin };
	}
	if (message?.type === "bridge:ping" || message?.type === "bridge:pair") {
		return handleBridge(message, sender);
	}
	if (!isExtensionPage(sender)) throw new Error("Extension page required");
	if (message?.type === "options:status") return status();
	if (message?.type === "options:sync") return syncRules();
	if (message?.type === "options:unpair") {
		await replaceDynamicRules([]);
		await clearPairing();
		return status();
	}
	if (message?.type === "options:set-origin") {
		if (typeof message.apiOrigin !== "string") {
			throw new Error("API origin is required");
		}
		const state = await getState();
		if (state.apiOrigin === message.apiOrigin) return status();
		await replaceDynamicRules([]);
		await changeApiOrigin(message.apiOrigin);
		return status();
	}
	if (message?.type === "blocked:get-navigation") {
		const tabID = sender.tab?.id;
		const navigation =
			typeof tabID === "number"
				? await readPendingNavigation(tabID)
				: undefined;
		return navigation && Date.now() - navigation.timestamp < 60_000
			? navigation
			: {};
	}
}

chrome.runtime.onMessage.addListener(
	(
		message: RuntimeMessage,
		sender: MessageSender,
		sendResponse: (response: unknown) => void,
	) => {
		handleMessage(message, sender)
			.then(sendResponse)
			.catch((error) =>
				sendResponse({
					error: error instanceof Error ? error.message : "Failed",
				}),
			);
		return true;
	},
);

function syncInBackground(): void {
	void syncRules().catch((error) =>
		console.warn("blkhole rule sync failed", error),
	);
}

chrome.runtime.onInstalled.addListener(() => {
	chrome.alarms.create(SYNC_ALARM, { periodInMinutes: 1 });
	syncInBackground();
});
chrome.runtime.onStartup.addListener(syncInBackground);
chrome.alarms.onAlarm.addListener((alarm: { name: string }) => {
	if (alarm.name === SYNC_ALARM) syncInBackground();
});
chrome.action?.onClicked.addListener(() => chrome.runtime.openOptionsPage());

chrome.webNavigation.onBeforeNavigate.addListener(
	(details: NavigationDetails) => {
		if (details.frameId !== 0) return;
		try {
			const url = new URL(details.url);
			if (url.protocol === "http:" || url.protocol === "https:") {
				const navigation = {
					hostname: url.hostname,
					timestamp: Date.now(),
				};
				pendingNavigation.set(details.tabId, navigation);
				persistPendingNavigation(details.tabId, navigation);
			}
		} catch {
			// Ignore malformed navigation events.
		}
	},
);
chrome.tabs.onRemoved.addListener((tabID: number) =>
	clearPendingNavigation(tabID),
);
