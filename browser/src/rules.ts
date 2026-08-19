import type { DynamicRule } from "./types";

export const DOMAINS_PER_RULE = 5_000;

export function normalizeDomains(domains: string[]): string[] {
	const normalized = new Set<string>();
	for (const value of domains) {
		const domain = value.trim().toLowerCase().replace(/\.$/, "");
		if (!domain || domain.length > 253) {
			throw new Error(`Invalid domain in rules response: ${value}`);
		}
		const labels = domain.split(".");
		if (
			labels.some(
				(label) =>
					!label ||
					label.length > 63 ||
					!/^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/.test(label),
			)
		) {
			throw new Error(`Invalid domain in rules response: ${value}`);
		}
		normalized.add(domain);
	}
	return [...normalized].sort();
}

export function createDynamicRules(domains: string[]): DynamicRule[] {
	const normalized = normalizeDomains(domains);
	const rules: DynamicRule[] = [];
	for (let offset = 0; offset < normalized.length; offset += DOMAINS_PER_RULE) {
		rules.push({
			id: rules.length + 1,
			priority: 1,
			action: {
				type: "redirect",
				redirect: { extensionPath: "/blocked.html" },
			},
			condition: {
				requestDomains: normalized.slice(offset, offset + DOMAINS_PER_RULE),
				resourceTypes: ["main_frame"],
			},
		});
	}
	return rules;
}

function getDynamicRules(): Promise<Array<{ id: number }>> {
	return new Promise((resolve, reject) => {
		chrome.declarativeNetRequest.getDynamicRules(
			(rules: Array<{ id: number }>) => {
				const error = chrome.runtime.lastError;
				if (error) reject(new Error(error.message));
				else resolve(rules);
			},
		);
	});
}

export async function replaceDynamicRules(domains: string[]): Promise<number> {
	const existing = await getDynamicRules();
	const addRules = createDynamicRules(domains);
	await new Promise<void>((resolve, reject) => {
		chrome.declarativeNetRequest.updateDynamicRules(
			{ removeRuleIds: existing.map((rule) => rule.id), addRules },
			() => {
				const error = chrome.runtime.lastError;
				if (error) reject(new Error(error.message));
				else resolve();
			},
		);
	});
	return normalizeDomains(domains).length;
}
