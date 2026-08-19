import { fetchRules } from "./api";
import { replaceDynamicRules } from "./rules";
import { clearPairing, getState, setState } from "./storage";

let currentSync: Promise<{ paired: boolean; changed: boolean }> | undefined;

async function runSync(): Promise<{ paired: boolean; changed: boolean }> {
	const state = await getState();
	if (!state.accessToken) {
		await replaceDynamicRules([]);
		return { paired: false, changed: false };
	}

	const response = await fetchRules(
		fetch,
		state.apiOrigin,
		state.accessToken,
		state.etag,
	);
	if (response.status === "unauthorized") {
		await replaceDynamicRules([]);
		await clearPairing();
		return { paired: false, changed: true };
	}
	if (response.status === "not-modified") {
		await setState({ ...state, lastSyncedAt: new Date().toISOString() });
		return { paired: true, changed: false };
	}

	const domainCount = await replaceDynamicRules(response.domains ?? []);
	await setState({
		...state,
		etag: response.etag,
		domainCount,
		lastSyncedAt: new Date().toISOString(),
	});
	return { paired: true, changed: true };
}

export function syncRules(): Promise<{ paired: boolean; changed: boolean }> {
	if (!currentSync) {
		currentSync = runSync().finally(() => {
			currentSync = undefined;
		});
	}
	return currentSync;
}
