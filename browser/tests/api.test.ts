import { describe, expect, test } from "bun:test";
import { fetchRules, pairClient } from "../src/api";
import { normalizeApiOrigin } from "../src/config";

describe("API client", () => {
	test("pairs with the v1 contract", async () => {
		let request: Request | undefined;
		const fetcher = (async (input: RequestInfo | URL, init?: RequestInit) => {
			request = new Request(input, init);
			return Response.json(
				{
					accessToken: "secret",
					client: {
						id: 1,
						name: "test",
						createdAt: "now",
						lastActiveAt: "now",
					},
					device: { id: 2, name: "Laptop" },
				},
				{ status: 201 },
			);
		}) as typeof fetch;

		await pairClient(
			fetcher,
			"https://blkhole.sh",
			"pair-token",
			"test",
			"firefox",
		);
		expect(request?.url).toBe("https://blkhole.sh/api/browser/v1/pair");
		expect(await request?.json()).toEqual({
			pairingToken: "pair-token",
			clientName: "test",
			browser: "firefox",
		});
	});

	test("sends the token and ETag when fetching rules", async () => {
		let request: Request | undefined;
		const fetcher = (async (input: RequestInfo | URL, init?: RequestInit) => {
			request = new Request(input, init);
			return new Response(null, { status: 304 });
		}) as typeof fetch;
		const result = await fetchRules(
			fetcher,
			"https://blkhole.sh",
			"secret",
			'"old"',
		);
		expect(result.status).toBe("not-modified");
		expect(request?.headers.get("authorization")).toBe("Bearer secret");
		expect(request?.headers.get("if-none-match")).toBe('"old"');
	});

	test("treats revoked credentials as unpaired", async () => {
		const fetcher = (async () =>
			new Response(null, { status: 401 })) as typeof fetch;
		const result = await fetchRules(fetcher, "https://blkhole.sh", "revoked");
		expect(result).toEqual({ status: "unauthorized" });
	});

	test("requires a bare HTTPS origin", () => {
		expect(() => normalizeApiOrigin("http://blkhole.sh")).toThrow();
		expect(() => normalizeApiOrigin("https://blkhole.sh/api")).toThrow();
		expect(normalizeApiOrigin("https://blkhole.sh")).toBe("https://blkhole.sh");
	});
});
