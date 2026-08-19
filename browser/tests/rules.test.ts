import { describe, expect, test } from "bun:test";
import {
	createDynamicRules,
	DOMAINS_PER_RULE,
	normalizeDomains,
} from "../src/rules";

describe("dynamic rules", () => {
	test("normalizes and deduplicates domains", () => {
		expect(
			normalizeDomains(["Example.COM.", "example.com", "a.example.com"]),
		).toEqual(["a.example.com", "example.com"]);
	});

	test("redirects top-level navigations only", () => {
		const [rule] = createDynamicRules(["example.com"]);
		expect(rule?.condition.resourceTypes).toEqual(["main_frame"]);
		expect(rule?.condition.requestDomains).toEqual(["example.com"]);
		expect(rule?.action.redirect.extensionPath).toBe("/blocked.html");
	});

	test("chunks without truncating large lists", () => {
		const domains = Array.from(
			{ length: DOMAINS_PER_RULE + 2 },
			(_, index) => `d${index}.example`,
		);
		const rules = createDynamicRules(domains);
		expect(rules).toHaveLength(2);
		expect(rules.flatMap((rule) => rule.condition.requestDomains)).toHaveLength(
			domains.length,
		);
	});
});
