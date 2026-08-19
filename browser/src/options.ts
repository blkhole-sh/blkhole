interface OptionsResponse {
	error?: string;
	apiOrigin: string;
	paired?: boolean;
	device?: { name: string };
	domainCount?: number;
}

function sendMessage(message: unknown): Promise<OptionsResponse> {
	return new Promise((resolve) => chrome.runtime.sendMessage(message, resolve));
}

async function render(): Promise<void> {
	const state = await sendMessage({ type: "options:status" });
	const origin = document.querySelector<HTMLInputElement>("[data-origin]");
	const status = document.querySelector<HTMLElement>("[data-status]");
	if (origin) origin.value = state.apiOrigin;
	if (status) {
		status.textContent = state.paired
			? `Paired with ${state.device?.name ?? "device"}. ${state.domainCount} domains active.`
			: "Not paired";
	}
}

document.addEventListener("DOMContentLoaded", () => {
	void render();
	document.querySelector("[data-save]")?.addEventListener("click", async () => {
		const input = document.querySelector<HTMLInputElement>("[data-origin]");
		const error = document.querySelector<HTMLElement>("[data-error]");
		const result = await sendMessage({
			type: "options:set-origin",
			apiOrigin: input?.value,
		});
		if (error) error.textContent = result.error ?? "";
		if (!result.error) await render();
	});
	document.querySelector("[data-sync]")?.addEventListener("click", async () => {
		const result = await sendMessage({ type: "options:sync" });
		const error = document.querySelector<HTMLElement>("[data-error]");
		if (error) error.textContent = result.error ?? "";
		await render();
	});
	document
		.querySelector("[data-unpair]")
		?.addEventListener("click", async () => {
			await sendMessage({ type: "options:unpair" });
			await render();
		});
});

export {};
