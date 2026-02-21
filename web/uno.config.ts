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
		"font-sans": "font-[Noto_Sans,sans-serif]",
		"font-display": "font-[Hedvig_Letters_Serif,serif]",
	},
});
