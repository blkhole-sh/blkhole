import { defineConfig } from "vite";
import solid from "vite-plugin-solid";
import UnoCSS from "unocss/vite";
import { resolve } from "path";

export default defineConfig({
	plugins: [solid(), UnoCSS()],
	resolve: {
		alias: {
			"~": resolve(__dirname, "./src"),
		},
	},
	build: {
		outDir: "../static",
		emptyOutDir: true,
		minify: "terser",
		cssMinify: true,
		reportCompressedSize: true,
		terserOptions: {
			compress: {
				drop_console: true,
				drop_debugger: true,
				pure_funcs: ["console.log", "console.info", "console.debug"],
				passes: 3,
				unsafe: true,
				toplevel: true,
			},
		},
		rollupOptions: {
			output: {
				manualChunks: {
					vendor: ["solid-js", "@solidjs/router"],
				},
			},
		},
	},
});
