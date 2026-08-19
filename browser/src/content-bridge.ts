import { validateBridgeRequest } from "./bridge";

function sendRuntimeMessage(message: unknown): Promise<unknown> {
	return new Promise((resolve, reject) => {
		chrome.runtime.sendMessage(message, (response: unknown) => {
			const error = chrome.runtime.lastError;
			if (error) reject(new Error(error.message));
			else resolve(response);
		});
	});
}

window.addEventListener("message", async (event: MessageEvent) => {
	if (event.source !== window || event.origin !== window.location.origin)
		return;
	if (window.location.protocol !== "https:") return;

	try {
		const status = (await sendRuntimeMessage({
			type: "bridge:get-config",
		})) as {
			apiOrigin?: string;
		};
		if (!status?.apiOrigin) return;
		const request = validateBridgeRequest(
			event.data,
			event.origin,
			status.apiOrigin,
		);
		if (!request) return;

		const response = await sendRuntimeMessage({
			type:
				request.type === "BLKHOLE_EXTENSION_PING"
					? "bridge:ping"
					: "bridge:pair",
			origin: event.origin,
			request,
		});
		window.postMessage(response, event.origin);
	} catch {
		// A rejected origin should be indistinguishable from an absent extension.
	}
});
