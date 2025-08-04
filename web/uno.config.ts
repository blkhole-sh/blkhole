import { defineConfig, presetWind4 } from "unocss";

export default defineConfig({
	presets: [
		presetWind4({
			preflights: {
				reset: true,
			},
		}),
	],
	shortcuts: {
		"font-hedvig": "font-[Hedvig_Letters_Serif,serif]",
		"font-inter": "font-[Inter,sans-serif]",
	},
});
