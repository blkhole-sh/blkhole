import { describe, expect, test } from "bun:test";
import { validateBridgeRequest } from "../src/bridge";

const ping = {
	source: "blkhole-webapp",
	type: "BLKHOLE_EXTENSION_PING",
	nonce: "123",
};

describe("webapp bridge", () => {
	test("only accepts the configured origin", () => {
		expect(
			validateBridgeRequest(ping, "https://blkhole.sh", "https://blkhole.sh"),
		).toEqual(ping);
		expect(
			validateBridgeRequest(ping, "https://evil.example", "https://blkhole.sh"),
		).toBeUndefined();
	});

	test("requires pairing API origin to equal the page origin", () => {
		const pair = {
			...ping,
			type: "BLKHOLE_EXTENSION_PAIR",
			pairingToken: "pair-token",
			apiOrigin: "https://other.example",
		};
		expect(
			validateBridgeRequest(pair, "https://blkhole.sh", "https://blkhole.sh"),
		).toBeUndefined();
	});
});
