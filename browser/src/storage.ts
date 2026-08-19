import { DEFAULT_API_ORIGIN, normalizeApiOrigin } from "./config";
import type { ExtensionState, PairResponse } from "./types";

const STATE_KEY = "extensionState";

function localGet(key: string): Promise<Record<string, unknown>> {
	return new Promise((resolve, reject) => {
		chrome.storage.local.get(key, (result: Record<string, unknown>) => {
			const error = chrome.runtime.lastError;
			if (error) reject(new Error(error.message));
			else resolve(result);
		});
	});
}

function localSet(value: Record<string, unknown>): Promise<void> {
	return new Promise((resolve, reject) => {
		chrome.storage.local.set(value, () => {
			const error = chrome.runtime.lastError;
			if (error) reject(new Error(error.message));
			else resolve();
		});
	});
}

export async function getState(): Promise<ExtensionState> {
	const result = await localGet(STATE_KEY);
	const stored = result[STATE_KEY] as Partial<ExtensionState> | undefined;
	const apiOrigin = stored?.apiOrigin ?? DEFAULT_API_ORIGIN;
	return {
		...stored,
		apiOrigin: apiOrigin ? normalizeApiOrigin(apiOrigin) : "",
	};
}

export async function setState(state: ExtensionState): Promise<void> {
	await localSet({ [STATE_KEY]: state });
}

export async function savePairing(response: PairResponse): Promise<void> {
	const state = await getState();
	await setState({
		apiOrigin: state.apiOrigin,
		accessToken: response.accessToken,
		client: response.client,
		device: response.device,
	});
}

export async function clearPairing(): Promise<void> {
	const state = await getState();
	await setState({ apiOrigin: state.apiOrigin });
}

export async function changeApiOrigin(apiOrigin: string): Promise<void> {
	await setState({ apiOrigin: normalizeApiOrigin(apiOrigin) });
}
