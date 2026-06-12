/// <reference types="vite/client" />

declare const __APP_VERSION__: string | null;

declare module "*.css" {
	const content: string;
	export default content;
}

declare module "*.svg" {
	const content: string;
	export default content;
}
