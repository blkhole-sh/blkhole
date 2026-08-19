export const DEFAULT_API_ORIGIN =
	typeof __BLKHOLE_DEFAULT_API_ORIGIN__ === "undefined"
		? ""
		: __BLKHOLE_DEFAULT_API_ORIGIN__;
export const PLATFORM =
	typeof __BLKHOLE_PLATFORM__ === "undefined"
		? "chromium"
		: __BLKHOLE_PLATFORM__;

export function normalizeApiOrigin(value: string): string {
	const url = new URL(value);
	if (url.protocol !== "https:") {
		throw new Error("The API origin must use HTTPS");
	}
	if (
		url.username ||
		url.password ||
		url.pathname !== "/" ||
		url.search ||
		url.hash
	) {
		throw new Error("Enter an origin without a path, query, or credentials");
	}
	return url.origin;
}
