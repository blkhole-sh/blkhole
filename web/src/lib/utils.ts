export const cx = (...classes: (string | undefined | null | false)[]) => {
	return classes.filter(Boolean).join(" ");
};

export const configUrl = (deviceId: string) =>
	`/api/devices/${deviceId}/config`;

export const isIOS = (os: string) => os === "iOS";
export const isMacOS = (os: string) => os === "macOS";
export const isAndroid = (os: string) => os === "Android";
export const isLinux = (os: string) => os === "Linux";
export const isWindows = (os: string) => os === "Windows";
export const isApple = (os: string) => isIOS(os) || isMacOS(os);
