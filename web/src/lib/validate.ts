/** Returns an error if the value is empty or whitespace-only. */
export const required =
	(msg = "Required") =>
	(v: string): string | undefined =>
		!v.trim() ? msg : undefined;

/** Returns an error if the value is a non-empty string that isn't a valid email address. */
export const email =
	(msg = "Invalid email") =>
	(v: string): string | undefined =>
		v.trim() && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v.trim()) ? msg : undefined;

/** Returns an error if the value is shorter than `n` characters. */
export const minLength =
	(n: number, msg?: string) =>
	(v: string): string | undefined =>
		v.length < n ? (msg ?? `At least ${n} characters`) : undefined;

/** Returns an error if the value is a non-empty string that doesn't start with http:// or https://. */
export const url =
	(msg = "Invalid URL") =>
	(v: string): string | undefined =>
		v.trim() && !/^https?:\/\/.+/.test(v.trim()) ? msg : undefined;

/** Returns an error if the value isn't a valid DNS hostname. */
export const domain =
	(msg = "Invalid domain") =>
	(v: string): string | undefined => {
		const value = v.trim();
		if (!value || value.length > 253) return msg;
		const labels = value.split(".");
		if (labels.length < 2) return msg;
		return labels.every(
			(label) =>
				label.length > 0 &&
				label.length <= 63 &&
				/^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/i.test(label),
		)
			? undefined
			: msg;
	};

/** Runs validators left-to-right and returns the first error message, or undefined if all pass. */
export const compose =
	(...validators: ((v: string) => string | undefined)[]) =>
	(v: string): string | undefined =>
		validators.reduce<string | undefined>((err, fn) => err ?? fn(v), undefined);
