import { defineConfig, presetWind4 } from "unocss";

// presetWind4 emits its theme variables (the `:root, :host` rule) and its
// reset properties (the `*, ::before, ::after, ::backdrop` rule) in first-use
// order, iterating Sets that are filled while files are scanned. That order is
// nondeterministic, so consecutive builds shuffle these declarations and change
// the stylesheet's content hash, which in turn renames every JS chunk that
// references it. We wrap the affected preflights and sort the declarations
// inside their innermost rule so the output is stable. See issue #87.
function sortInnermostDeclarations(css: string): string {
	// Match the deepest `{ ... }` (no nested braces) and sort its `;`-separated
	// declarations alphabetically.
	return css.replace(/\{([^{}]*)\}/g, (match, body: string) => {
		const decls = body
			.split(";")
			.map((d) => d.trim())
			.filter(Boolean);
		if (decls.length < 2) return match;
		return `{${decls.sort().join(";")}}`;
	});
}

function deterministicThemeOrder(preset: ReturnType<typeof presetWind4>) {
	for (const layer of ["theme", "properties"]) {
		const preflight = preset.preflights?.find((p) => p.layer === layer);
		if (!preflight) continue;
		const original = preflight.getCSS.bind(preflight);
		preflight.getCSS = async (ctx) => {
			const css = await original(ctx);
			return css ? sortInnermostDeclarations(css) : css;
		};
	}
	return preset;
}

export default defineConfig({
	content: {
		filesystem: ["src/**/*.{tsx,ts,jsx,js,html}"],
	},
	presets: [
		deterministicThemeOrder(
			presetWind4({
				preflights: {
					reset: true,
				},
			}),
		),
	],
	preflights: [
		{
			getCSS: () => `
				body {
					font-family: "Geist", sans-serif;
				}
				@keyframes pulse-strong {
					0%, 100% {
						opacity: 1;
					}
					50% {
						opacity: 0.5;
					}
				}
				.animate-pulse-strong {
					animation: pulse-strong 1s ease-in-out 2;
				}
			`,
		},
	],
	shortcuts: {
		"font-sans": "font-[Geist,sans-serif]",
		"font-display": "font-[Hedvig_Letters_Serif,serif]",
	},
});
