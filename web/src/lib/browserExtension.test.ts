import { afterEach, beforeEach, describe, expect, test } from "bun:test";
import {
	detectBrowserExtension,
	EXTENSION_MESSAGE_SOURCE,
	getExtensionInstallTarget,
	pairBrowserExtension,
} from "./browserExtension";

type MessageListener = (event: MessageEvent) => void;

class FakeWindow {
	readonly location = { origin: "https://app.blkhole.test" };
	readonly messages: Array<{ message: unknown; targetOrigin: string }> = [];
	readonly listeners = new Set<MessageListener>();
	readonly timers = new Map<number, () => void>();
	private nextTimerId = 1;

	addEventListener(type: string, listener: MessageListener) {
		if (type === "message") this.listeners.add(listener);
	}

	removeEventListener(type: string, listener: MessageListener) {
		if (type === "message") this.listeners.delete(listener);
	}

	postMessage(message: unknown, targetOrigin: string) {
		this.messages.push({ message, targetOrigin });
	}

	setTimeout(handler: TimerHandler) {
		const id = this.nextTimerId++;
		if (typeof handler === "function") this.timers.set(id, handler);
		return id;
	}

	clearTimeout(id: number) {
		this.timers.delete(id);
	}

	dispatch(
		data: unknown,
		origin = this.location.origin,
		source: unknown = this,
	) {
		const event = { data, origin, source } as MessageEvent;
		for (const listener of this.listeners) listener(event);
	}

	runTimers() {
		for (const handler of [...this.timers.values()]) handler();
	}
}

const originalWindow = Object.getOwnPropertyDescriptor(globalThis, "window");
let fakeWindow: FakeWindow;

beforeEach(() => {
	fakeWindow = new FakeWindow();
	Object.defineProperty(globalThis, "window", {
		configurable: true,
		value: fakeWindow,
	});
});

afterEach(() => {
	if (originalWindow) {
		Object.defineProperty(globalThis, "window", originalWindow);
	} else {
		Reflect.deleteProperty(globalThis, "window");
	}
});

describe("browser extension bridge", () => {
	test("detects an extension only from a matching window response", async () => {
		const detection = detectBrowserExtension();
		const request = fakeWindow.messages[0]?.message as {
			nonce: string;
		};
		const response = {
			source: EXTENSION_MESSAGE_SOURCE,
			type: "BLKHOLE_EXTENSION_PONG",
			nonce: request.nonce,
			version: "1.2.3",
			pairedDeviceId: "device-1",
		};

		fakeWindow.dispatch(response, fakeWindow.location.origin, {});
		fakeWindow.dispatch(response, "https://attacker.test");
		fakeWindow.dispatch({ ...response, source: "other-extension" });
		fakeWindow.dispatch({ ...response, type: "OTHER_RESPONSE" });
		fakeWindow.dispatch({ ...response, nonce: "wrong-nonce" });
		fakeWindow.dispatch(response);

		expect(await detection).toEqual({
			version: "1.2.3",
			pairedDeviceId: "device-1",
		});
		expect(fakeWindow.listeners.size).toBe(0);
		expect(fakeWindow.timers.size).toBe(0);
	});

	test("returns null when extension detection times out", async () => {
		const detection = detectBrowserExtension();

		fakeWindow.runTimers();

		expect(await detection).toBeNull();
		expect(fakeWindow.listeners.size).toBe(0);
	});

	test("passes pairing data through and returns the extension response", async () => {
		const pairing = pairBrowserExtension(
			"pairing-token",
			"https://api.blkhole.test",
		);
		const sent = fakeWindow.messages[0];
		const request = sent?.message as {
			nonce: string;
			pairingToken: string;
			apiOrigin: string;
			type: string;
		};

		expect(sent?.targetOrigin).toBe(fakeWindow.location.origin);
		expect(request).toMatchObject({
			pairingToken: "pairing-token",
			apiOrigin: "https://api.blkhole.test",
			type: "BLKHOLE_EXTENSION_PAIR",
		});

		fakeWindow.dispatch({
			source: EXTENSION_MESSAGE_SOURCE,
			type: "BLKHOLE_EXTENSION_PAIR_RESULT",
			nonce: request.nonce,
			success: true,
		});

		expect(await pairing).toEqual({ success: true, error: undefined });
	});

	test("returns a useful pairing error when the extension times out", async () => {
		const pairing = pairBrowserExtension("pairing-token", "https://api.test");

		fakeWindow.runTimers();

		expect(await pairing).toEqual({
			success: false,
			error: "The extension did not respond.",
		});
	});
});

describe("extension store routing", () => {
	test.each([
		["Mozilla/5.0 Firefox/141.0", "firefox", "Firefox Add-ons"],
		[
			"Mozilla/5.0 (Macintosh) AppleWebKit/605.1.15 Version/18.5 Safari/605.1.15",
			"safari",
			"App Store",
		],
		[
			"Mozilla/5.0 AppleWebKit/537.36 Chrome/138.0.0.0 Safari/537.36",
			"chromium",
			"Chrome Web Store",
		],
		[
			"Mozilla/5.0 (iPhone; CPU iPhone OS 18_5 like Mac OS X) AppleWebKit/605.1.15 CriOS/138.0 Mobile/15E148 Safari/604.1",
			"unsupported",
			"Safari on iOS",
		],
		[
			"Mozilla/5.0 (iPad; CPU OS 18_5 like Mac OS X) AppleWebKit/605.1.15 FxiOS/141.0 Mobile/15E148 Safari/605.1.15",
			"unsupported",
			"Safari on iOS",
		],
		["ExampleBrowser/1.0", "unsupported", "browser extension store"],
	] as const)("routes %s to the expected store", (userAgent, browser, name) => {
		expect(getExtensionInstallTarget(userAgent)).toMatchObject({
			browser,
			name,
		});
	});
});
