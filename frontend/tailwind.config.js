/** @type {import('tailwindcss').Config} */
export default {
	content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
	theme: {
		extend: {
			fontFamily: {
				hedvig: ["Hedvig Letters Serif", "serif"],
				inter: ["Inter", "sans-serif"],
			},
		},
	},
	plugins: [],
};
