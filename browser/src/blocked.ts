interface BlockedNavigation {
	hostname?: string;
}

function sendMessage(message: unknown): Promise<BlockedNavigation> {
	return new Promise((resolve) => chrome.runtime.sendMessage(message, resolve));
}

document.addEventListener("DOMContentLoaded", async () => {
	const result = await sendMessage({ type: "blocked:get-navigation" });
	const domain = document.querySelector<HTMLElement>("[data-domain]");
	if (domain && result?.hostname) domain.textContent = result.hostname;
	document
		.querySelector("[data-back]")
		?.addEventListener("click", () => history.back());
});

export {};
