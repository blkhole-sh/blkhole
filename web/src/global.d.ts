/// <reference types="vite/client" />

declare const __APP_VERSION__: string | null;

interface ImportMetaEnv {
	readonly VITE_BLKHOLE_EXTENSION_CHROMIUM_URL?: string;
	readonly VITE_BLKHOLE_EXTENSION_FIREFOX_URL?: string;
	readonly VITE_BLKHOLE_EXTENSION_SAFARI_URL?: string;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}

declare module "*.css" {
	const content: string;
	export default content;
}

declare module "*.svg" {
	const content: string;
	export default content;
}
