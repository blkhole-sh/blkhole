import { defineConfig, presetWind4 } from "unocss";

export default defineConfig({
	content: {
		filesystem: ["src/**/*.{tsx,ts,jsx,js,html}"],
	},
	presets: [
		presetWind4({
			preflights: {
				reset: true,
			},
		}),
	],
	preflights: [
		{
			getCSS: () => `
				body {
					font-family: "Noto Sans", sans-serif;
				}
			`,
		},
	],
	shortcuts: {
		"font-sans": "font-[Noto_Sans,sans-serif]",
		"font-display": "font-[Hedvig_Letters_Serif,serif]",
		"font-mono": "font-[Commit_Mono,monospace]",
	},
});
