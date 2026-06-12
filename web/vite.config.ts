import { defineConfig } from "vite";
import solid from "vite-plugin-solid";
import UnoCSS from "unocss/vite";
import { resolve } from "path";
import compression from "vite-plugin-compression";
import { readFileSync } from "fs";

const { version } = JSON.parse(
	readFileSync(resolve(__dirname, "./package.json"), "utf-8"),
);

export default defineConfig({
	server: {
		proxy: {
			"/api": "http://localhost:8080",
		},
	},
	define: {
		__APP_VERSION__: JSON.stringify(version ?? null),
	},
	plugins: [
		solid(),
		UnoCSS(),
		compression({
			algorithm: "brotliCompress",
			ext: ".br",
			threshold: 1024,
			deleteOriginFile: false,
		}),
		compression({
			algorithm: "gzip",
			ext: ".gz",
			threshold: 1024,
			deleteOriginFile: false,
		}),
	],
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
