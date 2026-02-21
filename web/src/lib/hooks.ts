import { useLocation } from "@solidjs/router";
import { onMount } from "solid-js";

export function useScrollToHash() {
	const location = useLocation();
	onMount(() => {
		if (location.hash) {
			setTimeout(() => {
				const element = document.querySelector(location.hash);
				if (element) {
					element.scrollIntoView({ behavior: "smooth", block: "center" });
				}
			}, 100);
		}
	});
}
