import { normalizeApiOrigin } from "./config";
import type { PairResponse } from "./types";

export interface RulesResponse {
	status: "changed" | "not-modified" | "unauthorized";
	domains?: string[];
	etag?: string;
}

async function errorMessage(response: Response): Promise<string> {
	const message = (await response.text()).trim();
	return message || `Request failed with status ${response.status}`;
}

export async function pairClient(
	fetcher: typeof fetch,
	apiOrigin: string,
	pairingToken: string,
	clientName: string,
	browser: string,
): Promise<PairResponse> {
	const response = await fetcher(
		`${normalizeApiOrigin(apiOrigin)}/api/browser/v1/pair`,
		{
			method: "POST",
			headers: {
				"Content-Type": "application/json",
				Accept: "application/json",
			},
			body: JSON.stringify({ pairingToken, clientName, browser }),
		},
	);
	if (response.status !== 201) throw new Error(await errorMessage(response));
	const result = (await response.json()) as PairResponse;
	if (!result.accessToken || !result.client || !result.device) {
		throw new Error("Pairing response is incomplete");
	}
	return result;
}

export async function fetchRules(
	fetcher: typeof fetch,
	apiOrigin: string,
	accessToken: string,
	etag?: string,
): Promise<RulesResponse> {
	const headers: Record<string, string> = {
		Accept: "application/json",
		Authorization: `Bearer ${accessToken}`,
	};
	if (etag) headers["If-None-Match"] = etag;

	const response = await fetcher(
		`${normalizeApiOrigin(apiOrigin)}/api/browser/v1/rules`,
		{
			headers,
		},
	);
	if (response.status === 304) return { status: "not-modified", etag };
	if (response.status === 401 || response.status === 403) {
		return { status: "unauthorized" };
	}
	if (!response.ok) throw new Error(await errorMessage(response));

	const result = (await response.json()) as { domains?: unknown };
	if (
		!Array.isArray(result.domains) ||
		!result.domains.every((domain) => typeof domain === "string")
	) {
		throw new Error("Rules response is invalid");
	}
	return {
		status: "changed",
		domains: result.domains,
		etag: response.headers.get("ETag") ?? undefined,
	};
}
