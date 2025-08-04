import { defineConfig } from "vite";
import solid from "vite-plugin-solid";
import { resolve } from "path";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
	plugins: [solid(), tailwindcss()],
	resolve: {
		alias: {
			"~": resolve(__dirname, "./src"),
		},
	},
	build: {
		outDir: "../assets",
		emptyOutDir: true,
		minify: "terser",
		cssMinify: true,
		reportCompressedSize: true,
		terserOptions: {
			compress: {
				drop_console: true,
				drop_debugger: true,
				pure_funcs: ['console.log', 'console.info', 'console.debug'],
				passes: 2,
			},
		},
		rollupOptions: {
			output: {
				manualChunks: {
					vendor: ['solid-js', '@solidjs/router'],
				},
			},
		},
	},
});

