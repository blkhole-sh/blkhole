import { describe, expect, it } from "bun:test";
import { domain } from "./validate";

describe("domain", () => {
	it("accepts DNS hostnames", () => {
		expect(domain()("example.com")).toBeUndefined();
		expect(domain()("ads.eu.example.com")).toBeUndefined();
		expect(domain()("xn--bcher-kva.example")).toBeUndefined();
	});

	it("rejects URLs and malformed hostnames", () => {
		expect(domain()("https://example.com")).toBe("Invalid domain");
		expect(domain()("example")).toBe("Invalid domain");
		expect(domain()("-ads.example.com")).toBe("Invalid domain");
		expect(domain()("ads..example.com")).toBe("Invalid domain");
	});
});
