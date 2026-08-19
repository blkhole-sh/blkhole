import { cp, mkdir, readFile, rm } from "node:fs/promises";
import { resolve } from "node:path";

const root = resolve(import.meta.dir, "..");
const requested = process.argv[2] ?? "all";
const platforms =
	requested === "all" ? ["chromium", "firefox", "safari"] : [requested];
const allowed = new Set(["chromium", "firefox", "safari"]);
const apiOrigin = process.env.BLKHOLE_DEFAULT_API_ORIGIN ?? "";

if (platforms.some((platform) => !allowed.has(platform))) {
	throw new Error(`Unknown platform: ${requested}`);
}
if (apiOrigin) {
	const parsedOrigin = new URL(apiOrigin);
	if (parsedOrigin.protocol !== "https:" || parsedOrigin.origin !== apiOrigin) {
		throw new Error(
			"BLKHOLE_DEFAULT_API_ORIGIN must be an HTTPS origin without a trailing slash",
		);
	}
}

for (const platform of platforms) {
	const output = resolve(root, "build", platform);
	await rm(output, { recursive: true, force: true });
	await mkdir(output, { recursive: true });
	for (const entry of ["background", "content-bridge", "blocked", "options"]) {
		const result = await Bun.build({
			entrypoints: [resolve(root, "src", `${entry}.ts`)],
			outdir: output,
			format: "iife",
			target: "browser",
			naming: `${entry}.js`,
			minify: false,
			define: {
				__BLKHOLE_DEFAULT_API_ORIGIN__: JSON.stringify(apiOrigin),
				__BLKHOLE_PLATFORM__: JSON.stringify(platform),
			},
		});
		if (!result.success)
			throw new AggregateError(result.logs, `Failed to build ${entry}`);
	}
	await cp(resolve(root, "static"), output, { recursive: true });
	const manifest = JSON.parse(
		await readFile(
			resolve(root, "platforms", platform, "manifest.json"),
			"utf8",
		),
	);
	await Bun.write(
		resolve(output, "manifest.json"),
		`${JSON.stringify(manifest, null, 2)}\n`,
	);
}
